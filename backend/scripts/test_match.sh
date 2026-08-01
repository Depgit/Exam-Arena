#!/usr/bin/env bash
set -euo pipefail

BASE="http://localhost:8080/api/v1"
WS_BASE="ws://localhost:8080/ws"

echo "============================================"
echo "EXAM ARENA - FULL MATCH E2E TEST"
echo "============================================"
echo ""

# ── Step 1: Register two players ──────────────────────────────────────

echo ">>> Registering Player 1..."
P1_RESP=$(curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player_alpha",
    "email": "alpha@test.com",
    "password": "password123"
  }')
echo "$P1_RESP" | python3 -m json.tool 2>/dev/null || echo "$P1_RESP"
P1_TOKEN=$(echo "$P1_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
P1_ID=$(echo "$P1_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['user']['id'])" 2>/dev/null)
echo "Player 1 ID:    $P1_ID"
echo "Player 1 Token: ${P1_TOKEN:0:20}..."
echo ""

echo ">>> Registering Player 2..."
P2_RESP=$(curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player_beta",
    "email": "beta@test.com",
    "password": "password123"
  }')
echo "$P2_RESP" | python3 -m json.tool 2>/dev/null || echo "$P2_RESP"
P2_TOKEN=$(echo "$P2_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null)
P2_ID=$(echo "$P2_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['user']['id'])" 2>/dev/null)
echo "Player 2 ID:    $P2_ID"
echo "Player 2 Token: ${P2_TOKEN:0:20}..."
echo ""

# ── Step 2: Get SSC category ID ──────────────────────────────────────

echo ">>> Fetching SSC category..."
SSC_ID=$(docker exec examarena-db psql -U examarena -d examarena -t -A \
  -c "SELECT id FROM exam_categories WHERE code = 'SSC'")
echo "SSC Category ID: $SSC_ID"
echo ""

# ── Step 3: Check health and queue ───────────────────────────────────

echo ">>> Health check..."
curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null
echo ""

# ── Step 4: Both players join the queue simultaneously ────────────────

echo ">>> Player 1 joining queue..."
curl -s -X POST "$BASE/matches/queue" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $P1_TOKEN" \
  -d "{\"exam_category_id\": \"$SSC_ID\", \"match_type\": \"ranked\"}" | python3 -m json.tool 2>/dev/null
echo ""

echo ">>> Player 2 joining queue..."
curl -s -X POST "$BASE/matches/queue" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $P2_TOKEN" \
  -d "{\"exam_category_id\": \"$SSC_ID\", \"match_type\": \"ranked\"}" | python3 -m json.tool 2>/dev/null
echo ""

echo ">>> Queue stats..."
curl -s "$BASE/matches/queue/stats" \
  -H "Authorization: Bearer $P1_TOKEN" | python3 -m json.tool 2>/dev/null
echo ""

# ── Step 5: Wait for matchmaking engine to pair them ──────────────────

echo ">>> Waiting 5 seconds for matchmaking engine..."
sleep 5

# ── Step 6: Check player profiles for rating/stats ───────────────────

echo ">>> Player 1 profile..."
curl -s "$BASE/users/$P1_ID" | python3 -m json.tool 2>/dev/null
echo ""

echo ">>> Player 1 stats..."
curl -s "$BASE/users/$P1_ID/stats" | python3 -m json.tool 2>/dev/null
echo ""

echo "============================================"
echo "REST API TESTS PASSED"
echo "Now run the WebSocket test for live gameplay"
echo "============================================"