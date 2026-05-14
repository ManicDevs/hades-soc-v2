#!/bin/bash
# Forever API - Installation and Setup Script
# Sets up the free tier maximizer as a system service

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SERVICE_NAME="forever-api"
INSTALL_DIR="/home/cerberus/Desktop/hades"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          Forever API - Installation Script                    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}❌ This script must be run as root (use sudo)${NC}"
        exit 1
    fi
}

# Function to build the binary
build_binary() {
    echo -e "${YELLOW}🔨 Building Forever API binary...${NC}"
    cd "$INSTALL_DIR"
    
    if ! go build -o bin/forever-api ./cmd/forever-api; then
        echo -e "${RED}❌ Build failed${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Build successful${NC}"
}

# Function to create environment file
create_env_file() {
    local env_file="/etc/forever-api.env"
    
    echo -e "${YELLOW}📝 Creating environment file...${NC}"
    
    cat > "$env_file" << 'EOF'
# Forever API Environment Configuration
# Add your API keys here

# Anthropic Claude (1000 requests/day free)
ANTHROPIC_API_KEY=your_anthropic_api_key_here

# Google Gemini (20 requests/day free)
GEMINI_API_KEY=your_gemini_api_key_here

# OpenAI GPT-4 (100 requests/day free)
OPENAI_API_KEY=your_openai_api_key_here

# Request settings
REQUEST_INTERVAL=30s
MAX_REQUESTS=0
MONITOR_INTERVAL=2m
VERBOSE=false
EOF

    chmod 600 "$env_file"
    echo -e "${GREEN}✅ Environment file created: $env_file${NC}"
    echo -e "${YELLOW}⚠️  Edit this file to add your API keys${NC}"
}

# Function to install systemd service
install_service() {
    echo -e "${YELLOW}🔧 Installing systemd service...${NC}"
    
    # Copy service file
    cp "$INSTALL_DIR/systemd/forever-api.service" "$SERVICE_FILE"
    
    # Reload systemd
    systemctl daemon-reload
    
    echo -e "${GREEN}✅ Service installed${NC}"
}

# Function to setup permissions
setup_permissions() {
    echo -e "${YELLOW}🔒 Setting up permissions...${NC}"
    
    # Create logs directory
    mkdir -p /var/log/forever-api
    chown cerberus:cerberus /var/log/forever-api
    
    # Set binary permissions
    chown cerberus:cerberus "$INSTALL_DIR/bin/forever-api"
    chmod 755 "$INSTALL_DIR/bin/forever-api"
    
    echo -e "${GREEN}✅ Permissions configured${NC}"
}

# Function to show usage instructions
show_usage() {
    echo ""
    echo -e "${BLUE}📚 Usage Instructions:${NC}"
    echo ""
    echo "1. Add your API keys to /etc/forever-api.env:"
    echo "   sudo nano /etc/forever-api.env"
    echo ""
    echo "2. Start the service:"
    echo "   sudo systemctl start forever-api"
    echo ""
    echo "3. Enable auto-start on boot:"
    echo "   sudo systemctl enable forever-api"
    echo ""
    echo "4. Check service status:"
    echo "   sudo systemctl status forever-api"
    echo ""
    echo "5. View logs:"
    echo "   sudo journalctl -u forever-api -f"
    echo ""
    echo "6. Stop the service:"
    echo "   sudo systemctl stop forever-api"
    echo ""
    echo "7. Restart the service:"
    echo "   sudo systemctl restart forever-api"
    echo ""
    echo -e "${YELLOW}💡 Quick test (run as user, not service):${NC}"
    echo "   cd $INSTALL_DIR"
    echo "   ./scripts/forever-api.sh -m 5 -i 10s -v"
    echo ""
}

# Main installation
main() {
    echo -e "${BLUE}🚀 Starting Forever API installation...${NC}"
    echo ""
    
    check_root
    build_binary
    create_env_file
    install_service
    setup_permissions
    
    echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
    show_usage
}

# Run main function
main "$@"
