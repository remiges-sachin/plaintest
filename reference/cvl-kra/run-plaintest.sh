#!/bin/bash

# PlainTest runner for PAN validation collections

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"
PLAINTEST_BIN=${PLAINTEST_BIN:-"$REPO_ROOT/plaintest"}

TEST_TYPE=""
ENVIRONMENT=""
ROW_SELECTION=""
ITERATION_COUNT=""
DEBUG_MODE=false
RUN_SYNC=true

usage() {
    cat <<USAGE
Usage: $0 -t <smoke|full> -e <localhost|uat> [options]

Required:
  -t    Test type: smoke | full
  -e    Environment: localhost | uat

Optional:
  -r    Row selection (PlainTest -r syntax, e.g. 2, 2-5, 1,3,5)
  -n    Iteration count (maps to Newman --iteration-count)
  -d    Debug mode (passes --verbose to Newman)
  -S    Skip script sync step
  -h    Show this help text

Environment variables:
  PLAINTEST_BIN  Path to plaintest binary (default: ../../plaintest)
USAGE
}

info()  { echo -e "${YELLOW}[i]${NC} $1"; }
pass()  { echo -e "${GREEN}[✓]${NC} $1"; }
fail()  { echo -e "${RED}[✗]${NC} $1"; }
step()  { echo -e "${BLUE}[→]${NC} $1"; }

while getopts ':t:e:r:n:dSh' opt; do
    case "$opt" in
        t) TEST_TYPE="$OPTARG" ;;
        e) ENVIRONMENT="$OPTARG" ;;
        r) ROW_SELECTION="$OPTARG" ;;
        n) ITERATION_COUNT="$OPTARG" ;;
        d) DEBUG_MODE=true ;;
        S) RUN_SYNC=false ;;
        h) usage; exit 0 ;;
        :) fail "Option -$OPTARG requires an argument"; usage; exit 1 ;;
        ?) fail "Invalid option: -$OPTARG"; usage; exit 1 ;;
    esac
done

if [[ -z "$TEST_TYPE" || -z "$ENVIRONMENT" ]]; then
    usage
    exit 1
fi

if [[ "$TEST_TYPE" != "smoke" && "$TEST_TYPE" != "full" ]]; then
    fail "Unknown test type: $TEST_TYPE"
    exit 1
fi

if [[ "$ENVIRONMENT" != "localhost" && "$ENVIRONMENT" != "uat" ]]; then
    fail "Unknown environment: $ENVIRONMENT"
    exit 1
fi

if [[ -n "$ITERATION_COUNT" && ! "$ITERATION_COUNT" =~ ^[0-9]+$ ]]; then
    fail "Iteration count must be numeric"
    exit 1
fi

if [[ ! -x "$PLAINTEST_BIN" ]]; then
    fail "PlainTest binary not found at: $PLAINTEST_BIN"
    exit 1
fi

if ! command -v newman >/dev/null 2>&1; then
    fail "Newman is required. Install with: npm install -g newman newman-reporter-htmlextra"
    exit 1
fi

cd "$SCRIPT_DIR"

if $RUN_SYNC; then
    step "Syncing Postman scripts"
    "$PLAINTEST_BIN" scripts sync
    pass "Postman scripts synced"
fi

ENV_FILE="environments/${ENVIRONMENT}.postman_environment.json"
if [[ ! -f "$ENV_FILE" ]]; then
    fail "Environment file missing: $ENV_FILE"
    exit 1
fi

mkdir -p reports

if [[ "$ENVIRONMENT" == "localhost" ]]; then
    if ! nc -z localhost 8083 >/dev/null 2>&1; then
        if [[ -x "ops/start-app.sh" ]]; then
            step "Starting local services via ops/start-app.sh"
            ./ops/start-app.sh
            pass "Local services started"
        else
            info "Local services not detected and no ops/start-app.sh available"
        fi
    fi
fi

cmd=("$PLAINTEST_BIN" run)

if [[ "$TEST_TYPE" == "smoke" ]]; then
    step "Running smoke collection"
    cmd+=("pan_validation_smoke" "-e" "$ENV_FILE")
    REPORT_HTML="reports/smoke-test-results-${ENVIRONMENT}.html"
    REPORT_JSON="reports/smoke-test-results-${ENVIRONMENT}.json"
    cmd+=("--reporters" "cli,htmlextra,json" "--reporter-htmlextra-export" "$REPORT_HTML" "--reporter-json-export" "$REPORT_JSON" "--color" "on" "--delay-request" "500")
else
    step "Running full CSV-driven suite"
    DATA_FILE="data/pan_validation_plaintest.csv"
    if [[ ! -f "$DATA_FILE" ]]; then
        fail "CSV data file missing: $DATA_FILE"
        exit 1
    fi
    cmd+=("get_password" "pan_validation_tests" "-e" "$ENV_FILE" "-d" "$DATA_FILE")
    if [[ -n "$ROW_SELECTION" ]]; then
        cmd+=("-r" "$ROW_SELECTION")
    fi
    if [[ -n "$ITERATION_COUNT" ]]; then
        cmd+=("--iteration-count" "$ITERATION_COUNT")
    fi
    REPORT_HTML="reports/pan-validation-test-results-${ENVIRONMENT}.html"
    REPORT_JSON="reports/pan-validation-test-results-${ENVIRONMENT}.json"
    cmd+=("--reporters" "cli,htmlextra,json" "--reporter-htmlextra-export" "$REPORT_HTML" "--reporter-json-export" "$REPORT_JSON" "--color" "on" "--delay-request" "1000" "--timeout-request" "10000" "--ignore-redirects")
fi

if $DEBUG_MODE; then
    cmd+=("--verbose")
fi

info "Executing: ${cmd[*]}"
"${cmd[@]}"
pass "PlainTest execution finished"

if [[ -f "$REPORT_HTML" ]]; then
    pass "HTML report: $REPORT_HTML"
fi
if [[ -f "$REPORT_JSON" ]]; then
    pass "JSON report: $REPORT_JSON"
fi
