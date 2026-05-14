#!/bin/bash
# HADES-V2 Docker Entrypoint Script

set -e

HADES_BIN="/opt/hades/bin/hades"
HADES_CONFIG="/etc/hades/hades.yaml"
HADES_DATA="/opt/hades/data"
HADES_LOGS="/var/log/hades"

echo "🚀 Starting HADES-V2 Enterprise Security Framework"
echo "=============================================="
echo "Version: $HADES_VERSION"
echo "Build: $HADES_BUILD"
echo "Log Level: $HADES_LOG_LEVEL"
echo ""

# Ensure data directory exists
mkdir -p "$HADES_DATA" "$HADES_LOGS"

# Check configuration
if [ ! -f "$HADES_CONFIG" ]; then
    echo "⚠️  Configuration file not found, using defaults"
    mkdir -p "$(dirname "$HADES_CONFIG")"
    cat > "$HADES_CONFIG" << 'EOF'
server:
  host: "0.0.0.0"
  port: 8443
  
security:
  anti_analysis: true
  quantum_resistant: true
  self_healing: true
  
network:
  p2p:
    enabled: true
    ports: [19001, 19002, 19003]
  
logging:
  level: "info"
  file: "/var/log/hades/hades.log"
EOF
fi

# Set up logging
mkdir -p "$(dirname "$HADES_LOGS/hades.log")"
touch "$HADES_LOGS/hades.log"

# Initialize HADES
echo "🔧 Initializing HADES-V2..."
cd /opt/hades

# Handle command line arguments
case "${1:-serve}" in
    "serve")
        echo "🌐 Starting HADES-V2 web server..."
        exec "$HADES_BIN" serve --config "$HADES_CONFIG"
        ;;
    "api")
        echo "🔌 Starting HADES-V2 API server..."
        exec "$HADES_BIN" api start --port 8080 --config "$HADES_CONFIG"
        ;;
    "worker")
        echo "⚙️  Starting HADES-V2 worker..."
        exec "$HADES_BIN" worker start --config "$HADES_CONFIG"
        ;;
    "monitor")
        echo "📊 Starting HADES-V2 monitoring..."
        exec "$HADES_BIN" monitor --config "$HADES_CONFIG"
        ;;
    "health")
        echo "🏥 Running health check..."
        exec "$HADES_BIN" health --config "$HADES_CONFIG"
        ;;
    *)
        echo "❌ Unknown command: $1"
        echo "Available commands: serve, api, worker, monitor, health"
        exit 1
        ;;
esac
