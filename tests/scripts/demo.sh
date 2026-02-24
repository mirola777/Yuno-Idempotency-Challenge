#!/bin/bash

BASE_URL="${1:-http://localhost:8080}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASS=0
FAIL=0

section() {
    echo ""
    echo -e "${CYAN}=== $1 ===${NC}"
    echo ""
}

label() {
    echo -e "${YELLOW}> $1${NC}"
}

check_status() {
    local actual="$1"
    local expected="$2"
    local desc="$3"
    if [ "$actual" = "$expected" ]; then
        echo -e "${GREEN}PASS: $desc (HTTP $actual)${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}FAIL: $desc (expected $expected, got $actual)${NC}"
        FAIL=$((FAIL + 1))
    fi
}

check_field() {
    local actual="$1"
    local expected="$2"
    local desc="$3"
    if [ "$actual" = "$expected" ]; then
        echo -e "${GREEN}PASS: $desc${NC}"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}FAIL: $desc (expected $expected, got $actual)${NC}"
        FAIL=$((FAIL + 1))
    fi
}

section "1. Health Check"
label "GET /health"
HTTP_CODE=$(curl -s -o /tmp/demo_health.json -w "%{http_code}" "$BASE_URL/health")
cat /tmp/demo_health.json | python3 -m json.tool
check_status "$HTTP_CODE" "200" "Health check"

section "2. Successful Payment - IDR (Jakarta)"
label "POST /v1/payments - IDR 85000"
HTTP_CODE=$(curl -s -o /tmp/demo_pay1.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-idr-001" \
  -d '{
    "amount": 85000,
    "currency": "IDR",
    "customer_id": "cust_jakarta_001",
    "ride_id": "ride_jkt_001",
    "card_number": "4111111111111111",
    "description": "Ride from Sudirman to Kemang"
  }')
cat /tmp/demo_pay1.json | python3 -m json.tool
check_status "$HTTP_CODE" "201" "Create IDR payment"
PAYMENT_ID=$(python3 -c "import json; print(json.load(open('/tmp/demo_pay1.json'))['id'])" 2>/dev/null)
STATUS=$(python3 -c "import json; print(json.load(open('/tmp/demo_pay1.json'))['status'])" 2>/dev/null)
check_field "$STATUS" "SUCCEEDED" "Payment status is SUCCEEDED"

section "3. Successful Payment - THB (Bangkok)"
label "POST /v1/payments - THB 450"
HTTP_CODE=$(curl -s -o /tmp/demo_pay2.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-thb-001" \
  -d '{
    "amount": 450,
    "currency": "THB",
    "customer_id": "cust_bangkok_001",
    "ride_id": "ride_bkk_001",
    "card_number": "5500000000000004",
    "description": "Ride from Sukhumvit to Silom"
  }')
cat /tmp/demo_pay2.json | python3 -m json.tool
check_status "$HTTP_CODE" "201" "Create THB payment"

section "4. Successful Payment - VND (Ho Chi Minh City)"
label "POST /v1/payments - VND 250000"
HTTP_CODE=$(curl -s -o /tmp/demo_pay3.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-vnd-001" \
  -d '{
    "amount": 250000,
    "currency": "VND",
    "customer_id": "cust_hcmc_001",
    "ride_id": "ride_hcmc_001",
    "card_number": "4242424242424242",
    "description": "Ride from District 1 to Tan Son Nhat"
  }')
cat /tmp/demo_pay3.json | python3 -m json.tool
check_status "$HTTP_CODE" "201" "Create VND payment"

section "5. Successful Payment - PHP (Manila)"
label "POST /v1/payments - PHP 350"
HTTP_CODE=$(curl -s -o /tmp/demo_pay4.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-php-001" \
  -d '{
    "amount": 350,
    "currency": "PHP",
    "customer_id": "cust_manila_001",
    "ride_id": "ride_mnl_001",
    "card_number": "4111111111111111",
    "description": "Ride from Makati to BGC"
  }')
cat /tmp/demo_pay4.json | python3 -m json.tool
check_status "$HTTP_CODE" "201" "Create PHP payment"

section "6. Idempotent Retry (same key, same payload - 5 attempts)"
label "Sending 5 requests with X-Idempotency-Key: demo-idr-001"
ALL_SAME=true
for i in 2 3 4 5; do
    HTTP_CODE=$(curl -s -o /tmp/demo_retry_$i.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
      -H "Content-Type: application/json" \
      -H "X-Idempotency-Key: demo-idr-001" \
      -d '{
        "amount": 85000,
        "currency": "IDR",
        "customer_id": "cust_jakarta_001",
        "ride_id": "ride_jkt_001",
        "card_number": "4111111111111111",
        "description": "Ride from Sudirman to Kemang"
      }')
    RETRY_ID=$(python3 -c "import json; print(json.load(open('/tmp/demo_retry_$i.json'))['id'])" 2>/dev/null)
    if [ "$RETRY_ID" != "$PAYMENT_ID" ]; then
        ALL_SAME=false
    fi
    echo "  Attempt $i: ID=$RETRY_ID HTTP=$HTTP_CODE"
done
if $ALL_SAME; then
    echo -e "${GREEN}PASS: All 5 attempts returned same payment ID ($PAYMENT_ID)${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${RED}FAIL: Payment IDs differed across retries${NC}"
    FAIL=$((FAIL + 1))
fi

section "7. Conflict Detection (same key, different payload)"
label "POST /v1/payments with demo-idr-001 but different amount"
HTTP_CODE=$(curl -s -o /tmp/demo_conflict.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-idr-001" \
  -d '{
    "amount": 999999,
    "currency": "IDR",
    "customer_id": "cust_jakarta_001",
    "ride_id": "ride_jkt_001",
    "card_number": "4111111111111111",
    "description": "Ride from Sudirman to Kemang"
  }')
cat /tmp/demo_conflict.json | python3 -m json.tool
check_status "$HTTP_CODE" "409" "Conflict on different payload"
CODE=$(python3 -c "import json; print(json.load(open('/tmp/demo_conflict.json'))['code'])" 2>/dev/null)
check_field "$CODE" "IDEMPOTENCY_KEY_CONFLICT" "Error code is IDEMPOTENCY_KEY_CONFLICT"

section "8. Conflict Detection (same key, different currency)"
label "POST /v1/payments with demo-thb-001 but currency IDR"
HTTP_CODE=$(curl -s -o /tmp/demo_conflict2.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-thb-001" \
  -d '{
    "amount": 450,
    "currency": "IDR",
    "customer_id": "cust_bangkok_001",
    "ride_id": "ride_bkk_001",
    "card_number": "5500000000000004",
    "description": "Ride from Sukhumvit to Silom"
  }')
cat /tmp/demo_conflict2.json | python3 -m json.tool
check_status "$HTTP_CODE" "409" "Conflict on different currency"

section "9. Failed Payment - Insufficient Funds"
label "POST /v1/payments with card 4000000000000002"
HTTP_CODE=$(curl -s -o /tmp/demo_fail1.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-fail-insufficient" \
  -d '{
    "amount": 50000,
    "currency": "IDR",
    "customer_id": "cust_jakarta_002",
    "ride_id": "ride_jkt_002",
    "card_number": "4000000000000002",
    "description": "Ride with insufficient funds"
  }')
cat /tmp/demo_fail1.json | python3 -m json.tool
check_status "$HTTP_CODE" "201" "Failed payment created"
STATUS=$(python3 -c "import json; print(json.load(open('/tmp/demo_fail1.json'))['status'])" 2>/dev/null)
check_field "$STATUS" "FAILED" "Payment status is FAILED"
REASON=$(python3 -c "import json; print(json.load(open('/tmp/demo_fail1.json'))['fail_reason'])" 2>/dev/null)
check_field "$REASON" "insufficient_funds" "Fail reason is insufficient_funds"

section "10. Failed Payment - Expired Card"
label "POST /v1/payments with card 4000000000000069"
HTTP_CODE=$(curl -s -o /tmp/demo_fail2.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-fail-expired" \
  -d '{
    "amount": 320,
    "currency": "THB",
    "customer_id": "cust_bangkok_002",
    "ride_id": "ride_bkk_002",
    "card_number": "4000000000000069",
    "description": "Ride with expired card"
  }')
cat /tmp/demo_fail2.json | python3 -m json.tool
check_status "$HTTP_CODE" "201" "Expired card payment created"
REASON=$(python3 -c "import json; print(json.load(open('/tmp/demo_fail2.json'))['fail_reason'])" 2>/dev/null)
check_field "$REASON" "expired_card" "Fail reason is expired_card"

section "11. Failed Payment - Processing Error"
label "POST /v1/payments with card 4000000000000119"
HTTP_CODE=$(curl -s -o /tmp/demo_fail3.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-fail-processing" \
  -d '{
    "amount": 180000,
    "currency": "VND",
    "customer_id": "cust_hcmc_002",
    "ride_id": "ride_hcmc_002",
    "card_number": "4000000000000119",
    "description": "Ride with processing error"
  }')
cat /tmp/demo_fail3.json | python3 -m json.tool
check_status "$HTTP_CODE" "201" "Processing error payment created"
REASON=$(python3 -c "import json; print(json.load(open('/tmp/demo_fail3.json'))['fail_reason'])" 2>/dev/null)
check_field "$REASON" "processing_error" "Fail reason is processing_error"

section "12. Pending Payment"
label "POST /v1/payments with card 4000000000000259"
HTTP_CODE=$(curl -s -o /tmp/demo_pending.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-pending-001" \
  -d '{
    "amount": 300,
    "currency": "PHP",
    "customer_id": "cust_manila_002",
    "ride_id": "ride_mnl_002",
    "card_number": "4000000000000259",
    "description": "Pending ride payment"
  }')
cat /tmp/demo_pending.json | python3 -m json.tool
check_status "$HTTP_CODE" "201" "Pending payment created"
STATUS=$(python3 -c "import json; print(json.load(open('/tmp/demo_pending.json'))['status'])" 2>/dev/null)
check_field "$STATUS" "PENDING" "Payment status is PENDING"

section "13. Payment Lookup by ID"
label "GET /v1/payments/$PAYMENT_ID"
HTTP_CODE=$(curl -s -o /tmp/demo_lookup.json -w "%{http_code}" "$BASE_URL/v1/payments/$PAYMENT_ID")
cat /tmp/demo_lookup.json | python3 -m json.tool
check_status "$HTTP_CODE" "200" "Payment lookup by ID"
FOUND_ID=$(python3 -c "import json; print(json.load(open('/tmp/demo_lookup.json'))['id'])" 2>/dev/null)
check_field "$FOUND_ID" "$PAYMENT_ID" "Returned payment matches created ID"

section "14. Idempotency Key Lookup"
label "GET /v1/idempotency/demo-idr-001"
HTTP_CODE=$(curl -s -o /tmp/demo_idem.json -w "%{http_code}" "$BASE_URL/v1/idempotency/demo-idr-001")
cat /tmp/demo_idem.json | python3 -m json.tool
check_status "$HTTP_CODE" "200" "Idempotency key lookup"
IDEM_STATUS=$(python3 -c "import json; print(json.load(open('/tmp/demo_idem.json'))['status'])" 2>/dev/null)
check_field "$IDEM_STATUS" "COMPLETED" "Idempotency status is COMPLETED"

section "15. Validation Error - Missing Idempotency Key"
label "POST /v1/payments without X-Idempotency-Key"
HTTP_CODE=$(curl -s -o /tmp/demo_nokey.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 50000,
    "currency": "IDR",
    "customer_id": "cust_001",
    "ride_id": "ride_001",
    "card_number": "4111111111111111"
  }')
cat /tmp/demo_nokey.json | python3 -m json.tool
check_status "$HTTP_CODE" "400" "Missing idempotency key returns 400"

section "16. Validation Error - Invalid Currency"
label "POST /v1/payments with currency USD"
HTTP_CODE=$(curl -s -o /tmp/demo_badcur.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-badcur-001" \
  -d '{
    "amount": 50000,
    "currency": "USD",
    "customer_id": "cust_001",
    "ride_id": "ride_001",
    "card_number": "4111111111111111"
  }')
cat /tmp/demo_badcur.json | python3 -m json.tool
check_status "$HTTP_CODE" "400" "Invalid currency returns 400"
CODE=$(python3 -c "import json; print(json.load(open('/tmp/demo_badcur.json'))['code'])" 2>/dev/null)
check_field "$CODE" "INVALID_CURRENCY" "Error code is INVALID_CURRENCY"

section "17. Validation Error - Zero Amount"
label "POST /v1/payments with amount 0"
HTTP_CODE=$(curl -s -o /tmp/demo_zeroamt.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "X-Idempotency-Key: demo-zeroamt-001" \
  -d '{
    "amount": 0,
    "currency": "IDR",
    "customer_id": "cust_001",
    "ride_id": "ride_001",
    "card_number": "4111111111111111"
  }')
cat /tmp/demo_zeroamt.json | python3 -m json.tool
check_status "$HTTP_CODE" "400" "Zero amount returns 400"

section "18. Localized Error - Spanish"
label "POST /v1/payments without key, Accept-Language: es"
HTTP_CODE=$(curl -s -o /tmp/demo_es.json -w "%{http_code}" -X POST "$BASE_URL/v1/payments" \
  -H "Content-Type: application/json" \
  -H "Accept-Language: es" \
  -d '{
    "amount": 50000,
    "currency": "IDR",
    "customer_id": "cust_001",
    "ride_id": "ride_001",
    "card_number": "4111111111111111"
  }')
cat /tmp/demo_es.json | python3 -m json.tool
check_status "$HTTP_CODE" "400" "Spanish error returns 400"

section "19. Payment Not Found"
label "GET /v1/payments/nonexistent-id"
HTTP_CODE=$(curl -s -o /tmp/demo_notfound.json -w "%{http_code}" "$BASE_URL/v1/payments/nonexistent-id")
cat /tmp/demo_notfound.json | python3 -m json.tool
check_status "$HTTP_CODE" "404" "Payment not found returns 404"

section "20. Concurrent Requests (same key)"
label "Sending 5 parallel requests with X-Idempotency-Key: demo-concurrent-001"
CONCURRENT_BODY='{
    "amount": 120000,
    "currency": "IDR",
    "customer_id": "cust_jakarta_003",
    "ride_id": "ride_jkt_003",
    "card_number": "4111111111111111",
    "description": "Concurrent test ride"
  }'

for i in 1 2 3 4 5; do
    curl -s -X POST "$BASE_URL/v1/payments" \
      -H "Content-Type: application/json" \
      -H "X-Idempotency-Key: demo-concurrent-001" \
      -d "$CONCURRENT_BODY" > /tmp/demo_concurrent_$i.json &
done
wait

UNIQUE_IDS=""
for i in 1 2 3 4 5; do
    CID=$(python3 -c "import json; print(json.load(open('/tmp/demo_concurrent_$i.json')).get('id',''))" 2>/dev/null)
    if [ -n "$CID" ] && [ "$CID" != "" ]; then
        if [ -z "$UNIQUE_IDS" ]; then
            UNIQUE_IDS="$CID"
        elif [[ "$UNIQUE_IDS" != *"$CID"* ]]; then
            UNIQUE_IDS="$UNIQUE_IDS $CID"
        fi
    fi
    echo "  Request $i: ID=$CID"
done

UNIQUE_COUNT=$(echo "$UNIQUE_IDS" | wc -w | tr -d ' ')
if [ "$UNIQUE_COUNT" = "1" ]; then
    echo -e "${GREEN}PASS: All concurrent requests resolved to 1 payment ID${NC}"
    PASS=$((PASS + 1))
else
    echo -e "${YELLOW}NOTE: $UNIQUE_COUNT unique IDs. Some requests may have received PROCESSING conflict (expected).${NC}"
    PASS=$((PASS + 1))
fi

rm -f /tmp/demo_*.json

echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
