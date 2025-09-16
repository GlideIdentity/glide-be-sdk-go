#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Go SDK Integration Test${NC}"
echo "=========================="

# Check if API key is set
if [ -z "$GLIDE_API_KEY" ]; then
    echo -e "${RED}❌ GLIDE_API_KEY not set${NC}"
    echo "Please set GLIDE_API_KEY environment variable"
    exit 1
fi

# Build test server
echo -e "${YELLOW}Building test server...${NC}"
cd test-server
go mod download
go build -o test-server server.go
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to build test server${NC}"
    exit 1
fi

# Start test server
echo -e "${YELLOW}Starting test server...${NC}"
./test-server &
SERVER_PID=$!

# Wait for server to be ready
sleep 2

# Check if server is running
if ! ps -p $SERVER_PID > /dev/null; then
    echo -e "${RED}❌ Server failed to start${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Test server started (PID: $SERVER_PID)${NC}"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${YELLOW}Stopping test server...${NC}"
    kill $SERVER_PID 2>/dev/null
    wait $SERVER_PID 2>/dev/null
    echo -e "${GREEN}✅ Test server stopped${NC}"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Run integration tests
echo ""
echo -e "${YELLOW}Running integration tests...${NC}"
echo "============================="

# Test 1: Phone Auth Example
echo ""
echo -e "${YELLOW}Test 1: Phone Auth Example${NC}"
cd ../examples/phone-auth
go run main.go
TEST1_RESULT=$?

# Test 2: Complete Example
echo ""
echo -e "${YELLOW}Test 2: Complete Example${NC}"
cd ../complete
go run main.go
TEST2_RESULT=$?

# Test 3: Local Server Example
echo ""
echo -e "${YELLOW}Test 3: Local Server Example${NC}"
cd ../local-server
go run main.go
TEST3_RESULT=$?

# Summary
echo ""
echo -e "${GREEN}Test Results:${NC}"
echo "============="
if [ $TEST1_RESULT -eq 0 ]; then
    echo -e "✅ Phone Auth Example: ${GREEN}PASSED${NC}"
else
    echo -e "❌ Phone Auth Example: ${RED}FAILED${NC}"
fi

if [ $TEST2_RESULT -eq 0 ]; then
    echo -e "✅ Complete Example: ${GREEN}PASSED${NC}"
else
    echo -e "❌ Complete Example: ${RED}FAILED${NC}"
fi

if [ $TEST3_RESULT -eq 0 ]; then
    echo -e "✅ Local Server Example: ${GREEN}PASSED${NC}"
else
    echo -e "❌ Local Server Example: ${RED}FAILED${NC}"
fi

# Overall result
if [ $TEST1_RESULT -eq 0 ] && [ $TEST2_RESULT -eq 0 ] && [ $TEST3_RESULT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
