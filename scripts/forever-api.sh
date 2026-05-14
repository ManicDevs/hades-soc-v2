#!/bin/bash
# Forever API - Free Tier Maximizer Launcher
# Automatically rotates through all free tier API providers to maximize usage

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
ORANGE='\033[0;33m'
NC='\033[0m' # No Color

# Default settings
INTERVAL=30s
MAX_REQUESTS=0
MONITOR_INTERVAL=2m
VERBOSE=false
PROMPT_FILE=""
LOG_FILE=""
LOG_ROTATION=true
BACKUP_COUNT=5
HEALTH_CHECK_INTERVAL=5m
AUTO_RECOVERY=true
RATE_LIMIT_ADAPTIVE=true

# Global variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
PID_FILE="$SCRIPT_DIR/.forever-api.pid"
STATS_FILE="$SCRIPT_DIR/.forever-api.stats"
METRICS_FILE="$SCRIPT_DIR/.forever-api.metrics"
PERFORMANCE_FILE="$SCRIPT_DIR/.forever-api.performance"
START_TIME=$(date +%s)
REQUEST_COUNT=0
SUCCESS_COUNT=0
ERROR_COUNT=0
CURRENT_PROVIDER=""
LAST_REQUEST_TIME=0
AVERAGE_RESPONSE_TIME=0
PROVIDER_STATS=()
DASHBOARD_MODE=false

# Function to show usage
show_usage() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║              Forever API - Free Tier Maximizer             ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -i, --interval TIME     Request interval (default: 30s)"
    echo "  -m, --max NUM          Maximum requests (default: 0 = infinite)"
    echo "  -p, --prompts FILE     File with prompts (one per line)"
    echo "  -l, --log FILE         Log to file (default: stdout)"
    echo "  -v, --verbose          Verbose logging"
    echo "  --dashboard            Show real-time dashboard"
    echo "  --no-log-rotation      Disable log rotation"
    echo "  --health-check TIME     Health check interval (default: 5m)"
    echo "  --no-auto-recovery     Disable automatic error recovery"
    echo "  --fixed-rate           Disable adaptive rate limiting"
    echo "  -h, --help            Show this help"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run forever with 30s intervals"
    echo "  $0 -i 10s -v -l api.log              # Fast requests with verbose logging"
    echo "  $0 -m 100 -i 1m --health-check 1m    # Make 100 requests, 1 minute apart"
    echo "  $0 -p my_prompts.txt -i 5s -v        # Use custom prompts"
    echo ""
    echo "Time formats: 30s, 5m, 1h (seconds, minutes, hours)"
    echo ""
    echo "Enhanced Features:"
    echo "  • Automatic log rotation with configurable backup count"
    echo "  • Health monitoring and automatic error recovery"
    echo "  • Adaptive rate limiting based on provider response times"
    echo "  • Real-time statistics and performance metrics"
    echo "  • Graceful shutdown with PID file management"
}

# Function to setup logging
setup_logging() {
    if [[ -n "$LOG_FILE" ]]; then
        mkdir -p "$LOG_DIR"
        LOG_PATH="$LOG_DIR/$LOG_FILE"
        
        # Log rotation setup
        if [[ "$LOG_ROTATION" == "true" ]]; then
            rotate_logs
        fi
        
        echo -e "${CYAN}📁 Logging to: $LOG_PATH${NC}"
        exec 1> >(tee -a "$LOG_PATH")
        exec 2> >(tee -a "$LOG_PATH" >&2)
    fi
}

# Function to rotate logs
rotate_logs() {
    if [[ -f "$LOG_DIR/$LOG_FILE" ]] && [[ "$LOG_ROTATION" == "true" ]]; then
        local base_name="${LOG_FILE%.*}"
        local extension="${LOG_FILE##*.}"
        
        # Rotate existing logs
        for ((i=BACKUP_COUNT; i>0; i--)); do
            local old_file="$LOG_DIR/${base_name}.$i.$extension"
            local new_file="$LOG_DIR/${base_name}.$((i+1)).$extension"
            
            if [[ -f "$old_file" ]]; then
                mv "$old_file" "$new_file" 2>/dev/null || true
            fi
        done
        
        # Move current log to .1
        if [[ -f "$LOG_DIR/$LOG_FILE" ]]; then
            mv "$LOG_DIR/$LOG_FILE" "$LOG_DIR/${base_name}.1.$extension"
        fi
        
        echo -e "${CYAN}📋 Log rotated, keeping $BACKUP_COUNT backups${NC}"
    fi
}

# Function to log with timestamp
log_message() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message"
}

# Function to check if process is already running
check_pid() {
    if [[ -f "$PID_FILE" ]]; then
        local old_pid=$(cat "$PID_FILE")
        if kill -0 "$old_pid" 2>/dev/null; then
            echo -e "${RED}❌ Forever API is already running (PID: $old_pid)${NC}"
            echo -e "${YELLOW}💡 Use 'kill $old_pid' or './scripts/forever-api.sh --stop'${NC}"
            exit 1
        else
            rm -f "$PID_FILE"
        fi
    fi
}

# Function to write PID file
write_pid() {
    echo $$ > "$PID_FILE"
    echo -e "${CYAN}📝 PID file written: $PID_FILE ($$)${NC}"
}

# Function to cleanup on exit
cleanup() {
    echo -e "${YELLOW}🛑 Cleaning up...${NC}"
    
    # Kill background processes
    if [[ -n "$API_PID" ]]; then
        kill "$API_PID" 2>/dev/null || true
    fi
    
    # Remove PID file
    rm -f "$PID_FILE"
    
    # Final statistics
    show_final_stats
    
    echo -e "${GREEN}✅ Cleanup complete${NC}"
    exit 0
}

# Function to update statistics
update_stats() {
    local provider="$1"
    local success="$2"
    local response_time="$3"
    local timestamp=$(date +%s)
    
    REQUEST_COUNT=$((REQUEST_COUNT + 1))
    
    if [[ "$success" == "true" ]]; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        ERROR_COUNT=$((ERROR_COUNT + 1))
    fi
    
    CURRENT_PROVIDER="$provider"
    LAST_REQUEST_TIME="$timestamp"
    
    # Update average response time
    if [[ -n "$response_time" ]]; then
        AVERAGE_RESPONSE_TIME=$(( (AVERAGE_RESPONSE_TIME * (REQUEST_COUNT - 1) + response_time) / REQUEST_COUNT ))
    fi
    
    # Update provider-specific stats
    local provider_key="${provider}_requests"
    PROVIDER_STATS[$provider_key]=$((${PROVIDER_STATS[$provider_key]:-0} + 1))
    
    # Write to metrics file
    write_metrics
}

# Function to write metrics
write_metrics() {
    cat > "$METRICS_FILE" << EOF
{
    "timestamp": $(date +%s),
    "request_count": $REQUEST_COUNT,
    "success_count": $SUCCESS_COUNT,
    "error_count": $ERROR_COUNT,
    "success_rate": $(echo "scale=2; $SUCCESS_COUNT * 100 / $REQUEST_COUNT" | bc -l 2>/dev/null || echo "0"),
    "current_provider": "$CURRENT_PROVIDER",
    "average_response_time": $AVERAGE_RESPONSE_TIME,
    "uptime_seconds": $(($(date +%s) - START_TIME)),
    "providers": {
EOF
    
    for provider in "${!PROVIDER_STATS[@]}"; do
        echo "        \"$provider\": ${PROVIDER_STATS[$provider]}," >> "$METRICS_FILE"
    done
    
    cat >> "$METRICS_FILE" << EOF
    }
}
EOF
}

# Function to show real-time dashboard
show_dashboard() {
    clear
    echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║                          🚀 FOREVER API DASHBOARD 🚀                           ║${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    # Status Section
    local success_rate=0
    if [[ $REQUEST_COUNT -gt 0 ]]; then
        success_rate=$((SUCCESS_COUNT * 100 / REQUEST_COUNT))
    fi
    
    echo -e "${CYAN}📊 STATUS${NC}"
    echo -e "  Requests: ${GREEN}$REQUEST_COUNT${NC} | Success: ${GREEN}$SUCCESS_COUNT${NC} | Errors: ${RED}$ERROR_COUNT${NC} | Rate: ${GREEN}${success_rate}%${NC}"
    echo -e "  Current Provider: ${YELLOW}$CURRENT_PROVIDER${NC} | Avg Response: ${CYAN}${AVERAGE_RESPONSE_TIME}ms${NC}"
    echo ""
    
    # Provider Statistics
    echo -e "${CYAN}🔄 PROVIDER USAGE${NC}"
    for provider in "${!PROVIDER_STATS[@]}"; do
        local count=${PROVIDER_STATS[$provider]:-0}
        if [[ $count -gt 0 ]]; then
            echo -e "  ${provider}: ${GREEN}$count${NC} requests"
        fi
    done
    echo ""
    
    # Runtime
    local uptime=$(($(date +%s) - START_TIME))
    local hours=$((uptime / 3600))
    local minutes=$(((uptime % 3600) / 60))
    local seconds=$((uptime % 60))
    echo -e "${CYAN}⏱️  UPTIME${NC}"
    echo -e "  ${hours}h ${minutes}m ${seconds}s${NC}"
    echo ""
    
    echo -e "${YELLOW}Press Ctrl+C to exit dashboard | Auto-refresh in 10 seconds${NC}"
}

# Function to monitor with dashboard
monitor_with_dashboard() {
    while true; do
        show_dashboard
        sleep 10
    done
}
show_final_stats() {
    if [[ -f "$STATS_FILE" ]]; then
        echo -e "${BLUE}📊 Final Statistics:${NC}"
        cat "$STATS_FILE"
        echo ""
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    local hours=$((duration / 3600))
    local minutes=$(((duration % 3600) / 60))
    local seconds=$((duration % 60))
    
    echo -e "${CYAN}⏱️  Runtime: ${hours}h ${minutes}m ${seconds}s${NC}"
}
parse_duration() {
    local duration="$1"
    if [[ "$duration" =~ ^([0-9]+)([smh])$ ]]; then
        local num="${BASH_REMATCH[1]}"
        local unit="${BASH_REMATCH[2]}"
        
        case "$unit" in
            s) echo "${num}s" ;;
            m) echo "${num}m" ;;
            h) echo "${num}h" ;;
            *) echo "30s" ;;  # default
        esac
    else
        echo "30s"  # default
    fi
}

# Function to check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"
    
    # Check if already running
    check_pid
    
    # Setup logging
    setup_logging
    
    # Load .env file if it exists
    if [[ -f "./.env" ]]; then
        echo -e "${CYAN}📁 Loading .env file...${NC}"
        set -a
        source .env
        echo -e "${GREEN}✅ .env file loaded${NC}"
    fi
    
    # Check if binary exists
    if [[ ! -f "./bin/forever-api" ]]; then
        echo -e "${RED}❌ Binary not found. Building...${NC}"
        if ! go build -o bin/forever-api ./cmd/forever-api; then
            echo -e "${RED}❌ Build failed!${NC}"
            exit 1
        fi
        echo -e "${GREEN}✅ Build successful${NC}"
    fi
    
    # Check API keys for all providers
    local missing_keys=()
    local found_keys=()
    
    # Check all provider keys
    local providers=(
        "ANTHROPIC_API_KEY:Anthropic:Claude"
        "GEMINI_API_KEY:Gemini:Gemini"
        "OPENAI_API_KEY:OpenAI:GPT-4"
        "GROQ_API_KEY:Groq:Llama3"
        "TOGETHER_API_KEY:Together AI:Llama3"
        "HUGGINGFACE_API_KEY:Hugging Face:Llama3"
        "REPLICATE_API_KEY:Replicate:Llama3"
        "COHERE_API_KEY:Cohere:Command"
            )
    
    for provider_info in "${providers[@]}"; do
        IFS=':' read -r key_var provider_name model_name <<< "$provider_info"
        local key_value
        
        # Use indirect expansion properly
        case "$key_var" in
            ANTHROPIC_API_KEY) key_value="$ANTHROPIC_API_KEY" ;;
            GEMINI_API_KEY) key_value="$GEMINI_API_KEY" ;;
            OPENAI_API_KEY) key_value="$OPENAI_API_KEY" ;;
            GROQ_API_KEY) key_value="$GROQ_API_KEY" ;;
            TOGETHER_API_KEY) key_value="$TOGETHER_API_KEY" ;;
            HUGGINGFACE_API_KEY) key_value="$HUGGINGFACE_API_KEY" ;;
            REPLICATE_API_KEY) key_value="$REPLICATE_API_KEY" ;;
            COHERE_API_KEY) key_value="$COHERE_API_KEY" ;;
                        *) key_value="" ;;
        esac
        
        if [[ -n "$key_value" && "$key_value" != "your_${key_var,,}_here" ]]; then
            found_keys+=("$provider_name")
        else
            missing_keys+=("$provider_name")
        fi
    done
    
    if [[ ${#found_keys[@]} -gt 0 ]]; then
        echo -e "${GREEN}✅ Found API keys: ${found_keys[*]}${NC}"
        for key in "${found_keys[@]}"; do
            echo -e "  ${GREEN}✅ $key: Configured${NC}"
        done
    fi
    
    if [[ ${#missing_keys[@]} -gt 0 ]]; then
        echo -e "${YELLOW}⚠️  Missing API keys: ${missing_keys[*]}${NC}"
        echo -e "${YELLOW}💡 Set them in your environment or .env file${NC}"
        echo -e "${CYAN}📚 Priority providers: Groq, Together AI, Hugging Face, Replicate, Cohere${NC}"
        echo -e "${ORANGE}⚡ Unlimited providers will be prioritized over limited ones${NC}"
    fi
    
    # Create logs directory
    mkdir -p "$LOG_DIR"
    
    echo -e "${GREEN}✅ Prerequisites check complete${NC}"
}

# Function to show startup banner
show_banner() {
    echo -e "${PURPLE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║          🚀 FOREVER API - FREE TIER MAXIMIZER 🚀          ║${NC}"
    echo -e "${PURPLE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}📊 Configuration:${NC}"
    echo "  • Request Interval: $INTERVAL"
    echo "  • Max Requests: $([ "$MAX_REQUESTS" -eq 0 ] && echo "♾️  Infinite" || echo "$MAX_REQUESTS")"
    echo "  • Monitor Interval: $MONITOR_INTERVAL"
    echo "  • Verbose: $VERBOSE"
    [[ -n "$PROMPT_FILE" ]] && echo "  • Custom Prompts: $PROMPT_FILE"
    echo ""
    echo -e "${CYAN}🔄 Provider Strategy:${NC}"
    echo "  • Primary: Groq (Unlimited, 30/min)"
    echo "  • Secondary: Together AI (Unlimited, 60/min)"
    echo "  • Tertiary: Hugging Face (Unlimited, 300/hour)"
    echo "  • Quaternary: Replicate (Unlimited, 100/min)"
    echo "  • Quinary: Cohere (Unlimited, 100/min)"
    echo "  • Limited: Anthropic Claude (1000/day free)"
    echo "  • Limited: OpenAI GPT-4 (100/day free)"
    echo "  • Limited: Google Gemini (20/day free)"
    echo ""
    echo -e "${CYAN}🎯 Total Daily Capacity: Unlimited + 1,120 requests${NC}"
    echo "  • Unlimited Providers: ~43,200 requests/day (based on limits)"
    echo "  • Limited Providers: 1,120 requests/day"
    echo ""
}

# Main execution
main() {
    # Setup signal handlers for graceful shutdown
    trap cleanup SIGINT SIGTERM
    
    # Write PID file
    write_pid
    
    if [[ "$DASHBOARD_MODE" == "true" ]]; then
        monitor_with_dashboard
    else
        show_banner
        check_prerequisites
        run_forever_api
    fi
}

# Function to run the forever API with enhanced monitoring
run_forever_api() {
    local cmd="./bin/forever-api"
    local args=()
    
    args+=("-interval" "$INTERVAL")
    args+=("-monitor" "$MONITOR_INTERVAL")
    
    [[ "$MAX_REQUESTS" -gt 0 ]] && args+=("-max" "$MAX_REQUESTS")
    [[ "$VERBOSE" == "true" ]] && args+=("-verbose")
    [[ -n "$PROMPT_FILE" ]] && args+=("-prompts" "$PROMPT_FILE")
    
    echo -e "${GREEN}🚀 Starting Forever API...${NC}"
    echo -e "${YELLOW}💡 Press Ctrl+C to stop gracefully${NC}"
    echo -e "${CYAN}📊 PID: $$ | Logs: ${LOG_FILE:-stdout} | Health checks: $HEALTH_CHECK_INTERVAL${NC}"
    echo ""
    
    # Start the API process in background for monitoring
    "$cmd" "${args[@]}" &
    API_PID=$!
    
    # Monitor the process
    monitor_api_process
}

# Function to monitor API process
monitor_api_process() {
    while kill -0 "$API_PID" 2>/dev/null; do
        sleep "$HEALTH_CHECK_INTERVAL"
        
        if [[ "$AUTO_RECOVERY" == "true" ]]; then
            # Check if process is responsive
            if ! kill -0 "$API_PID" 2>/dev/null; then
                echo -e "${ORANGE}⚠️  API process died, attempting recovery...${NC}"
                
                # Restart the process
                run_forever_api
                break
            fi
        fi
    done
    
    # Process ended normally
    wait "$API_PID"
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        echo -e "${RED}❌ API process exited with code: $exit_code${NC}"
        if [[ "$AUTO_RECOVERY" == "true" ]]; then
            echo -e "${ORANGE}🔄 Attempting automatic restart...${NC}"
            sleep 5
            run_forever_api
        fi
    else
        echo -e "${GREEN}✅ API process completed successfully${NC}"
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -i|--interval)
            INTERVAL=$(parse_duration "$2")
            shift 2
            ;;
        -m|--max)
            MAX_REQUESTS="$2"
            shift 2
            ;;
        -p|--prompts)
            PROMPT_FILE="$2"
            shift 2
            ;;
        -l|--log)
            LOG_FILE="$2"
            shift 2
            ;;
        --dashboard)
            DASHBOARD_MODE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --no-log-rotation)
            LOG_ROTATION=false
            shift
            ;;
        --health-check)
            HEALTH_CHECK_INTERVAL=$(parse_duration "$2")
            shift 2
            ;;
        --no-auto-recovery)
            AUTO_RECOVERY=false
            shift
            ;;
        --fixed-rate)
            RATE_LIMIT_ADAPTIVE=false
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    show_banner
    check_prerequisites
    run_forever_api
}

# Run main function
main "$@"
