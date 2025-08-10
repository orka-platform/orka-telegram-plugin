#!/bin/bash

# Test script for the Orka LLM Plugin
# This script helps test the plugin functionality

set -e

PLUGIN_PORT=50052
PLUGIN_BINARY="./orka-llm-plugin"
TEST_CLIENT="./test_client"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.23+ first."
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_status "Go version: $GO_VERSION"
}

# Build the plugin
build_plugin() {
    print_status "Building LLM plugin..."
    
    if [ ! -f "go.mod" ]; then
        print_error "go.mod not found. Are you in the correct directory?"
        exit 1
    fi
    
    # Download dependencies
    print_status "Downloading dependencies..."
    go mod tidy
    go mod download
    
    # Build the plugin
    print_status "Building plugin binary..."
    go build -o "$PLUGIN_BINARY" main.go
    
    if [ -f "$PLUGIN_BINARY" ]; then
        print_success "Plugin built successfully: $PLUGIN_BINARY"
    else
        print_error "Failed to build plugin"
        exit 1
    fi
}

# Build test client
build_test_client() {
    print_status "Building test client..."
    go build -o "$TEST_CLIENT" test_client.go
    
    if [ -f "$TEST_CLIENT" ]; then
        print_success "Test client built successfully: $TEST_CLIENT"
    else
        print_error "Failed to build test client"
        exit 1
    fi
}

# Check if port is available
check_port() {
    if lsof -Pi :$PLUGIN_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_warning "Port $PLUGIN_PORT is already in use"
        return 1
    else
        print_success "Port $PLUGIN_PORT is available"
        return 0
    fi
}

# Start the plugin
start_plugin() {
    print_status "Starting LLM plugin on port $PLUGIN_PORT..."
    
    if [ ! -f "$PLUGIN_BINARY" ]; then
        print_error "Plugin binary not found. Run 'build' first."
        exit 1
    fi
    
    # Check if port is available
    if ! check_port; then
        print_warning "Please stop the process using port $PLUGIN_PORT and try again"
        exit 1
    fi
    
    # Start plugin in background
    "$PLUGIN_BINARY" --port "$PLUGIN_PORT" &
    PLUGIN_PID=$!
    
    # Wait a moment for plugin to start
    sleep 2
    
    # Check if plugin is running
    if kill -0 $PLUGIN_PID 2>/dev/null; then
        print_success "Plugin started successfully (PID: $PLUGIN_PID)"
        echo "$PLUGIN_PID" > .plugin.pid
    else
        print_error "Failed to start plugin"
        exit 1
    fi
}

# Stop the plugin
stop_plugin() {
    if [ -f .plugin.pid ]; then
        PLUGIN_PID=$(cat .plugin.pid)
        if kill -0 $PLUGIN_PID 2>/dev/null; then
            print_status "Stopping plugin (PID: $PLUGIN_PID)..."
            kill $PLUGIN_PID
            rm -f .plugin.pid
            print_success "Plugin stopped"
        else
            print_warning "Plugin process not found"
            rm -f .plugin.pid
        fi
    else
        print_warning "No plugin PID file found"
    fi
}

# Run tests
run_tests() {
    print_status "Running tests..."
    
    if [ ! -f "$TEST_CLIENT" ]; then
        print_error "Test client not found. Run 'build' first."
        exit 1
    fi
    
    # Check if plugin is running
    if ! lsof -Pi :$PLUGIN_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_error "Plugin is not running. Start it first with 'start'"
        exit 1
    fi
    
    print_status "Running test client..."
    "$TEST_CLIENT"
}

# Clean up
cleanup() {
    print_status "Cleaning up..."
    stop_plugin
    rm -f .plugin.pid
    print_success "Cleanup complete"
}

# Main function
main() {
    case "${1:-help}" in
        "build")
            check_go
            build_plugin
            build_test_client
            ;;
        "start")
            start_plugin
            ;;
        "stop")
            stop_plugin
            ;;
        "test")
            run_tests
            ;;
        "clean")
            cleanup
            rm -f "$PLUGIN_BINARY" "$TEST_CLIENT"
            print_success "All build artifacts removed"
            ;;
        "restart")
            stop_plugin
            sleep 1
            start_plugin
            ;;
        "status")
            if [ -f .plugin.pid ]; then
                PLUGIN_PID=$(cat .plugin.pid)
                if kill -0 $PLUGIN_PID 2>/dev/null; then
                    print_success "Plugin is running (PID: $PLUGIN_PID)"
                else
                    print_warning "Plugin PID file exists but process is not running"
                    rm -f .plugin.pid
                fi
            else
                print_warning "Plugin is not running"
            fi
            
            if lsof -Pi :$PLUGIN_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
                print_success "Port $PLUGIN_PORT is in use"
            else
                print_warning "Port $PLUGIN_PORT is not in use"
            fi
            ;;
        "help"|*)
            echo "Usage: $0 {build|start|stop|test|clean|restart|status|help}"
            echo ""
            echo "Commands:"
            echo "  build   - Build the plugin and test client"
            echo "  start   - Start the plugin on port $PLUGIN_PORT"
            echo "  stop    - Stop the running plugin"
            echo "  test    - Run the test client (plugin must be running)"
            echo "  clean   - Clean up all artifacts and stop plugin"
            echo "  restart - Restart the plugin"
            echo "  status  - Check plugin status"
            echo "  help    - Show this help message"
            echo ""
            echo "Example workflow:"
            echo "  $0 build    # Build everything"
            echo "  $0 start    # Start the plugin"
            echo "  $0 test     # Run tests"
            echo "  $0 stop     # Stop the plugin"
            echo "  $0 clean    # Clean up everything"
            ;;
    esac
}

# Trap cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"