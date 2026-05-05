#!/bin/bash

# Hades-V2 Production Deployment Script
# This script automates the deployment of Hades-V2 in production environments

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DEPLOY_ENV="${DEPLOY_ENV:-production}"
BACKUP_DIR="${PROJECT_DIR}/backups"
LOG_FILE="${PROJECT_DIR}/logs/deploy.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
    exit 1
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if running as root for system-wide installation
    if [[ "$DEPLOY_ENV" == "system" ]] && [[ $EUID -ne 0 ]]; then
        error "System-wide deployment requires root privileges. Use sudo or run as root."
    fi
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed. Please install Docker first."
    fi
    
    # Check if Docker Compose is installed
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        error "Docker Compose is not installed. Please install Docker Compose first."
    fi
    
    # Check if required ports are available
    local ports=(8443 8080 5432 6379)
    for port in "${ports[@]}"; do
        if lsof -i ":$port" &> /dev/null; then
            warning "Port $port is already in use. This may cause conflicts."
        fi
    done
    
    # Create necessary directories
    mkdir -p "$BACKUP_DIR"
    mkdir -p "$PROJECT_DIR/logs"
    mkdir -p "$PROJECT_DIR/data"
    mkdir -p "$PROJECT_DIR/config"
    
    success "Prerequisites check completed"
}

# Backup existing data
backup_data() {
    log "Creating backup of existing data..."
    
    local backup_name="hades-backup-$(date +%Y%m%d-%H%M%S)"
    local backup_path="$BACKUP_DIR/$backup_name"
    
    mkdir -p "$backup_path"
    
    # Backup database if exists
    if [[ -f "$PROJECT_DIR/data/hades.db" ]]; then
        cp "$PROJECT_DIR/data/hades.db" "$backup_path/"
        log "Database backed up to $backup_path/hades.db"
    fi
    
    # Backup configuration
    if [[ -f "$PROJECT_DIR/config/hades.yaml" ]]; then
        cp "$PROJECT_DIR/config/hades.yaml" "$backup_path/"
        log "Configuration backed up to $backup_path/hades.yaml"
    fi
    
    # Backup logs
    if [[ -d "$PROJECT_DIR/logs" ]]; then
        cp -r "$PROJECT_DIR/logs" "$backup_path/"
        log "Logs backed up to $backup_path/logs/"
    fi
    
    success "Backup created: $backup_path"
}

# Build application
build_application() {
    log "Building Hades-V2 application..."
    
    cd "$PROJECT_DIR"
    
    # Clean previous builds
    go clean -cache
    
    # Run tests
    log "Running tests..."
    go test -v ./... || {
        error "Tests failed. Aborting deployment."
    }
    
    # Build binary
    log "Building binary..."
    go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o hades ./cmd/hades || {
        error "Build failed."
    }
    
    # Build Docker image
    log "Building Docker image..."
    docker build -t hades-v2:latest . || {
        error "Docker build failed."
    }
    
    success "Application built successfully"
}

# Deploy using Docker Compose
deploy_docker() {
    log "Deploying using Docker Compose..."
    
    cd "$PROJECT_DIR"
    
    # Stop existing services
    log "Stopping existing services..."
    docker-compose down || true
    
    # Pull latest images
    log "Pulling latest images..."
    docker-compose pull
    
    # Start services
    log "Starting services..."
    docker-compose up -d
    
    # Wait for services to be ready
    log "Waiting for services to be ready..."
    sleep 30
    
    # Check service health
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -f http://localhost:8443/api/health &> /dev/null; then
            success "Services are ready and healthy"
            return 0
        fi
        
        log "Waiting for services... (attempt $attempt/$max_attempts)"
        sleep 10
        ((attempt++))
    done
    
    error "Services failed to become ready within timeout"
}

# Deploy as system service
deploy_system() {
    log "Deploying as system service..."
    
    cd "$PROJECT_DIR"
    
    # Create hades user if not exists
    if ! id "hades" &>/dev/null; then
        useradd -r -s /bin/false hades
        log "Created hades user"
    fi
    
    # Install binary
    cp hades /usr/local/bin/
    chmod +x /usr/local/bin/hades
    
    # Create directories
    mkdir -p /opt/hades/{data,logs,config}
    chown -R hades:hades /opt/hades
    
    # Copy configuration
    if [[ -f "config/hades.yaml" ]]; then
        cp config/hades.yaml /opt/hades/config/
    fi
    
    # Create systemd service
    cat > /etc/systemd/system/hades.service << EOF
[Unit]
Description=Hades-V2 Enterprise Security Framework
After=network.target

[Service]
Type=simple
User=hades
Group=hades
WorkingDirectory=/opt/hades
ExecStart=/usr/local/bin/hades web start --port 8443 --config /opt/hades/config/hades.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    
    # Reload systemd and start service
    systemctl daemon-reload
    systemctl enable hades
    systemctl start hades
    
    # Check service status
    if systemctl is-active --quiet hades; then
        success "Hades service started successfully"
    else
        error "Hades service failed to start"
    fi
}

# Initialize database
initialize_database() {
    log "Initializing database..."
    
    cd "$PROJECT_DIR"
    
    # Run migrations
    ./hades migrate init || {
        error "Database initialization failed"
    }
    
    ./hades migrate up || {
        error "Database migration failed"
    }
    
    success "Database initialized successfully"
}

# Create default admin user
create_admin_user() {
    log "Creating default admin user..."
    
    cd "$PROJECT_DIR"
    
    # Check if admin user already exists
    if ./hades user list | grep -q "admin"; then
        log "Admin user already exists"
        return 0
    fi
    
    # Create admin user
    local admin_password="${HADES_ADMIN_PASSWORD:-$(openssl rand -base64 32)}"
    
    ./hades user create \
        --username admin \
        --email admin@localhost \
        --role admin \
        --password "$admin_password" || {
        error "Failed to create admin user"
    }
    
    success "Admin user created successfully"
    log "Admin password: $admin_password"
    log "Save this password securely!"
}

# Run health checks
run_health_checks() {
    log "Running health checks..."
    
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        # Check web interface
        if curl -f http://localhost:8443/api/health &> /dev/null; then
            success "Web interface is healthy"
        else
            log "Web interface not ready yet... (attempt $attempt/$max_attempts)"
        fi
        
        # Check API server
        if curl -f http://localhost:8080/api/health &> /dev/null; then
            success "API server is healthy"
        else
            log "API server not ready yet... (attempt $attempt/$max_attempts)"
        fi
        
        # Check database
        if ./hades migrate status &> /dev/null; then
            success "Database is accessible"
        else
            log "Database not accessible yet... (attempt $attempt/$max_attempts)"
        fi
        
        sleep 5
        ((attempt++))
    done
    
    success "All health checks completed"
}

# Show deployment summary
show_summary() {
    log "Deployment summary:"
    echo "=================="
    echo "Environment: $DEPLOY_ENV"
    echo "Web Interface: http://localhost:8443"
    echo "API Server: http://localhost:8080"
    echo "Admin User: admin"
    echo "Configuration: $PROJECT_DIR/config/hades.yaml"
    echo "Logs: $PROJECT_DIR/logs/"
    echo "Data: $PROJECT_DIR/data/"
    echo "=================="
    
    if [[ "$DEPLOY_ENV" == "docker" ]]; then
        echo "Docker Commands:"
        echo "  docker-compose logs -f    # View logs"
        echo "  docker-compose down       # Stop services"
        echo "  docker-compose up -d      # Start services"
    else
        echo "Service Commands:"
        echo "  sudo systemctl status hades    # Check status"
        echo "  sudo systemctl restart hades   # Restart service"
        echo "  sudo journalctl -u hades -f     # View logs"
    fi
    
    success "Deployment completed successfully!"
}

# Main deployment function
main() {
    log "Starting Hades-V2 deployment..."
    log "Environment: $DEPLOY_ENV"
    log "Project directory: $PROJECT_DIR"
    
    # Check prerequisites
    check_prerequisites
    
    # Backup existing data
    backup_data
    
    # Build application
    build_application
    
    # Deploy based on environment
    case "$DEPLOY_ENV" in
        "docker")
            deploy_docker
            ;;
        "system")
            deploy_system
            ;;
        *)
            error "Unknown deployment environment: $DEPLOY_ENV. Use 'docker' or 'system'."
            ;;
    esac
    
    # Initialize database
    initialize_database
    
    # Create admin user
    create_admin_user
    
    # Run health checks
    run_health_checks
    
    # Show summary
    show_summary
}

# Script usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -e, --env ENV     Deployment environment (docker|system) [default: docker]"
    echo "  -h, --help       Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  DEPLOY_ENV       Deployment environment"
    echo "  HADES_ADMIN_PASSWORD  Admin user password (auto-generated if not set)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Deploy with Docker (default)"
    echo "  $0 -e docker         # Deploy with Docker"
    echo "  $0 -e system         # Deploy as system service"
    echo "  HADES_ADMIN_PASSWORD=secret $0  # Deploy with custom admin password"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            DEPLOY_ENV="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $1. Use -h for help."
            ;;
    esac
done

# Set default environment
if [[ -z "${DEPLOY_ENV:-}" ]]; then
    DEPLOY_ENV="docker"
fi

# Validate environment
if [[ "$DEPLOY_ENV" != "docker" && "$DEPLOY_ENV" != "system" ]]; then
    error "Invalid environment: $DEPLOY_ENV. Use 'docker' or 'system'."
fi

# Run main function
main
