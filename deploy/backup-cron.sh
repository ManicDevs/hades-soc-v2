#!/bin/bash
# HADES-V2 Backup Script
# Automated backup system for all environments

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/opt/hades/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
COMPRESSION="${COMPRESSION:-gzip}"
ENCRYPT="${ENCRYPT:-true}"
GPG_RECIPIENT="${GPG_RECIPIENT:-backup@hades.localhost}"
LOG_FILE="${LOG_FILE:-/var/log/hades-backup.log}"
ENVIRONMENT="${ENVIRONMENT:-prod}"

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

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Create backup directory
create_backup_dir() {
    local backup_date=$(date +%Y%m%d_%H%M%S)
    local env_backup_dir="${BACKUP_DIR}/${ENVIRONMENT}/${backup_date}"
    
    mkdir -p "$env_backup_dir"
    echo "$env_backup_dir"
}

# Backup databases
backup_databases() {
    local backup_dir="$1"
    log_info "Starting database backups..."
    
    # PostgreSQL backup
    if docker ps | grep -q "postgres-${ENVIRONMENT}"; then
        log "Backing up PostgreSQL..."
        docker exec "postgres-${ENVIRONMENT}" pg_dump -U hades_${ENVIRONMENT}_user hades_${ENVIRONMENT} | \
            gzip > "${backup_dir}/postgresql_backup.sql.gz"
        log_info "PostgreSQL backup completed"
    else
        log_warn "PostgreSQL container not found"
    fi
    
    # Redis backup
    if docker ps | grep -q "redis-${ENVIRONMENT}"; then
        log "Backing up Redis..."
        docker exec "redis-${ENVIRONMENT}" redis-cli --rdb - | \
            gzip > "${backup_dir}/redis_backup.rdb.gz"
        log_info "Redis backup completed"
    else
        log_warn "Redis container not found"
    fi
}

# Backup application data
backup_app_data() {
    local backup_dir="$1"
    log_info "Starting application data backup..."
    
    # Backup data directory
    if [ -d "/opt/hades/data" ]; then
        log "Backing up application data..."
        tar -czf "${backup_dir}/app_data.tar.gz" -C /opt/hades data/
        log_info "Application data backup completed"
    else
        log_warn "Application data directory not found"
    fi
    
    # Backup configuration files
    if [ -d "/opt/hades/config" ]; then
        log "Backing up configuration..."
        tar -czf "${backup_dir}/config.tar.gz" -C /opt/hades config/
        log_info "Configuration backup completed"
    fi
    
    # Backup logs
    if [ -d "/opt/hades/logs" ]; then
        log "Backing up logs..."
        tar -czf "${backup_dir}/logs.tar.gz" -C /opt/hades logs/
        log_info "Logs backup completed"
    fi
}

# Backup Docker volumes
backup_docker_volumes() {
    local backup_dir="$1"
    log_info "Starting Docker volumes backup..."
    
    # Get all HADES volumes
    local volumes=$(docker volume ls --filter "name=hades" --format "{{.Name}}")
    
    for volume in $volumes; do
        log "Backing up Docker volume: $volume"
        docker run --rm \
            -v "$volume:/data:ro" \
            -v "$backup_dir:/backup" \
            alpine tar -czf "/backup/volume_${volume}.tar.gz" -C /data .
        log_info "Volume $volume backup completed"
    done
}

# Backup SSL certificates
backup_ssl_certs() {
    local backup_dir="$1"
    log_info "Starting SSL certificates backup..."
    
    if [ -d "/etc/ssl/certs/hades" ]; then
        log "Backing up SSL certificates..."
        tar -czf "${backup_dir}/ssl_certs.tar.gz" -C /etc/ssl/certs hades/
        log_info "SSL certificates backup completed"
    fi
    
    if [ -d "/etc/ssl/private/hades" ]; then
        log "Backing up SSL private keys..."
        tar -czf "${backup_dir}/ssl_private.tar.gz" -C /etc/ssl/private hades/
        log_info "SSL private keys backup completed"
    fi
}

# Encrypt backup files
encrypt_backups() {
    local backup_dir="$1"
    
    if [ "$ENCRYPT" = "true" ]; then
        log_info "Encrypting backup files..."
        
        for file in "${backup_dir}"/*; do
            if [ -f "$file" ] && [[ "$file" != *.gpg ]]; then
                log "Encrypting: $(basename "$file")"
                gpg --trust-model always --encrypt -r "$GPG_RECIPIENT" --output "$file.gpg" "$file"
                rm "$file"
            fi
        done
        
        log_info "Backup encryption completed"
    fi
}

# Create backup manifest
create_manifest() {
    local backup_dir="$1"
    local manifest_file="${backup_dir}/manifest.json"
    
    log "Creating backup manifest..."
    
    cat > "$manifest_file" << EOF
{
    "backup_date": "$(date -Iseconds)",
    "environment": "$ENVIRONMENT",
    "version": "$(git describe --tags --always 2>/dev/null || echo 'unknown')",
    "hostname": "$(hostname)",
    "backup_type": "automated",
    "encryption_enabled": $ENCRYPT,
    "compression": "$COMPRESSION",
    "files": [
EOF
    
    # Add file list
    first=true
    for file in "${backup_dir}"/*; do
        if [ -f "$file" ]; then
            if [ "$first" = false ]; then
                echo "," >> "$manifest_file"
            fi
            echo "        \"$(basename "$file")\"" >> "$manifest_file"
            first=false
        fi
    done
    
    cat >> "$manifest_file" << EOF
    ],
    "checksums": {
EOF
    
    # Add checksums
    first=true
    for file in "${backup_dir}"/*; do
        if [ -f "$file" ]; then
            checksum=$(sha256sum "$file" | cut -d' ' -f1)
            if [ "$first" = false ]; then
                echo "," >> "$manifest_file"
            fi
            echo "        \"$(basename "$file")\": \"$checksum\"" >> "$manifest_file"
            first=false
        fi
    done
    
    cat >> "$manifest_file" << EOF
    }
}
EOF
    
    log_info "Backup manifest created"
}

# Cleanup old backups
cleanup_old_backups() {
    log_info "Cleaning up old backups (retention: ${RETENTION_DAYS} days)..."
    
    find "$BACKUP_DIR" -type d -name "${ENVIRONMENT}/*" -mtime +${RETENTION_DAYS} -exec rm -rf {} \;
    
    log_info "Old backup cleanup completed"
}

# Verify backup integrity
verify_backup() {
    local backup_dir="$1"
    log_info "Verifying backup integrity..."
    
    local errors=0
    
    # Check manifest exists
    if [ ! -f "${backup_dir}/manifest.json" ]; then
        log_error "Backup manifest not found"
        ((errors++))
    fi
    
    # Verify file checksums
    if [ -f "${backup_dir}/manifest.json" ]; then
        for file in "${backup_dir}"/*; do
            if [ -f "$file" ] && [[ "$file" != *.json ]]; then
                expected_checksum=$(jq -r ".checksums[\"$(basename "$file")\"]" "${backup_dir}/manifest.json")
                actual_checksum=$(sha256sum "$file" | cut -d' ' -f1)
                
                if [ "$expected_checksum" != "$actual_checksum" ]; then
                    log_error "Checksum mismatch for $(basename "$file")"
                    ((errors++))
                fi
            fi
        done
    fi
    
    if [ $errors -eq 0 ]; then
        log_info "Backup verification completed successfully"
    else
        log_error "Backup verification failed with $errors errors"
        return 1
    fi
}

# Send notification
send_notification() {
    local status="$1"
    local backup_dir="$2"
    
    if [ -n "${SLACK_WEBHOOK:-}" ]; then
        local message
        if [ "$status" = "success" ]; then
            message="✅ HADES-V2 backup completed successfully for $ENVIRONMENT environment"
        else
            message="❌ HADES-V2 backup failed for $ENVIRONMENT environment"
        fi
        
        curl -X POST "$SLACK_WEBHOOK" \
            -H 'Content-type: application/json' \
            --data "{\"text\":\"$message\"}" \
            2>/dev/null || log_warn "Failed to send Slack notification"
    fi
}

# Main backup function
main() {
    log "Starting HADES-V2 backup for $ENVIRONMENT environment..."
    
    # Create backup directory
    local backup_dir
    backup_dir=$(create_backup_dir)
    
    # Perform backups
    backup_databases "$backup_dir"
    backup_app_data "$backup_dir"
    backup_docker_volumes "$backup_dir"
    backup_ssl_certs "$backup_dir"
    
    # Create manifest
    create_manifest "$backup_dir"
    
    # Verify backup
    if verify_backup "$backup_dir"; then
        # Encrypt backups
        encrypt_backups "$backup_dir"
        
        # Cleanup old backups
        cleanup_old_backups
        
        log_info "Backup completed successfully: $backup_dir"
        send_notification "success" "$backup_dir"
        
        exit 0
    else
        log_error "Backup verification failed"
        send_notification "failure" "$backup_dir"
        
        exit 1
    fi
}

# Handle signals
trap 'log_error "Backup interrupted"; exit 1' INT TERM

# Run main function
main "$@"
