#!/usr/bin/env bash
set -euo pipefail

BASE="http://localhost:8080/api/v1"

# Get SSC ID
SSC_ID=$(docker exec examarena-db psql -U examarena -d examarena -t -A \
  -c "SELECT id FROM exam_categories WHERE code = 'SSC'")

echo "SSC Category: $SSC_ID"

# Register players
P1=$(curl -s -X POST "$BASE/auth/register" -H "Content-Type: application/json" \
  -d '{"username":"friend_p1","email":"fp1@test.com","password":"password123"}')
P1_TOKEN=$(echo "$P1" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

P2=$(curl -s -X POST "$BASE/auth/register" -H "Content-Type: application/json" \
  -d '{"username":"friend_p2","email":"fp2@test.com","password":"password123"}')
P2_TOKEN=$(echo "$P2" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")

echo "Player 1 token: ${P1_TOKEN:0:20}..."
echo "Player 2 token: ${P2_TOKEN:0:20}..."

# Player 1 creates a friend match
echo ""
echo ">>> Player 1 creating friend match..."
CREATE_RESP=$(curl -s -X POST "$BASE/matches/friend" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $P1_TOKEN" \
  -d "{\"exam_category_id\": \"$SSC_ID\"}")
echo "$CREATE_RESP" | python3 -m json.tool

ROOM_CODE=$(echo "$CREATE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['room_code'])")
MATCH_ID=$(echo "$CREATE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['match_id'])")
echo "Room Code: $ROOM_CODE"
echo "Match ID:  $MATCH_ID"

# Player 2 joins the friend match
echo ""
echo ">>> Player 2 joining with room code $ROOM_CODE..."
JOIN_RESP=$(curl -s -X POST "$BASE/matches/friend/join" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $P2_TOKEN" \
  -d "{\"room_code\": \"$ROOM_CODE\"}")
echo "$JOIN_RESP" | python3 -m json.tool

echo ""
echo ">>> Match should be starting now!"
echo ">>> Connect via WebSocket to play:"
echo ""
echo "Player 1: wscat -c 'ws://localhost:8080/ws?token=$P1_TOKEN'"
echo "Player 2: wscat -c 'ws://localhost:8080/ws?token=$P2_TOKEN'"
echo ""
echo ">>> Check match details:"
echo "curl -s '$BASE/matches/$MATCH_ID' -H 'Authorization: Bearer $P1_TOKEN' | python3 -m json.tool"