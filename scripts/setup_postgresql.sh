#!/bin/bash

# PostgreSQL Setup Script with Passwordless Sudo
# This script sets up PostgreSQL with passwordless sudo access for Hades

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
}

success() {
    echo -e "${GREEN}[SUCCESS] $1${NC}"
}

warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        error "This script should not be run as root. Run it as a regular user with sudo privileges."
        exit 1
    fi
}

# Setup passwordless sudo for PostgreSQL commands
setup_passwordless_sudo() {
    log "Setting up passwordless sudo for PostgreSQL commands..."
    
    # Create sudoers file for PostgreSQL
    SUDOERS_FILE="/etc/sudoers.d/hades-postgresql"
    CURRENT_USER=$(whoami)
    
    # Create temporary sudoers file
    TEMP_SUDOERS=$(mktemp)
    cat > "$TEMP_SUDOERS" << EOF
# Hades PostgreSQL passwordless sudo configuration
$CURRENT_USER ALL=(postgres) NOPASSWD: /usr/bin/psql, /usr/bin/createuser, /usr/bin/dropuser, /usr/bin/createdb, /usr/bin/dropdb
$CURRENT_USER ALL=(root) NOPASSWD: /bin/systemctl restart postgresql, /bin/systemctl start postgresql, /bin/systemctl stop postgresql, /bin/systemctl status postgresql
EOF
    
    # Validate and install sudoers file
    if sudo visudo -cf "$TEMP_SUDOERS" 2>/dev/null; then
        sudo cp "$TEMP_SUDOERS" "$SUDOERS_FILE"
        sudo chmod 440 "$SUDOERS_FILE"
        success "Passwordless sudo configured for PostgreSQL commands"
    else
        error "Failed to validate sudoers configuration"
        rm "$TEMP_SUDOERS"
        exit 1
    fi
    
    rm "$TEMP_SUDOERS"
}

# Install PostgreSQL
install_postgresql() {
    log "Installing PostgreSQL..."
    
    # Update package list
    sudo apt update
    
    # Install PostgreSQL and contrib
    sudo apt install -y postgresql postgresql-contrib postgresql-client
    
    # Enable and start PostgreSQL service
    sudo systemctl enable postgresql
    sudo systemctl start postgresql
    
    success "PostgreSQL installed and started"
}

# Create Hades database and user
create_database() {
    log "Creating Hades database and user..."
    
    # Create database
    sudo -u postgres createdb hades_toolkit 2>/dev/null || {
        warning "Database hades_toolkit might already exist"
    }
    
    # Create user with password
    sudo -u postgres psql -c "DROP USER IF EXISTS hades;" 2>/dev/null || true
    sudo -u postgres psql -c "CREATE USER hades WITH PASSWORD 'hades_password';" 2>/dev/null
    
    # Grant privileges
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE hades_toolkit TO hades;" 2>/dev/null
    sudo -u postgres psql -c "ALTER USER hades CREATEDB;" 2>/dev/null
    
    success "Hades database and user created"
}

# Create replica databases for distributed setup
create_replicas() {
    log "Creating replica databases for distributed setup..."
    
    # Create replica databases
    for i in 1 2; do
        REPLICA_DB="hades_toolkit_replica${i}"
        sudo -u postgres createdb "$REPLICA_DB" 2>/dev/null || {
            warning "Replica database $REPLICA_DB might already exist"
        }
        sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $REPLICA_DB TO hades;" 2>/dev/null
    done
    
    success "Replica databases created"
}

# Setup PostgreSQL configuration for distributed access
configure_postgresql() {
    log "Configuring PostgreSQL for distributed access..."
    
    PG_CONF="/etc/postgresql/*/main/postgresql.conf"
    PG_HBA="/etc/postgresql/*/main/pg_hba.conf"
    
    # Backup original files
    sudo cp $PG_CONF "${PG_CONF}.backup"
    sudo cp $PG_HBA "${PG_HBA}.backup"
    
    # Configure for multiple ports (primary: 5432, replicas: 5433, 5434)
    sudo sed -i "s/#port = 5432/port = 5432/" $PG_CONF
    
    # Configure pg_hba.conf for local connections
    sudo bash -c 'cat >> /etc/postgresql/*/main/pg_hba.conf << EOF
# Hades distributed database configuration
local   all             all                                     trust
host    all             all             127.0.0.1/32            trust
host    all             all             ::1/128                 trust
EOF'
    
    # Restart PostgreSQL
    sudo systemctl restart postgresql
    
    success "PostgreSQL configured for distributed access"
}

# Setup additional PostgreSQL instances for replicas
setup_replica_instances() {
    log "Setting up additional PostgreSQL instances for replicas..."
    
    # Create data directories for replicas
    sudo mkdir -p /var/lib/postgresql/replica1
    sudo mkdir -p /var/lib/postgresql/replica2
    sudo chown -R postgres:postgres /var/lib/postgresql/replica1
    sudo chown -R postgres:postgres /var/lib/postgresql/replica2
    
    # Initialize replica databases
    sudo -u postgres /usr/lib/postgresql/*/bin/initdb -D /var/lib/postgresql/replica1 2>/dev/null || {
        warning "Replica 1 might already be initialized"
    }
    sudo -u postgres /usr/lib/postgresql/*/bin/initdb -D /var/lib/postgresql/replica2 2>/dev/null || {
        warning "Replica 2 might already be initialized"
    }
    
    # Create replica configuration files
    create_replica_config 1 5433
    create_replica_config 2 5434
    
    success "Replica instances configured"
}

# Create configuration for a replica instance
create_replica_config() {
    local replica_num=$1
    local port=$2
    local data_dir="/var/lib/postgresql/replica${replica_num}"
    local config_file="${data_dir}/postgresql.conf"
    
    sudo bash -c "cat > '$config_file' << EOF
# PostgreSQL replica $replica_num configuration
port = $port
max_connections = 100
shared_buffers = 128MB
effective_cache_size = 4GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 4MB
min_wal_size = 1GB
max_wal_size = 4GB
EOF"
    
    # Create systemd service for replica
    sudo bash -c "cat > /etc/systemd/system/postgresql-replica${replica_num}.service << EOF
[Unit]
Description=PostgreSQL database server replica $replica_num
After=network.target

[Service]
Type=notify
User=postgres
Group=postgres
ExecStart=/usr/lib/postgresql/*/bin/postgres -D $data_dir
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGINT
TimeoutSec=0

[Install]
WantedBy=multi-user.target
EOF"
    
    # Enable and start replica service
    sudo systemctl daemon-reload
    sudo systemctl enable "postgresql-replica${replica_num}"
    sudo systemctl start "postgresql-replica${replica_num}" || {
        warning "Failed to start replica $replica_num, will try manual start"
    }
}

# Test connections
test_connections() {
    log "Testing database connections..."
    
    # Test primary connection
    if psql -h localhost -p 5432 -U hades -d hades_toolkit -c "SELECT 1;" >/dev/null 2>&1; then
        success "Primary database connection successful"
    else
        error "Primary database connection failed"
        return 1
    fi
    
    # Test replica connections
    for i in 1 2; do
        port=$((5432 + i))
        if psql -h localhost -p $port -U hades -d hades_toolkit_replica${i} -c "SELECT 1;" >/dev/null 2>&1; then
            success "Replica $i connection successful (port $port)"
        else
            warning "Replica $i connection failed (port $port), this is expected if replica services aren't running"
        fi
    done
}

# Create hot-swap configuration files
create_hotswap_configs() {
    log "Creating hot-swap configuration files..."
    
    CONFIG_DIR="/home/$(whoami)/.hades/database"
    mkdir -p "$CONFIG_DIR"
    
    # Development configuration
    cat > "$CONFIG_DIR/dev.json" << EOF
{
  "type": "postgresql",
  "primary": {
    "id": "primary-1",
    "host": "localhost",
    "port": 5432,
    "database": "hades_toolkit",
    "username": "hades",
    "password": "hades_password",
    "ssl_mode": "disable",
    "weight": 100,
    "healthy": true
  },
  "replicas": [
    {
      "id": "replica-1",
      "host": "localhost",
      "port": 5433,
      "database": "hades_toolkit_replica1",
      "username": "hades",
      "password": "hades_password",
      "ssl_mode": "disable",
      "weight": 80,
      "healthy": true
    },
    {
      "id": "replica-2",
      "host": "localhost",
      "port": 5434,
      "database": "hades_toolkit_replica2",
      "username": "hades",
      "password": "hades_password",
      "ssl_mode": "disable",
      "weight": 60,
      "healthy": true
    }
  ],
  "failover_enabled": true,
  "health_check_interval": "30s",
  "max_retries": 3,
  "connection_timeout": "10s",
  "load_balancing": "least_connections"
}
EOF

    # SQLite fallback configuration
    cat > "$CONFIG_DIR/sqlite_fallback.json" << EOF
{
  "type": "sqlite3",
  "primary": {
    "id": "sqlite-primary",
    "host": "",
    "port": 0,
    "database": "/home/$(whoami)/.hades/hades.db",
    "username": "",
    "password": "",
    "ssl_mode": "",
    "weight": 100,
    "healthy": true
  },
  "replicas": [],
  "failover_enabled": false,
  "health_check_interval": "60s",
  "max_retries": 1,
  "connection_timeout": "5s",
  "load_balancing": "round_robin"
}
EOF

    # Production configuration
    cat > "$CONFIG_DIR/prod.json" << EOF
{
  "type": "postgresql",
  "primary": {
    "id": "prod-primary",
    "host": "localhost",
    "port": 5432,
    "database": "hades_toolkit",
    "username": "hades",
    "password": "hades_password",
    "ssl_mode": "require",
    "weight": 100,
    "healthy": true
  },
  "replicas": [
    {
      "id": "prod-replica-1",
      "host": "localhost",
      "port": 5433,
      "database": "hades_toolkit_replica1",
      "username": "hades",
      "password": "hades_password",
      "ssl_mode": "require",
      "weight": 90,
      "healthy": true
    }
  ],
  "failover_enabled": true,
  "health_check_interval": "15s",
  "max_retries": 5,
  "connection_timeout": "8s",
  "load_balancing": "least_connections"
}
EOF

    success "Hot-swap configuration files created in $CONFIG_DIR"
}

# Main execution
main() {
    log "Starting PostgreSQL setup with passwordless sudo..."
    
    check_root
    setup_passwordless_sudo
    install_postgresql
    create_database
    create_replicas
    configure_postgresql
    setup_replica_instances
    test_connections
    create_hotswap_configs
    
    success "PostgreSQL setup completed successfully!"
    log ""
    log "Configuration files created in: /home/$(whoami)/.hades/database/"
    log "Available configurations:"
    log "  - dev.json (development with replicas)"
    log "  - sqlite_fallback.json (SQLite fallback)"
    log "  - prod.json (production with SSL)"
    log ""
    log "Database connections:"
    log "  - Primary: localhost:5432 (hades_toolkit)"
    log "  - Replica 1: localhost:5433 (hades_toolkit_replica1)"
    log "  - Replica 2: localhost:5434 (hades_toolkit_replica2)"
    log ""
    log "You can now use passwordless sudo for PostgreSQL commands!"
}

# Run main function
main "$@"
