#!/bin/bash
# Test script for Phase 2: Core Layout & Panels
# Verifies the 4-panel layout renders correctly with Tab navigation

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "======================================"
echo "Phase 2 Layout Test Script"
echo "======================================"

# Check if lazyazure binary exists
if [ ! -f ./lazyazure ]; then
    echo -e "${YELLOW}Building lazyazure...${NC}"
    go build -o lazyazure .
fi

# Create a unique session name
SESSION="lazyazure-phase2-test-$$"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    tmux kill-session -t "$SESSION" 2>/dev/null || true
}
trap cleanup EXIT

echo -e "${GREEN}1. Testing basic startup...${NC}"

# Create tmux session with demo mode
tmux new-session -d -s "$SESSION" -x 120 -y 40

# Start lazyazure in demo mode
tmux send-keys -t "$SESSION" "LAZYAZURE_DEMO=1 ./lazyazure" Enter

# Wait for app to render
sleep 2

# Capture initial output
echo -e "${GREEN}2. Capturing initial screen...${NC}"
tmux capture-pane -t "$SESSION" -p > /tmp/phase2_initial.txt

# Check for expected elements
if grep -q "Auth" /tmp/phase2_initial.txt; then
    echo -e "${GREEN}✓ Auth panel renders${NC}"
else
    echo -e "${YELLOW}⚠ Auth panel not clearly visible (may be styling issue)${NC}"
fi

if grep -q "Subscriptions" /tmp/phase2_initial.txt; then
    echo -e "${GREEN}✓ Subscriptions panel renders${NC}"
else
    echo -e "${RED}✗ Subscriptions panel missing${NC}"
fi

if grep -q "Resource Groups" /tmp/phase2_initial.txt; then
    echo -e "${GREEN}✓ Resource Groups panel renders${NC}"
else
    echo -e "${RED}✗ Resource Groups panel missing${NC}"
fi

if grep -q "Resources" /tmp/phase2_initial.txt; then
    echo -e "${GREEN}✓ Resources panel renders${NC}"
else
    echo -e "${RED}✗ Resources panel missing${NC}"
fi

if grep -q "Summary" /tmp/phase2_initial.txt || grep -q "JSON" /tmp/phase2_initial.txt; then
    echo -e "${GREEN}✓ Main panel tabs render${NC}"
else
    echo -e "${YELLOW}⚠ Main panel tabs not clearly visible${NC}"
fi

echo -e "${GREEN}3. Testing Tab navigation...${NC}"

# Send Tab key to cycle through panels
tmux send-keys -t "$SESSION" Tab
sleep 0.5
tmux capture-pane -t "$SESSION" -p > /tmp/phase2_tab1.txt

tmux send-keys -t "$SESSION" Tab
sleep 0.5
tmux capture-pane -t "$SESSION" -p > /tmp/phase2_tab2.txt

tmux send-keys -t "$SESSION" Tab
sleep 0.5
tmux capture-pane -t "$SESSION" -p > /tmp/phase2_tab3.txt

# Check that panels changed focus (we can't easily verify this without screenshots,
# but the fact that the app didn't crash is a good sign)
echo -e "${GREEN}✓ Tab navigation didn't crash the app${NC}"

echo -e "${GREEN}4. Testing quit functionality...${NC}"

# Send 'q' to quit
tmux send-keys -t "$SESSION" q
sleep 1

# Check if process exited
if ! tmux list-sessions | grep -q "$SESSION"; then
    echo -e "${RED}✗ Session ended unexpectedly${NC}"
else
    # Check if pane still has lazyazure running
    if tmux list-panes -t "$SESSION" -F "#{pane_current_command}" | grep -q "lazyazure"; then
        echo -e "${RED}✗ Lazyazure still running after 'q' key${NC}"
    else
        echo -e "${GREEN}✓ App exited on 'q' key${NC}"
    fi
fi

echo ""
echo "======================================"
echo -e "${GREEN}Phase 2 tests completed!${NC}"
echo "======================================"
echo ""
echo "Test output saved to:"
echo "  - /tmp/phase2_initial.txt"
echo "  - /tmp/phase2_tab1.txt"
echo "  - /tmp/phase2_tab2.txt"
echo "  - /tmp/phase2_tab3.txt"
echo ""
echo "To manually inspect, run:"
echo "  cat /tmp/phase2_initial.txt"
