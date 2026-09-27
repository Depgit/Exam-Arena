#!/usr/bin/env python3
"""
Bulk import questions from a JSON file into Exam Arena API.
Uses only Python standard library (no pip dependencies required).
"""

import argparse
import json
import os
import sys
import urllib.request
import urllib.error
from concurrent.futures import ThreadPoolExecutor, as_completed


def login(base_url, login_identifier, password):
    """Authenticate with admin credentials and return the JWT token."""
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
            # Backend wraps responses in {"success": true, "data": {...}}
            auth_data = data.get("data") if isinstance(data.get("data"), dict) else data
            token = auth_data.get("token")
            role = auth_data.get("user", {}).get("role")
            if not token:
                sys.exit(f"[X] No token returned in login response: {data}")
            if role != "admin":
                print(f"[!] Warning: User role is '{role}', expected 'admin'. Request may be forbidden.")
            return token
    except urllib.error.HTTPError as e:
        err_msg = e.read().decode("utf-8", errors="replace")
        sys.exit(f"[X] Login failed ({e.code}): {err_msg}")
    except Exception as e:
        sys.exit(f"[X] Failed to connect to {login_url}: {e}")


def create_and_publish_question(base_url, token, question, publish=False):
    """POST a question and optionally publish it."""
    url = f"{base_url.rstrip('/')}/api/v1/admin/questions"
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }

    # 1. Create Question
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

    # 2. Optionally publish
    if publish and qid:
        pub_url = f"{base_url.rstrip('/')}/api/v1/admin/questions/{qid}/publish"
        pub_req = urllib.request.Request(pub_url, headers=headers, method="PUT")
        try:
            with urllib.request.urlopen(pub_req, timeout=15) as resp:
                pass
        except Exception as e:
            return True, f"Created ({qid}) but publish failed: {e}", question.get("body", "")[:40]

    return True, f"Success ({qid})", question.get("body", "")[:40]


def main():
    parser = argparse.ArgumentParser(description="Bulk import questions into Exam Arena API")
    parser.add_argument("--url", default="http://localhost:8080", help="Base API URL (default: http://localhost:8080)")
    parser.add_argument("--file", default="question_bank.json", help="Path to question bank JSON file (default: question_bank.json)")
    parser.add_argument("--login", help="Admin username or email")
    parser.add_argument("--password", help="Admin password")
    parser.add_argument("--token", help="Admin JWT token (skips login step if provided)")
    parser.add_argument("--publish", action="store_true", help="Automatically publish each question after creation (default questions are created as 'draft')")
    parser.add_argument("--concurrency", type=int, default=8, help="Number of concurrent workers (default: 8)")
    parser.add_argument("--limit", type=int, default=0, help="Import only first N questions for testing (0 = import all)")

    args = parser.parse_args()

    # Locate JSON file (support both project root and scripts/ directory)
    filepath = args.file
    if not os.path.exists(filepath):
        alt_path = os.path.join("..", filepath)
        if os.path.exists(alt_path):
            filepath = alt_path
        else:
            sys.exit(f"[X] File not found: {args.file} (or {alt_path})")

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

    # Get Token
    token = args.token
    if not token:
        if not args.login or not args.password:
            sys.exit("[X] You must provide either --token OR both --login and --password")
        print(f"[*] Authenticating as '{args.login}'...")
        token = login(args.url, args.login, args.password)
        print("[+] Logged in successfully.")

    # Bulk insert
    print(f"[*] Uploading {total} questions with concurrency = {args.concurrency} (publish={args.publish})...")
    success_count = 0
    fail_count = 0
    failed_items = []

    with ThreadPoolExecutor(max_workers=args.concurrency) as executor:
        futures = {
            executor.submit(create_and_publish_question, args.url, token, q, args.publish): i
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

            # Progress update
            completed = success_count + fail_count
            if completed % 25 == 0 or completed == total:
                print(f"    Progress: {completed}/{total} (Success: {success_count}, Failed: {fail_count})")

    print("\n" + "=" * 50)
    print(f"Finished! Total: {total} | Succeeded: {success_count} | Failed: {fail_count}")

    if failed_items:
        print("\nFailed questions (first 10):")
        for num, snip, err in failed_items[:10]:
            print(f"  - #{num} '{snip}...': {err}")


if __name__ == "__main__":
    main()
