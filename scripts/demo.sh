#!/bin/bash

BASE_URL="${1:-http://localhost:8080}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

section() {
    echo ""
    echo -e "${CYAN}=== $1 ===${NC}"
    echo ""
}

label() {
    echo -e "${YELLOW}> $1${NC}"
}

section "1. Health Check"
label "GET /health"
curl -s "$BASE_URL/health" | python3 -m json.tool
echo ""

section "2. Create Payment (first request)"
label "POST /v1/payments with X-Idempotency-Key: demo-key-001"
RESPONSE=$(curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-key-001" \
  -d '{
    "amount": 85000,
    "currency": "IDR",
    "customer_id": "cust_jakarta_001",
    "ride_id": "ride_jkt_demo_001",
    "card_number": "4111111111111111",
    "description": "Ride from Sudirman to Kemang"
  }')
echo "$RESPONSE" | python3 -m json.tool
PAYMENT_ID=$(echo "$RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo -e "${GREEN}Payment ID: $PAYMENT_ID${NC}"

section "3. Idempotent Retry (same key, same payload)"
label "POST /v1/payments with X-Idempotency-Key: demo-key-001 (retry)"
RETRY_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-key-001" \
  -d '{
    "amount": 85000,
    "currency": "IDR",
    "customer_id": "cust_jakarta_001",
    "ride_id": "ride_jkt_demo_001",
    "card_number": "4111111111111111",
    "description": "Ride from Sudirman to Kemang"
  }')
echo "$RETRY_RESPONSE" | python3 -m json.tool
RETRY_ID=$(echo "$RETRY_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
if [ "$PAYMENT_ID" = "$RETRY_ID" ]; then
    echo -e "${GREEN}PASS: Same payment ID returned ($RETRY_ID)${NC}"
else
    echo -e "${RED}FAIL: Different payment IDs ($PAYMENT_ID vs $RETRY_ID)${NC}"
fi

section "4. Conflict Detection (same key, different payload)"
label "POST /v1/payments with X-Idempotency-Key: demo-key-001 but different amount"
curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-key-001" \
  -d '{
    "amount": 999999,
    "currency": "IDR",
    "customer_id": "cust_jakarta_001",
    "ride_id": "ride_jkt_demo_001",
    "card_number": "4111111111111111",
    "description": "Ride from Sudirman to Kemang"
  }' | python3 -m json.tool
echo -e "${GREEN}Expected: 409 IDEMPOTENCY_KEY_CONFLICT${NC}"

section "5. Failed Payment (insufficient funds)"
label "POST /v1/payments with card 4000000000000002"
curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-key-fail-001" \
  -d '{
    "amount": 50000,
    "currency": "IDR",
    "customer_id": "cust_jakarta_002",
    "ride_id": "ride_jkt_demo_002",
    "card_number": "4000000000000002",
    "description": "Ride with insufficient funds"
  }' | python3 -m json.tool

section "6. Pending Payment"
label "POST /v1/payments with card 4000000000000259"
curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-key-pending-001" \
  -d '{
    "amount": 300,
    "currency": "PHP",
    "customer_id": "cust_manila_001",
    "ride_id": "ride_mnl_demo_001",
    "card_number": "4000000000000259",
    "description": "Pending ride payment"
  }' | python3 -m json.tool

section "7. Payment Lookup by ID"
label "GET /v1/payments/$PAYMENT_ID"
curl -s "$BASE_URL/v1/payments/$PAYMENT_ID" | python3 -m json.tool

section "8. Idempotency Key Lookup"
label "GET /v1/idempotency/demo-key-001"
curl -s "$BASE_URL/v1/idempotency/demo-key-001" | python3 -m json.tool

section "9. Missing Idempotency Key (validation error)"
label "POST /v1/payments without X-Idempotency-Key header"
curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 50000,
    "currency": "IDR",
    "customer_id": "cust_001",
    "ride_id": "ride_001",
    "card_number": "4111111111111111"
  }' | python3 -m json.tool

section "10. Concurrent Requests (same key)"
label "Sending 3 parallel requests with X-Idempotency-Key: demo-key-concurrent"
CONCURRENT_BODY='{
    "amount": 120000,
    "currency": "IDR",
    "customer_id": "cust_jakarta_003",
    "ride_id": "ride_jkt_demo_003",
    "card_number": "4111111111111111",
    "description": "Concurrent test ride"
  }'

curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-key-concurrent" \
  -d "$CONCURRENT_BODY" > /tmp/concurrent_1.json &

curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-key-concurrent" \
  -d "$CONCURRENT_BODY" > /tmp/concurrent_2.json &

curl -s -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-key-concurrent" \
  -d "$CONCURRENT_BODY" > /tmp/concurrent_3.json &

wait

echo "Response 1:"
cat /tmp/concurrent_1.json | python3 -m json.tool
echo ""
echo "Response 2:"
cat /tmp/concurrent_2.json | python3 -m json.tool
echo ""
echo "Response 3:"
cat /tmp/concurrent_3.json | python3 -m json.tool

ID1=$(python3 -c "import json; print(json.load(open('/tmp/concurrent_1.json')).get('id',''))" 2>/dev/null)
ID2=$(python3 -c "import json; print(json.load(open('/tmp/concurrent_2.json')).get('id',''))" 2>/dev/null)
ID3=$(python3 -c "import json; print(json.load(open('/tmp/concurrent_3.json')).get('id',''))" 2>/dev/null)

echo ""
if [ -n "$ID1" ] && [ "$ID1" = "$ID2" ] && [ "$ID2" = "$ID3" ]; then
    echo -e "${GREEN}PASS: All 3 concurrent requests returned the same payment ID ($ID1)${NC}"
elif [ -n "$ID1" ] && [ "$ID1" = "$ID2" ]; then
    echo -e "${YELLOW}PARTIAL: 2 of 3 returned same ID. One may have hit PROCESSING state (expected).${NC}"
else
    echo -e "${YELLOW}NOTE: Concurrent responses may vary. Check payment IDs above.${NC}"
fi

rm -f /tmp/concurrent_1.json /tmp/concurrent_2.json /tmp/concurrent_3.json

echo ""
echo -e "${CYAN}=== Demo Complete ===${NC}"
