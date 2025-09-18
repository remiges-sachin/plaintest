#!/bin/bash

# CVL KRA Backend Startup Script
# Starts the CVL KRA backend application with required dependencies

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[i]${NC} $1"
}

# Checks if a command exists in PATH
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Waits for service to accept connections on specified port
wait_for_service() {
    local host=$1
    local port=$2
    local service=$3
    local max_attempts=30
    local attempt=1

    print_info "Waiting for $service to be ready..."

    while ! nc -z $host $port >/dev/null 2>&1; do
        if [ $attempt -eq $max_attempts ]; then
            print_error "$service failed to start on $host:$port"
            return 1
        fi
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    echo ""
    print_status "$service is ready on $host:$port"
}

# Main script
echo "===================================="
echo "CVL KRA Backend Startup Script"
echo "===================================="

# Get the script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

# Change to project root directory
cd "$PROJECT_ROOT"
print_info "Working directory: $PROJECT_ROOT"

# Source .env file for database configuration
if [ -f .env ]; then
    set -a  # automatically export all variables
    source .env
    set +a
    print_status "Loaded configuration from .env file"
else
    print_error ".env file not found!"
    exit 1
fi

# Check prerequisites
print_info "Checking prerequisites..."

if ! command_exists docker; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command_exists docker-compose; then
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

if ! command_exists go; then
    print_error "Go is not installed. Please install Go 1.23.3 or higher."
    exit 1
fi

if ! command_exists make; then
    print_error "Make is not installed. Please install Make."
    exit 1
fi

print_status "All prerequisites are installed"

# Remove any existing containers with conflicting names
print_info "Cleaning up any existing containers..."
docker-compose down >/dev/null 2>&1 || true
docker rm -f mssql-service postgres-service redis-service etcd-service minio_service cvl-kra_mssql cvl-kra_postgres cvl-kra_redis cvl-kra_etcd cvl-kra_minio >/dev/null 2>&1 || true

# Start Docker services
print_info "Starting Docker services with project name: cvl-kra"
export COMPOSE_PROJECT_NAME=cvl-kra
docker-compose up -d

# Wait for services to be ready
print_info "Waiting for services to start..."
sleep 10  # Initial wait for containers to start

# Check each service
wait_for_service localhost 1433 "SQL Server"
wait_for_service localhost 5433 "PostgreSQL"
wait_for_service localhost 6378 "Redis"
wait_for_service localhost 9000 "MinIO"
wait_for_service localhost 2379 "etcd"
wait_for_service localhost 8080 "Keycloak"

# Additional wait time for SQL Server initialization
print_info "Waiting for SQL Server to be fully initialized..."
sleep 10

# Create required databases
print_info "Ensuring databases exist..."

# Create SQL Server databases
print_info "Creating SQL Server databases if they don't exist..."

# Wait for SQL Server to accept commands
print_info "Waiting for SQL Server to accept commands (this may take up to 90 seconds)..."
max_attempts=45
attempt=1

# Locate sqlcmd executable path
SQLCMD_PATH="/opt/mssql-tools/bin/sqlcmd"
if ! docker exec cvl-kra_mssql test -f "$SQLCMD_PATH" 2>/dev/null; then
    SQLCMD_PATH="/opt/mssql-tools18/bin/sqlcmd"
fi

while ! docker exec cvl-kra_mssql $SQLCMD_PATH -S localhost -U $db_user -P $db_password -Q "SELECT 1" -C &> /dev/null; do
    if [ $attempt -eq $max_attempts ]; then
        print_error "SQL Server failed to accept commands after 90 seconds"
        print_info "Checking SQL Server logs..."
        docker logs cvl-kra_mssql --tail 20
        exit 1
    fi
    echo -n "."
    sleep 2
    attempt=$((attempt + 1))
done
echo ""
print_status "SQL Server is ready to accept commands"

# Create main application database
if docker exec cvl-kra_mssql $SQLCMD_PATH -S localhost -U $db_user -P $db_password -Q "IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = '$db_name') CREATE DATABASE $db_name" -C; then
    print_status "SQL Server database '$db_name' created/verified"
else
    print_error "Failed to create SQL Server database '$db_name'"
fi

# Create Keycloak database
if docker exec cvl-kra_mssql $SQLCMD_PATH -S localhost -U $db_user -P $db_password -Q "IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'cvl_kra_app') CREATE DATABASE cvl_kra_app" -C; then
    print_status "SQL Server database 'cvl_kra_app' created/verified"
else
    print_error "Failed to create Keycloak database 'cvl_kra_app'"
fi

# PostgreSQL database is created automatically by POSTGRES_DB env var in docker-compose
print_status "PostgreSQL database 'cvl_app' created automatically via POSTGRES_DB env var"

print_status "Database creation completed"

# Run database migrations
print_info "Running database migrations..."

# Note: The makefile includes config/dbconfig.env automatically

# SQL Server migrations
print_info "Running SQL Server migrations..."
if make goose-schema-up 2>&1 | grep -q "no change"; then
    print_info "Schema already up to date"
elif make goose-schema-up; then
    print_status "Schema migration successful"
else
    print_error "Schema migration failed"
fi

if make goose-seed-up 2>&1 | grep -q "no change"; then
    print_info "Seed data already up to date"
elif make goose-seed-up; then
    print_status "Seed data migration successful"
else
    print_error "Seed data migration failed"
fi

if make goose-dataaddon-up 2>&1 | grep -q "no change"; then
    print_info "Data addon already up to date"
elif make goose-dataaddon-up; then
    print_status "Data addon migration successful"
else
    print_error "Data addon migration failed"
fi
print_status "SQL Server migrations completed"

# PostgreSQL migrations
print_info "Running PostgreSQL migrations..."
if make goose-pg-schema-up 2>&1 | grep -q "no change"; then
    print_info "PG Schema already up to date"
elif make goose-pg-schema-up; then
    print_status "PG Schema migration successful"
else
    print_error "PG Schema migration failed"
fi

if make goose-pg-seed-up 2>&1 | grep -q "no change"; then
    print_info "PG Seed data already up to date"
elif make goose-pg-seed-up; then
    print_status "PG Seed data migration successful"
else
    print_error "PG Seed data migration failed"
fi

if make goose-pg-dataaddon-up 2>&1 | grep -q "no change"; then
    print_info "PG Data addon already up to date"
elif make goose-pg-dataaddon-up; then
    print_status "PG Data addon migration successful"
else
    print_error "PG Data addon migration failed"
fi

if make sqlc-generate; then
    print_status "SQLC generation successful"
else
    print_error "SQLC generation failed"
fi
print_status "PostgreSQL migrations completed"

# Setup Rigel configuration
print_info "Setting up Rigel configuration..."
if make setup-rigel; then
    print_status "Rigel configuration initialized"
else
    print_error "Rigel setup failed"
fi

# Build the application
print_info "Building the application..."
go build -o cvl-kra-backend .
print_status "Application built successfully"

# Run the application
print_info "Starting CVL KRA Backend on port 8083..."
echo ""
echo "===================================="
echo "Application is starting..."
echo "URL: http://localhost:8083"
echo "Press Ctrl+C to stop"
echo "===================================="
echo ""

./cvl-kra-backend
