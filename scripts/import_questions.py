#!/usr/bin/env python3
"""
Bulk import questions from a JSON file into Exam Arena.
Supports:
  1. Direct PostgreSQL connection (recommended): FAST bulk insert using psycopg2,
     psycopg, or local `psql` CLI, with auto-detection of .env DATABASE_URL.
  2. HTTP REST API mode (--api): Authenticates with admin JWT and uploads via API.
"""

import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile
import urllib.error
import urllib.parse
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime, timezone
from urllib.parse import quote, unquote, urlsplit, urlunsplit


# =====================================================================
# Environment & Connection Helpers
# =====================================================================

def find_env_database_url():
    """Look for DATABASE_URL in environment or nearest .env file."""
    if os.getenv("DATABASE_URL"):
        return os.getenv("DATABASE_URL")

    candidates = [
        os.path.join(os.getcwd(), ".env"),
        os.path.join(os.getcwd(), "..", ".env"),
        os.path.join(os.path.dirname(__file__), "..", ".env"),
    ]
    for path in candidates:
        if os.path.exists(path):
            try:
                with open(path, "r", encoding="utf-8") as f:
                    for line in f:
                        line = line.strip()
                        if line.startswith("#"):
                            continue
                        if line.startswith("DATABASE_URL="):
                            val = line.split("=", 1)[1].strip().strip('"').strip("'")
                            if val:
                                return val
            except Exception:
                pass
    return None


def build_postgres_url(base_url, user=None, password=None, host=None, port=None, dbname=None):
    """Construct or override a PostgreSQL connection URL."""
    if not base_url:
        host = host or "localhost"
        port = port or 5432
        dbname = dbname or "examarena"
        auth = ""
        if user and password:
            auth = f"{quote(user)}:{quote(password)}@"
        elif user:
            auth = f"{quote(user)}@"
        return f"postgresql://{auth}{host}:{port}/{dbname}"

    parts = urlsplit(base_url)
    eff_user = quote(user) if user else (quote(unquote(parts.username)) if parts.username else None)
    eff_pass = quote(password) if password is not None else (quote(unquote(parts.password)) if parts.password else None)
    eff_host = host or parts.hostname or "localhost"
    eff_port = port or parts.port
    eff_dbname = dbname or parts.path.lstrip("/") or "examarena"

    auth = ""
    if eff_user and eff_pass:
        auth = f"{eff_user}:{eff_pass}@"
    elif eff_user:
        auth = f"{eff_user}@"

    netloc = f"{auth}{eff_host}"
    if eff_port:
        netloc += f":{eff_port}"

    query = parts.query
    return urlunsplit((parts.scheme or "postgresql", netloc, f"/{eff_dbname}", query, parts.fragment))


# =====================================================================
# Direct PostgreSQL Importer
# =====================================================================

def import_via_psycopg(driver, db_url, questions, publish=False):
    """Import questions directly into PostgreSQL using psycopg2 or psycopg3."""
    print(f"[*] Connecting to database via {driver.__name__}...")
    conn = driver.connect(db_url)
    conn.autocommit = False

    status = "published" if publish else "draft"
    now_utc = datetime.now(timezone.utc) if publish else None

    success_count = 0
    fail_count = 0
    failed_items = []

    print(f"[*] Inserting {len(questions)} questions (publish={publish})...")
    with conn.cursor() as cur:
        for idx, q in enumerate(questions, start=1):
            try:
                # Savepoint per question so one failure doesn't rollback everything
                cur.execute(f"SAVEPOINT q_{idx}")

                cat_id = q.get("exam_category_id")
                topic_id = q.get("topic_id")
                q_type = q.get("question_type", "mcq_single")
                diff = q.get("difficulty", "medium")
                lang = q.get("language", "en")
                body = q.get("body", "")
                explanation = q.get("explanation")
                est_sec = q.get("estimated_time_seconds", 60)
                options = q.get("options", [])

                cur.execute("""
                    INSERT INTO questions (
                        exam_category_id, topic_id, question_type, difficulty,
                        language, body, explanation, estimated_time_seconds,
                        status, published_at
                    )
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    RETURNING id;
                """, (cat_id, topic_id, q_type, diff, lang, body, explanation, est_sec, status, now_utc))

                qid = cur.fetchone()[0]

                for order_idx, opt in enumerate(options, start=1):
                    cur.execute("""
                        INSERT INTO question_options (question_id, option_text, is_correct, order_index)
                        VALUES (%s, %s, %s, %s);
                    """, (qid, opt.get("option_text", ""), bool(opt.get("is_correct", False)), order_idx))

                cur.execute(f"RELEASE SAVEPOINT q_{idx}")
                success_count += 1
            except Exception as e:
                cur.execute(f"ROLLBACK TO SAVEPOINT q_{idx}")
                fail_count += 1
                failed_items.append((idx, q.get("body", "")[:40], str(e).strip()))

            total_done = success_count + fail_count
            if total_done % 50 == 0 or total_done == len(questions):
                print(f"    Progress: {total_done}/{len(questions)} (Success: {success_count}, Failed: {fail_count})")

    conn.commit()
    conn.close()
    return success_count, fail_count, failed_items


def escape_sql(val):
    if val is None:
        return "NULL"
    if isinstance(val, bool):
        return "TRUE" if val else "FALSE"
    if isinstance(val, (int, float)):
        return str(val)
    escaped = str(val).replace("'", "''").replace("\\", "\\\\")
    return f"'{escaped}'"


def sanitize_url_for_psql(db_url):
    """Filter out non-libpq query parameters like default_query_exec_mode."""
    parts = urlsplit(db_url)
    if not parts.query:
        return db_url
    allowed = {
        "sslmode", "sslcompression", "sslcert", "sslkey", "sslrootcert",
        "sslcrl", "requirepeer", "application_name", "connect_timeout",
        "options", "keepalives", "keepalives_idle", "keepalives_interval",
        "keepalives_count", "target_session_attrs"
    }
    query_items = urllib.parse.parse_qsl(parts.query)
    filtered = [(k, v) for k, v in query_items if k.lower() in allowed]
    new_query = urllib.parse.urlencode(filtered)
    return urlunsplit((parts.scheme, parts.netloc, parts.path, new_query, parts.fragment))


def import_via_psql_cli(db_url, questions, publish=False):
    """Fallback: Generate SQL and execute via the local psql command-line client."""
    psql_url = sanitize_url_for_psql(db_url)
    print("[*] Generating SQL script for psql execution...")
    status = "published" if publish else "draft"
    pub_sql = "NOW()" if publish else "NULL"

    sql_lines = ["BEGIN;"]
    for idx, q in enumerate(questions, start=1):
        cat_id = escape_sql(q.get("exam_category_id"))
        topic_id = escape_sql(q.get("topic_id"))
        q_type = escape_sql(q.get("question_type", "mcq_single"))
        diff = escape_sql(q.get("difficulty", "medium"))
        lang = escape_sql(q.get("language", "en"))
        body = escape_sql(q.get("body", ""))
        explanation = escape_sql(q.get("explanation"))
        est_sec = int(q.get("estimated_time_seconds", 60))
        options = q.get("options", [])

        opt_rows = []
        for o_idx, opt in enumerate(options, start=1):
            o_text = escape_sql(opt.get("option_text", ""))
            o_corr = "TRUE" if opt.get("is_correct") else "FALSE"
            opt_rows.append(f"({o_text}, {o_corr}, {o_idx})")

        if opt_rows:
            opt_values = ",\n        ".join(opt_rows)
            stmt = f"""
-- Question #{idx}
WITH ins_q AS (
    INSERT INTO questions (
        exam_category_id, topic_id, question_type, difficulty,
        language, body, explanation, estimated_time_seconds,
        status, published_at
    )
    VALUES ({cat_id}, {topic_id}, {q_type}, {diff}, {lang}, {body}, {explanation}, {est_sec}, '{status}', {pub_sql})
    RETURNING id
)
INSERT INTO question_options (question_id, option_text, is_correct, order_index)
SELECT ins_q.id, o.col1, o.col2::boolean, o.col3::smallint
FROM ins_q, (
    VALUES
        {opt_values}
) AS o(col1, col2, col3);
"""
            sql_lines.append(stmt)

    sql_lines.append("COMMIT;")
    sql_content = "\n".join(sql_lines)

    with tempfile.NamedTemporaryFile("w", suffix=".sql", delete=False) as f:
        f.write(sql_content)
        tmp_sql_path = f.name

    try:
        print(f"[*] Executing SQL through psql ({len(questions)} questions)...")
        proc = subprocess.run(
            ["psql", psql_url, "-v", "ON_ERROR_STOP=1", "-f", tmp_sql_path],
            capture_output=True,
            text=True
        )
        if proc.returncode != 0:
            print(f"[X] psql error: {proc.stderr}")
            return 0, len(questions), [(0, "All", proc.stderr[:200])]
        print("[+] Questions and options imported successfully via psql.")
        return len(questions), 0, []
    finally:
        if os.path.exists(tmp_sql_path):
            os.remove(tmp_sql_path)


# =====================================================================
# HTTP API Importer (Alternative Mode)
# =====================================================================

def login_api(base_url, login_identifier, password):
    """Authenticate with admin credentials over HTTP and return JWT token."""
    login_url = f"{base_url.rstrip('/')}/api/v1/auth/login"
    payload = json.dumps({"login": login_identifier, "password": password}).encode("utf-8")

    req = urllib.request.Request(
        login_url,
        data=payload,
        headers={"Content-Type": "application/json"},
        method="POST"
    )

    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            auth_data = data.get("data") if isinstance(data.get("data"), dict) else data
            token = auth_data.get("token")
            role = auth_data.get("user", {}).get("role")
            if not token:
                sys.exit(f"[X] No token returned in login response: {data}")
            if role != "admin":
                print(f"[!] Warning: User role is '{role}', expected 'admin'.")
            return token
    except urllib.error.HTTPError as e:
        err_msg = e.read().decode("utf-8", errors="replace")
        sys.exit(f"[X] Login failed ({e.code}): {err_msg}")
    except Exception as e:
        sys.exit(f"[X] Failed to connect to {login_url}: {e}")


def create_and_publish_question_api(base_url, token, question, publish=False):
    """POST a question via REST API and optionally publish it."""
    url = f"{base_url.rstrip('/')}/api/v1/admin/questions"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }

    body_data = json.dumps(question).encode("utf-8")
    req = urllib.request.Request(url, data=body_data, headers=headers, method="POST")

    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            resp_data = json.loads(resp.read().decode("utf-8"))
            q_info = resp_data.get("data") if isinstance(resp_data.get("data"), dict) else resp_data
            qid = q_info.get("id")
    except urllib.error.HTTPError as e:
        err_msg = e.read().decode("utf-8", errors="replace")
        return False, f"Create failed ({e.code}): {err_msg}", question.get("body", "")[:40]
    except Exception as e:
        return False, f"Request error: {e}", question.get("body", "")[:40]

    if publish and qid:
        pub_url = f"{base_url.rstrip('/')}/api/v1/admin/questions/{qid}/publish"
        pub_req = urllib.request.Request(pub_url, headers=headers, method="PUT")
        try:
            with urllib.request.urlopen(pub_req, timeout=15) as resp:
                pass
        except Exception as e:
            return True, f"Created ({qid}) but publish failed: {e}", question.get("body", "")[:40]

    return True, f"Success ({qid})", question.get("body", "")[:40]


# =====================================================================
# Main CLI Entrypoint
# =====================================================================

def main():
    parser = argparse.ArgumentParser(
        description="Bulk import questions into Exam Arena via PostgreSQL direct connection or HTTP API."
    )
    # File & execution options
    parser.add_argument("--file", default="question_bank.json", help="Path to JSON file (default: question_bank.json)")
    parser.add_argument("--publish", action="store_true", help="Publish questions immediately (default: draft)")
    parser.add_argument("--limit", type=int, default=0, help="Import only first N questions for testing (0 = all)")

    # Direct PostgreSQL connection options (Default Mode)
    parser.add_argument("--db-user", "--user", "-U", help="PostgreSQL username")
    parser.add_argument("--db-password", "--password", "-W", help="PostgreSQL password")
    parser.add_argument("--db-host", "-H", help="PostgreSQL host (default: from .env or localhost)")
    parser.add_argument("--db-port", "-p", type=int, help="PostgreSQL port (default: from .env or 5432)")
    parser.add_argument("--db-name", "-d", help="PostgreSQL database name (default: from .env or examarena)")
    parser.add_argument("--db-url", help="Direct PostgreSQL connection string / DATABASE_URL")

    # HTTP API options (Alternative Mode)
    parser.add_argument("--api", action="store_true", help="Force import via HTTP REST API instead of direct DB")
    parser.add_argument("--url", default="http://localhost:8080", help="API base URL (default: http://localhost:8080)")
    parser.add_argument("--login", help="Admin username/email for API login")
    parser.add_argument("--token", help="Admin JWT token (skips API login if provided)")
    parser.add_argument("--concurrency", type=int, default=8, help="API concurrency workers (default: 8)")

    args = parser.parse_args()

    # 1. Locate and parse JSON file
    filepath = args.file
    if not os.path.exists(filepath):
        alt_path = os.path.join("..", filepath)
        if os.path.exists(alt_path):
            filepath = alt_path
        else:
            sys.exit(f"[X] Question file not found: {args.file} (or {alt_path})")

    print(f"[*] Reading questions from {filepath}...")
    try:
        with open(filepath, "r", encoding="utf-8") as f:
            questions = json.load(f)
    except Exception as e:
        sys.exit(f"[X] Failed to parse JSON: {e}")

    if not isinstance(questions, list):
        sys.exit("[X] Expected a JSON array of questions.")

    if args.limit > 0:
        questions = questions[:args.limit]
        print(f"[*] Test mode: limited to first {len(questions)} question(s).")

    total = len(questions)
    print(f"[*] Total questions to import: {total}")

    # 2. Determine execution mode: API or Direct DB
    # If user explicitly requested --api or provided --token/--login without DB parameters
    use_api = args.api or (bool(args.token or args.login) and not any([args.db_user, args.db_password, args.db_url]))

    if use_api:
        print("[*] Running in HTTP API mode...")
        token = args.token
        if not token:
            login_id = args.login
            password = args.db_password or args.login
            if not login_id or not password:
                sys.exit("[X] In API mode, you must provide --token OR both --login and --password")
            print(f"[*] Authenticating with API as '{login_id}'...")
            token = login_api(args.url, login_id, password)
            print("[+] Logged in successfully.")

        print(f"[*] Uploading {total} questions with concurrency = {args.concurrency} (publish={args.publish})...")
        success_count = 0
        fail_count = 0
        failed_items = []

        with ThreadPoolExecutor(max_workers=args.concurrency) as executor:
            futures = {
                executor.submit(create_and_publish_question_api, args.url, token, q, args.publish): i
                for i, q in enumerate(questions)
            }
            for future in as_completed(futures):
                idx = futures[future]
                ok, msg, snippet = future.result()
                if ok:
                    success_count += 1
                else:
                    fail_count += 1
                    failed_items.append((idx + 1, snippet, msg))

                completed = success_count + fail_count
                if completed % 25 == 0 or completed == total:
                    print(f"    Progress: {completed}/{total} (Success: {success_count}, Failed: {fail_count})")
    else:
        # Direct Database Mode
        base_env_url = args.db_url or find_env_database_url()
        db_url = build_postgres_url(
            base_url=base_env_url,
            user=args.db_user,
            password=args.db_password,
            host=args.db_host,
            port=args.db_port,
            dbname=args.db_name
        )

        masked_url = db_url
        if "@" in masked_url and ":" in masked_url.split("@")[0]:
            pre, post = masked_url.split("@", 1)
            scheme_user = pre.rsplit(":", 1)[0]
            masked_url = f"{scheme_user}:******@{post}"
        print(f"[*] Target PostgreSQL URL: {masked_url}")

        # Try psycopg2 or psycopg3 first
        db_driver = None
        try:
            import psycopg2
            db_driver = psycopg2
        except ImportError:
            try:
                import psycopg
                db_driver = psycopg
            except ImportError:
                pass

        if db_driver:
            success_count, fail_count, failed_items = import_via_psycopg(
                db_driver, db_url, questions, publish=args.publish
            )
        elif shutil.which("psql"):
            print("[!] Neither 'psycopg2' nor 'psycopg' found in Python. Falling back to 'psql' CLI...")
            success_count, fail_count, failed_items = import_via_psql_cli(
                db_url, questions, publish=args.publish
            )
        else:
            sys.exit(
                "[X] Could not connect to PostgreSQL:\n"
                "    Please install psycopg2 ('pip install psycopg2-binary') or install the 'psql' command-line tool."
            )

    print("\n" + "=" * 50)
    print(f"Finished! Total: {total} | Succeeded: {success_count} | Failed: {fail_count}")

    if failed_items:
        print("\nFailed questions (first 10):")
        for num, snip, err in failed_items[:10]:
            print(f"  - #{num} '{snip}...': {err}")


if __name__ == "__main__":
    main()
