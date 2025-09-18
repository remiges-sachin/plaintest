#!/bin/bash

# CVL KRA Backend Stop Script
# Stops the CVL KRA backend application and optionally its dependencies

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[i]${NC} $1"
}

# Main script
echo "===================================="
echo "CVL KRA Backend Stop Script"
echo "===================================="

# Stop the application
print_info "Stopping CVL KRA Backend application..."
pkill -f cvl-kra-backend || true
print_status "Application stopped"

# Ask if user wants to stop Docker services
echo ""
read -p "Do you want to stop Docker services as well? (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_info "Stopping Docker services..."
    docker-compose down
    print_status "Docker services stopped"

    # Ask if user wants to remove volumes
    echo ""
    read -p "Do you want to remove data volumes? WARNING: This will delete all data! (y/n): " -n 1 -r
    echo ""

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Removing Docker volumes..."
        docker-compose down -v
        print_status "Docker volumes removed"
    fi
else
    print_info "Docker services are still running"
    print_info "To stop them later, run: docker-compose down"
fi

echo ""
echo "===================================="
echo "Shutdown complete"
echo "===================================="
