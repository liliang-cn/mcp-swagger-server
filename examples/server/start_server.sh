#!/bin/bash

echo "🐾 Starting Local Petstore API Server..."
echo "========================================"

# Check if we're in the server directory
if [ ! -f "main.go" ]; then
    echo "Error: main.go not found. Please run this script from the examples/server directory."
    exit 1
fi

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go to run this server."
    exit 1
fi

# Check if gorilla/mux is available
if ! go list github.com/gorilla/mux &> /dev/null; then
    echo "Installing gorilla/mux dependency..."
    go get github.com/gorilla/mux
fi

echo "Starting server on port 8080..."
echo "Press Ctrl+C to stop the server"
echo ""
echo "Once started, you can test with:"
echo "  curl http://localhost:8080/health"
echo "  curl http://localhost:8080/v2/pets"
echo ""

# Start the server
go run main.go