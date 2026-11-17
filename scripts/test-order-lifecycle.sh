#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"

for dependency in curl jq; do
    if ! command -v "$dependency" >/dev/null 2>&1; then
        printf 'Missing required command: %s\n' "$dependency" >&2
        exit 1
    fi
done

TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT
RESPONSE_FILE="$TEMP_DIR/response.json"

request() {
    local method="$1"
    local path="$2"
    local expected_status="$3"
    shift 3
    local actual_status

    printf '\n>>> %s %s\n' "$method" "$path"
    if ! actual_status=$(curl --verbose --silent --show-error \
        --connect-timeout 5 --max-time 30 \
        --request "$method" \
        --header 'Accept: application/json' \
        --output "$RESPONSE_FILE" \
        --write-out '%{http_code}' \
        "$@" "$BASE_URL$path"); then
        printf 'Request failed. Is the server running at %s?\n' "$BASE_URL" >&2
        exit 1
    fi

    if [[ -s "$RESPONSE_FILE" ]]; then
        if jq empty "$RESPONSE_FILE" 2>/dev/null; then
            jq . "$RESPONSE_FILE"
        else
            cat "$RESPONSE_FILE"
        fi
    fi

    if [[ "$actual_status" != "$expected_status" ]]; then
        printf 'Expected HTTP %s, received %s.\n' "$expected_status" "$actual_status" >&2
        exit 1
    fi
}

assert_order() {
    local expected_status="$1"
    if ! jq --exit-status \
        --arg id "$ORDER_ID" --arg status "$expected_status" \
        '.id == $id and .status == $status' "$RESPONSE_FILE" >/dev/null; then
        printf 'Expected order %s with status %s.\n' "$ORDER_ID" "$expected_status" >&2
        exit 1
    fi
}

request POST /orders 201 \
    --header 'Content-Type: application/json' \
    --data '{
        "customer_id": "rest-test-customer",
        "items": [
            {"product_id": "product-1", "quantity": 2, "price_cents": 1500, "currency": "USD"},
            {"product_id": "product-2", "quantity": 1, "price_cents": 2500, "currency": "USD"}
        ]
    }'

ORDER_ID=$(jq --exit-status --raw-output \
    '.id | select(type == "string" and length > 0)' "$RESPONSE_FILE")
assert_order pending
ORDER_PATH="/orders/$(jq --null-input --raw-output --arg id "$ORDER_ID" '$id | @uri')"

request GET "$ORDER_PATH" 200
assert_order pending

request POST "$ORDER_PATH/confirm" 204
request GET "$ORDER_PATH" 200
assert_order confirmed

request POST "$ORDER_PATH/cancel" 204
request GET "$ORDER_PATH" 200
assert_order canceled

printf '\nOrder lifecycle passed: %s (pending -> confirmed -> canceled).\n' "$ORDER_ID"