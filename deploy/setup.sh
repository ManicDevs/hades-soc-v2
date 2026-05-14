#!/bin/bash
# HADES-V2 Initial Configuration Setup Script

set -e

HADES_HOME="/opt/hades"
HADES_CONFIG="/etc/hades/hades.yaml"
HADES_DATA="/opt/hades/data"
HADES_LOGS="/var/log/hades"
HADES_CERTS="/etc/hades/certs"

echo "🚀 HADES-V2 Initial Configuration Setup"
echo "===================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ This script must be run as root"
    exit 1
fi

# Create directory structure
echo "📁 Creating HADES directory structure..."
mkdir -p "$HADES_HOME"/{bin,config,data,logs}
mkdir -p "$HADES_CERTS"
mkdir -p "$(dirname "$HADES_CONFIG")"
mkdir -p "$HADES_DATA"
mkdir -p "$HADES_LOGS"

# Generate SSL certificates
echo "🔐 Generating SSL certificates..."
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "$HADES_CERTS/server.key" \
    -out "$HADES_CERTS/server.crt" \
    -subj "/C=US/ST=CA/L=San Francisco/O=HADES/CN=hades.local" \
    2>/dev/null

chmod 600 "$HADES_CERTS/server.key"
chmod 644 "$HADES_CERTS/server.crt"

echo "✅ SSL certificates generated"

# Set up logging
echo "📋 Setting up logging..."
touch "$HADES_LOGS/hades.log"
touch "$HADES_LOGS/access.log"
touch "$HADES_LOGS/error.log"
touch "$HADES_LOGS/audit.log"

# Create systemd service
echo "⚙️ Creating systemd service..."
cat > /etc/systemd/system/hades.service << 'EOF'
[Unit]
Description=HADES-V2 Enterprise Security Framework
After=network.target

[Service]
Type=forking
User=root
Group=root
WorkingDirectory=$HADES_HOME
ExecStart=$HADES_HOME/bin/hades serve --config $HADES_CONFIG
ExecReload=/bin/kill -USR1 $MAINPID
KillMode=mixed
TimeoutStopSec=5
PrivateTmp=true
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable hades.service

echo "✅ systemd service created and enabled"

# Set up firewall rules
echo "🔥 Setting up firewall rules..."
iptables -F INPUT
iptables -F OUTPUT
iptables -F FORWARD

# Allow established connections
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# Allow SSH
iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# Allow HADES services
iptables -A INPUT -p tcp --dport 8443 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# Allow P2P ports
iptables -A INPUT -p tcp --dport 19001 -j ACCEPT
iptables -A INPUT -p tcp --dport 19002 -j ACCEPT
iptables -A INPUT -p tcp --dport 19003 -j ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT

# Default deny
iptables -A INPUT -j DROP

# Save rules
iptables-save > /etc/iptables/rules.v4
echo "✅ Firewall rules configured"

# Create admin user
echo "👤 Creating HADES admin user..."
if ! id "hades" &>/dev/null; then
    useradd -m -s /bin/bash -c "HADES-V2 Admin User" hades
    usermod -aG sudo hades
    echo "hades:$(openssl rand -base64 32)" | chpasswd
    echo "✅ HADES admin user created"
else
    echo "✅ HADES admin user already exists"
fi

# Set permissions
echo "🔒 Setting permissions..."
chown -R hades:hades "$HADES_HOME"
chmod 755 "$HADES_HOME"
chmod 644 "$HADES_CONFIG"
chmod 755 "$HADES_DATA"
chmod 755 "$HADES_LOGS"

# Create startup script
echo "🚀 Creating startup script..."
cat > "$HADES_HOME/start.sh" << 'EOF'
#!/bin/bash
echo "Starting HADES-V2..."
cd $HADES_HOME
./bin/hades serve --config $HADES_CONFIG
EOF

chmod +x "$HADES_HOME/start.sh"

echo "✅ Startup script created"

# Health check script
echo "🏥 Creating health check script..."
cat > "$HADES_HOME/health-check.sh" << 'EOF'
#!/bin/bash
echo "HADES-V2 Health Check"
echo "===================="

# Check if HADES process is running
if pgrep -f "hades" > /dev/null; then
    echo "✅ HADES process: RUNNING"
else
    echo "❌ HADES process: NOT RUNNING"
fi

# Check ports
for port in 8443 8080 19001 19002 19003; do
    if netstat -tlnp | grep -q ":$port "; then
        echo "✅ Port $port: LISTENING"
    else
        echo "❌ Port $port: NOT LISTENING"
    fi
done

# Check configuration
if [ -f "$HADES_CONFIG" ]; then
    echo "✅ Configuration: EXISTS"
else
    echo "❌ Configuration: MISSING"
fi

# Check logs
if [ -f "$HADES_LOGS/hades.log" ]; then
    echo "✅ Log file: EXISTS"
    echo "Last 5 lines:"
    tail -5 "$HADES_LOGS/hades.log"
else
    echo "❌ Log file: MISSING"
fi
EOF

chmod +x "$HADES_HOME/health-check.sh"

echo "✅ Health check script created"

# Create backup script
echo "💾 Creating backup script..."
cat > "$HADES_HOME/backup.sh" << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/hades/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

echo "Creating HADES backup: $DATE"

# Backup configuration
tar -czf "$BACKUP_DIR/hades-config-$DATE.tar.gz" -C /etc/hades .

# Backup data
tar -czf "$BACKUP_DIR/hades-data-$DATE.tar.gz" -C "$HADES_DATA" .

# Backup logs
tar -czf "$BACKUP_DIR/hades-logs-$DATE.tar.gz" -C "$HADES_LOGS" .

# Cleanup old backups (keep last 7 days)
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
ls -lh "$BACKUP_DIR"
EOF

chmod +x "$HADES_HOME/backup.sh"

echo "✅ Backup script created"

# Create update script
echo "🔄 Creating update script..."
cat > "$HADES_HOME/update.sh" << 'EOF'
#!/bin/bash
echo "Updating HADES-V2..."

# Stop services
systemctl stop hades

# Backup current
$HADES_HOME/backup.sh

# Update binary (if new version available)
# This would integrate with your build system

# Start services
systemctl start hades

echo "HADES-V2 update completed"
EOF

chmod +x "$HADES_HOME/update.sh"

echo "✅ Update script created"

echo ""
echo "🎯 HADES-V2 Initial Setup Complete!"
echo "=================================="
echo ""
echo "📁 Directory Structure:"
echo "  $HADES_HOME/ - HADES installation"
echo "  $HADES_CONFIG - Configuration file"
echo "  $HADES_DATA - Data directory"
echo "  $HADES_LOGS - Log directory"
echo ""
echo "🚀 Services:"
echo "  systemd: hades.service"
echo "  firewall: iptables rules configured"
echo "  ssl: certificates generated"
echo ""
echo "👤 User Management:"
echo "  admin: hades (with sudo access)"
echo "  password: randomly generated"
echo ""
echo "🔧 Scripts Created:"
echo "  start.sh - Start HADES"
echo "  health-check.sh - Health monitoring"
echo "  backup.sh - Backup automation"
echo "  update.sh - Update automation"
echo ""
echo "🎮 Next Steps:"
echo "  1. Copy HADES binary to $HADES_HOME/bin/hades"
echo "  2. Run: systemctl start hades"
echo "  3. Access: https://localhost:8443"
echo "  4. Monitor: $HADES_HOME/health-check.sh"
echo ""
echo "✅ HADES-V2 is ready for production deployment!"
