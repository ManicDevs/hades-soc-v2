#!/bin/bash
# Hades API Quota Status Check Script
# Monitors and reports on API quota status across providers

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

STATUS_FILE="/tmp/hades_quota_status.json"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              Hades API Quota Status Check                  ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to check if jq is available
check_jq() {
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is required but not installed. Please install jq to parse JSON.${NC}"
        echo "Ubuntu/Debian: sudo apt-get install jq"
        echo "macOS: brew install jq"
        exit 1
    fi
}

# Function to display quota status
display_quota_status() {
    if [[ ! -f "$STATUS_FILE" ]]; then
        echo -e "${YELLOW}⚠️  Quota status file not found: $STATUS_FILE${NC}"
        echo "This might mean:"
        echo "  - The API client hasn't been initialized yet"
        echo "  - The quota monitor isn't running"
        echo "  - The system hasn't made any API calls yet"
        echo ""
        echo "To start the quota monitor, run:"
        echo "  go run cmd/quota-monitor/main.go"
        return
    fi

    echo -e "${YELLOW}📊 Current API Quota Status:${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Parse and display the JSON status
    if jq -e '.[]' "$STATUS_FILE" &> /dev/null; then
        local models=$(jq -r '.[] | @base64' "$STATUS_FILE")
        
        while IFS= read -r model; do
            local model_data=$(echo "$model" | base64 --decode)
            local model_name=$(echo "$model_data" | jq -r '.model')
            local provider=$(echo "$model_data" | jq -r '.provider')
            local used=$(echo "$model_data" | jq -r '.used')
            local limit=$(echo "$model_data" | jq -r '.limit')
            local remaining=$(echo "$model_data" | jq -r '.remaining')
            local is_exhausted=$(echo "$model_data" | jq -r '.is_exhausted')
            
            # Determine status indicator
            local status_icon="✅"
            local status_color="$GREEN"
            if [[ "$is_exhausted" == "true" ]]; then
                status_icon="❌"
                status_color="$RED"
            elif [[ "$remaining" -lt 5 ]]; then
                status_icon="⚠️"
                status_color="$YELLOW"
            fi
            
            echo -e "${status_color}Model: $model_name ($provider)${NC}"
            echo -e "  Status: ${status_color}$status_icon $used/$limit ($remaining remaining)${NC}"
            echo ""
        done <<< "$models"
    else
        echo -e "${RED}Error parsing quota status file${NC}"
    fi
}

# Function to check API keys
check_api_keys() {
    echo -e "${YELLOW}🔑 API Key Configuration:${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    local keys=("ANTHROPIC_API_KEY" "GEMINI_API_KEY" "OPENAI_API_KEY")
    local providers=("Anthropic" "Gemini" "OpenAI")
    
    for i in "${!keys[@]}"; do
        local key="${keys[$i]}"
        local provider="${providers[$i]}"
        local value="${!key}"
        
        if [[ -n "$value" && "$value" != "your_api_key_here" && "$value" != *"CHANGE_ME"* ]]; then
            echo -e "  ${GREEN}✅ $provider: Configured${NC}"
        else
            echo -e "  ${RED}❌ $provider: Not configured${NC}"
        fi
    done
    echo ""
}

# Function to show recommendations
show_recommendations() {
    echo -e "${YELLOW}💡 Recommendations:${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if [[ -f "$STATUS_FILE" ]]; then
        local exhausted_count=$(jq '[.[] | select(.is_exhausted == true)] | length' "$STATUS_FILE" 2>/dev/null || echo "0")
        local low_count=$(jq '[.[] | select(.remaining < 5 and .is_exhausted == false)] | length' "$STATUS_FILE" 2>/dev/null || echo "0")
        
        if [[ "$exhausted_count" -gt 0 ]]; then
            echo -e "${RED}• $exhausted_count model(s) have exhausted quota${NC}"
            echo -e "${RED}  - Consider upgrading to paid tiers${NC}"
            echo -e "${RED}  - Wait for daily quota reset (24 hours)${NC}"
            echo -e "${RED}  - Configure additional API providers${NC}"
        fi
        
        if [[ "$low_count" -gt 0 ]]; then
            echo -e "${YELLOW}• $low_count model(s) have low quota${NC}"
            echo -e "${YELLOW}  - Monitor usage closely${NC}"
            echo -e "${YELLOW}  - Prepare backup providers${NC}"
        fi
        
        if [[ "$exhausted_count" -eq 0 && "$low_count" -eq 0 ]]; then
            echo -e "${GREEN}• All models have healthy quota levels${NC}"
            echo -e "${GREEN}• Continue normal operation${NC}"
        fi
    else
        echo -e "${YELLOW}• Start the quota monitor to track usage${NC}"
        echo -e "${YELLOW}• Configure API keys for all providers${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}• To view real-time quota status:${NC}"
    echo -e "  ${BLUE}tail -f $STATUS_FILE${NC}"
    echo ""
    echo -e "${BLUE}• To reset quota tracking:${NC}"
    echo -e "  ${BLUE}rm $STATUS_FILE${NC}"
    echo ""
}

# Main execution
main() {
    check_jq
    display_quota_status
    check_api_keys
    show_recommendations
}

# Run main function
main "$@"
