#!/usr/bin/env bash
# E2E smoke per api-frontend.md — outputs JSON lines for report
set -euo pipefail

BASE="${BASE_URL:-http://localhost:8080}"
TS=$(date +%s)
EMAIL="e2e-${TS}@example.com"
PASS="test-pass-12345"
CUSTOM="E2E${TS}"
NEW_CUSTOM="E2E${TS}New"

record() { printf '%s\n' "$1"; }

# --- A: Guest ---
A1_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" "${BASE}/health")
A1_BODY=$(jq -c . /tmp/r.json 2>/dev/null || echo "{}")
if [[ "$A1_HTTP" == "200" && "$A1_BODY" == *'"status":"ok"'* ]]; then
  record "A1|PASS|200|${A1_BODY}"
else
  record "A1|FAIL|${A1_HTTP}|${A1_BODY}"
fi

GUEST_RESP=$(curl -s -w "\n%{http_code}" -X POST "${BASE}/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/guest"}')
GUEST_HTTP=$(echo "$GUEST_RESP" | tail -1)
GUEST_BODY=$(echo "$GUEST_RESP" | sed '$d')
GUEST_CODE=$(echo "$GUEST_BODY" | jq -r '.code // empty')
if [[ "$GUEST_HTTP" == "201" && -n "$GUEST_CODE" ]]; then
  record "A2|PASS|201|code=${GUEST_CODE}"
else
  record "A2|FAIL|${GUEST_HTTP}|no code"
fi

REDIR=$(curl -s -o /dev/null -w "%{http_code}\n%{redirect_url}" "${BASE}/${GUEST_CODE}")
REDIR_HTTP=$(echo "$REDIR" | head -1)
REDIR_LOC=$(echo "$REDIR" | tail -1)
if [[ "$REDIR_HTTP" == "302" && "$REDIR_LOC" == *"example.com/guest"* ]]; then
  record "A3|PASS|302|Location ok"
else
  record "A3|FAIL|${REDIR_HTTP}|${REDIR_LOC}"
fi

A4_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" "${BASE}/api/links")
A4_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
if [[ "$A4_HTTP" == "401" && "$A4_CODE" == "unauthorized" ]]; then
  record "A4|PASS|401|unauthorized"
else
  record "A4|FAIL|${A4_HTTP}|${A4_CODE}"
fi

# --- B: Auth ---
REG=$(curl -s -w "\n%{http_code}" -X POST "${BASE}/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\"}")
REG_HTTP=$(echo "$REG" | tail -1)
REG_BODY=$(echo "$REG" | sed '$d')
TOKEN=$(echo "$REG_BODY" | jq -r '.access_token // empty')
if [[ "$REG_HTTP" == "201" && -n "$TOKEN" ]]; then
  record "B1|PASS|201|token received"
else
  record "B1|FAIL|${REG_HTTP}|no token"
fi

SHORT=$(curl -s -w "\n%{http_code}" -X POST "${BASE}/shorten" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"url\":\"https://example.com/user\",\"custom_code\":\"${CUSTOM}\"}")
SHORT_HTTP=$(echo "$SHORT" | tail -1)
SHORT_BODY=$(echo "$SHORT" | sed '$d')
USER_CODE=$(echo "$SHORT_BODY" | jq -r '.code // empty')
if [[ "$SHORT_HTTP" == "201" && "$USER_CODE" == "$CUSTOM" ]]; then
  record "B2|PASS|201|code=${USER_CODE}"
else
  record "B2|FAIL|${SHORT_HTTP}|code=${USER_CODE}"
fi

# A5 after B1
A5_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" \
  -H "Authorization: Bearer ${TOKEN}" \
  "${BASE}/api/links/${GUEST_CODE}/stats")
A5_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
if [[ "$A5_HTTP" == "403" && "$A5_CODE" == "link_forbidden" ]]; then
  record "A5|PASS|403|link_forbidden"
else
  record "A5|FAIL|${A5_HTTP}|${A5_CODE}"
fi

LIST=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer ${TOKEN}" "${BASE}/api/links")
LIST_HTTP=$(echo "$LIST" | tail -1)
LIST_BODY=$(echo "$LIST" | sed '$d')
HAS_USER=$(echo "$LIST_BODY" | jq --arg c "$USER_CODE" '[.links[].code] | index($c) != null')
HAS_GUEST=$(echo "$LIST_BODY" | jq --arg c "$GUEST_CODE" '[.links[].code] | index($c) != null')
QR_URL=$(echo "$LIST_BODY" | jq -r --arg c "$USER_CODE" '.links[] | select(.code==$c) | .qr_url')
if [[ "$LIST_HTTP" == "200" && "$HAS_USER" == "true" && "$HAS_GUEST" == "false" && "$QR_URL" == "/api/links/${USER_CODE}/qr" ]]; then
  record "A6|PASS|200|guest not in list"
  record "B3|PASS|200|qr_url=${QR_URL}"
else
  record "A6|FAIL|${LIST_HTTP}|user=${HAS_USER} guest=${HAS_GUEST}"
  record "B3|FAIL|${LIST_HTTP}|qr_url=${QR_URL}"
fi

curl -s -o /dev/null "${BASE}/${USER_CODE}"
curl -s -o /dev/null "${BASE}/${USER_CODE}"
STATS=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer ${TOKEN}" \
  "${BASE}/api/links/${USER_CODE}/stats?limit=10")
STATS_HTTP=$(echo "$STATS" | tail -1)
STATS_BODY=$(echo "$STATS" | sed '$d')
CLICKS=$(echo "$STATS_BODY" | jq -r '.clicks_count // 0')
VISITS_LEN=$(echo "$STATS_BODY" | jq '.visits | length')
if [[ "$STATS_HTTP" == "200" && "$CLICKS" -ge 2 && "$VISITS_LEN" -ge 1 ]]; then
  record "B4|PASS|302|2 redirects"
  record "B5|PASS|200|clicks=${CLICKS} visits=${VISITS_LEN}"
else
  record "B4|FAIL|-|clicks=${CLICKS}"
  record "B5|FAIL|${STATS_HTTP}|clicks=${CLICKS} visits=${VISITS_LEN}"
fi

curl -s -o /tmp/qr.png -w "%{http_code}" -H "Authorization: Bearer ${TOKEN}" \
  "${BASE}/api/links/${USER_CODE}/qr?size=256" > /tmp/qr_http.txt
QR_HTTP=$(cat /tmp/qr_http.txt)
if [[ "$QR_HTTP" == "200" ]] && xxd /tmp/qr.png 2>/dev/null | head -1 | grep -q "8950 4e47"; then
  record "B6|PASS|200|PNG ok"
else
  record "B6|FAIL|${QR_HTTP}|PNG check"
fi

PATCH=$(curl -s -w "\n%{http_code}" -X PATCH "${BASE}/api/links/${USER_CODE}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"code\":\"${NEW_CUSTOM}\",\"original_url\":\"https://example.com/new\"}")
PATCH_HTTP=$(echo "$PATCH" | tail -1)
PATCH_BODY=$(echo "$PATCH" | sed '$d')
NEW_CODE=$(echo "$PATCH_BODY" | jq -r '.code // empty')
OLD_REDIR=$(curl -s -o /dev/null -w "%{http_code}" "${BASE}/${USER_CODE}")
if [[ "$PATCH_HTTP" == "200" && "$NEW_CODE" == "$NEW_CUSTOM" && "$OLD_REDIR" == "404" ]]; then
  record "B7|PASS|200|old code 404"
else
  record "B7|FAIL|${PATCH_HTTP}|new=${NEW_CODE} old_redir=${OLD_REDIR}"
fi

QR2_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer ${TOKEN}" \
  "${BASE}/api/links/${NEW_CODE}/qr")
if [[ "$QR2_HTTP" == "200" ]]; then
  record "B8|PASS|200|"
else
  record "B8|FAIL|${QR2_HTTP}|"
fi

DEL_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
  -H "Authorization: Bearer ${TOKEN}" \
  "${BASE}/api/links/${NEW_CODE}")
if [[ "$DEL_HTTP" == "204" ]]; then
  record "B9|PASS|204|"
else
  record "B9|FAIL|${DEL_HTTP}|"
fi

AFTER_DEL=$(curl -s -o /tmp/r.json -w "%{http_code}" "${BASE}/${NEW_CODE}")
AFTER_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
if [[ "$AFTER_DEL" == "404" && "$AFTER_CODE" == "link_not_found" ]]; then
  record "B10|PASS|404|"
else
  record "B10|FAIL|${AFTER_DEL}|${AFTER_CODE}"
fi

LIST2=$(curl -s -H "Authorization: Bearer ${TOKEN}" "${BASE}/api/links")
HAS_NEW=$(echo "$LIST2" | jq --arg c "$NEW_CODE" '[.links[].code] | index($c) != null')
if [[ "$HAS_NEW" == "false" ]]; then
  record "B11|PASS|200|removed from list"
else
  record "B11|FAIL|-|still in list"
fi

REUSE=$(curl -s -w "\n%{http_code}" -X POST "${BASE}/shorten" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d "{\"url\":\"https://example.com/reuse\",\"custom_code\":\"${NEW_CUSTOM}\"}")
REUSE_HTTP=$(echo "$REUSE" | tail -1)
if [[ "$REUSE_HTTP" == "201" ]]; then
  record "B12|PASS|201|alias reused"
else
  record "B12|FAIL|${REUSE_HTTP}|"
fi

LOGIN=$(curl -s -w "\n%{http_code}" -X POST "${BASE}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\"}")
LOGIN_HTTP=$(echo "$LOGIN" | tail -1)
LOGIN_BODY=$(echo "$LOGIN" | sed '$d')
TOKEN2=$(echo "$LOGIN_BODY" | jq -r '.access_token // empty')
if [[ "$LOGIN_HTTP" == "200" && -n "$TOKEN2" ]]; then
  record "C1|PASS|200|new token"
else
  record "C1|FAIL|${LOGIN_HTTP}|"
fi

C2_HTTP=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer ${TOKEN}" "${BASE}/api/links")
if [[ "$C2_HTTP" == "200" ]]; then
  record "C2|PASS|200|old token still valid"
else
  record "C2|FAIL|${C2_HTTP}|"
fi

# --- Negatives (need a live link for some) ---
NEG_LINK_CODE="${NEW_CUSTOM}"

N1_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -X POST "${BASE}/shorten" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid" \
  -d '{"url":"https://example.com/x"}')
N1_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N1_HTTP" == "401" && "$N1_CODE" == "unauthorized" ]] && record "N1|PASS|401|" || record "N1|FAIL|${N1_HTTP}|${N1_CODE}"

N2_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -X POST "${BASE}/shorten" \
  -H "Content-Type: application/json" -d '{}')
N2_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N2_HTTP" == "400" && "$N2_CODE" == "url_required" ]] && record "N2|PASS|400|" || record "N2|FAIL|${N2_HTTP}|${N2_CODE}"

N3_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -X POST "${BASE}/shorten" \
  -H "Content-Type: application/json" -d '{"url":"not-a-url"}')
N3_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N3_HTTP" == "400" && "$N3_CODE" == "invalid_url" ]] && record "N3|PASS|400|" || record "N3|FAIL|${N3_HTTP}|${N3_CODE}"

N4_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -X POST "${BASE}/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/x","custom_code":"bad-code!"}')
N4_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N4_HTTP" == "400" && "$N4_CODE" == "invalid_code" ]] && record "N4|PASS|400|" || record "N4|FAIL|${N4_HTTP}|${N4_CODE}"

DUP_CODE="Dup${TS}"
curl -s -o /dev/null -X POST "${BASE}/shorten" -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN2}" \
  -d "{\"url\":\"https://example.com/d1\",\"custom_code\":\"${DUP_CODE}\"}"
N5_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -X POST "${BASE}/shorten" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN2}" \
  -d "{\"url\":\"https://example.com/d2\",\"custom_code\":\"${DUP_CODE}\"}")
N5_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N5_HTTP" == "409" && "$N5_CODE" == "code_already_exists" ]] && record "N5|PASS|409|" || record "N5|FAIL|${N5_HTTP}|${N5_CODE}"

N6_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -X POST "${BASE}/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\"}")
N6_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N6_HTTP" == "409" && "$N6_CODE" == "email_already_exists" ]] && record "N6|PASS|409|" || record "N6|FAIL|${N6_HTTP}|${N6_CODE}"

N7_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -X POST "${BASE}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"wrong-password\"}")
N7_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N7_HTTP" == "401" && "$N7_CODE" == "invalid_credentials" ]] && record "N7|PASS|401|" || record "N7|FAIL|${N7_HTTP}|${N7_CODE}"

N8_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" "${BASE}/api/links")
N8_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N8_HTTP" == "401" && "$N8_CODE" == "unauthorized" ]] && record "N8|PASS|401|" || record "N8|FAIL|${N8_HTTP}|${N8_CODE}"

N9_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -H "Authorization: Bearer ${TOKEN2}" \
  "${BASE}/api/links/nope/stats")
N9_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N9_HTTP" == "404" && "$N9_CODE" == "link_not_found" ]] && record "N9|PASS|404|" || record "N9|FAIL|${N9_HTTP}|${N9_CODE}"

N10_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -H "Authorization: Bearer ${TOKEN2}" \
  "${BASE}/api/links/${NEG_LINK_CODE}/stats?limit=200")
N10_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N10_HTTP" == "400" && "$N10_CODE" == "invalid_limit" ]] && record "N10|PASS|400|" || record "N10|FAIL|${N10_HTTP}|${N10_CODE}"

N11_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -H "Authorization: Bearer ${TOKEN2}" \
  "${BASE}/api/links/${NEG_LINK_CODE}/qr?size=64")
N11_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N11_HTTP" == "400" && "$N11_CODE" == "invalid_size" ]] && record "N11|PASS|400|" || record "N11|FAIL|${N11_HTTP}|${N11_CODE}"

N12_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" -X PATCH \
  -H "Content-Type: application/json" -H "Authorization: Bearer ${TOKEN2}" \
  "${BASE}/api/links/${NEG_LINK_CODE}" -d '{}')
N12_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N12_HTTP" == "400" && "$N12_CODE" == "nothing_to_update" ]] && record "N12|PASS|400|" || record "N12|FAIL|${N12_HTTP}|${N12_CODE}"

N13_HTTP=$(curl -s -o /tmp/r.json -w "%{http_code}" "${BASE}/api/nonexistent")
N13_CODE=$(jq -r '.error.code // empty' /tmp/r.json)
[[ "$N13_HTTP" == "404" && "$N13_CODE" == "not_found" ]] && record "N13|PASS|404|" || record "N13|FAIL|${N13_HTTP}|${N13_CODE}"
