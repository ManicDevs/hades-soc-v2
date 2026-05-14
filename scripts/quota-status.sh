#!/bin/bash

# Hades API Quota Status Display
# Shows current quota limits and usage for all configured providers

echo "╔════════════════════════════════════════════════════════════╗"
echo "║              Hades API Quota Status Report                ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Load environment variables from .env file
if [[ -f ".env" ]]; then
    set -a
    source .env
    set +a
    echo "✅ Loaded configuration from .env file"
else
    echo "⚠️  .env file not found, checking environment variables only"
fi
echo ""

# Check if API keys are configured
echo "🔑 API Key Configuration:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check each provider's API key
providers=(
    "ANTHROPIC_API_KEY:Anthropic Claude:1000:sk-ant-"
    "GEMINI_API_KEY:Google Gemini:20:AIza"
    "OPENAI_API_KEY:OpenAI GPT-4:100:sk-"
    "GROQ_API_KEY:Groq:UNLIMITED:gsk_"
    "TOGETHER_API_KEY:Together AI:UNLIMITED:"
    "HUGGINGFACE_API_KEY:Hugging Face:UNLIMITED:hf_"
    "REPLICATE_API_KEY:Replicate:UNLIMITED:r8_"
    "COHERE_API_KEY:Cohere:UNLIMITED:"
    "PERPLEXITY_API_KEY:Perplexity:5000:"
    "MISTRAL_API_KEY:Mistral AI:1000:"
)

total_daily_limit=0
configured_providers=0

for provider_info in "${providers[@]}"; do
    IFS=':' read -r env_var name limit prefix <<< "$provider_info"
    
    if [[ -n "${!env_var}" ]]; then
        configured_providers=$((configured_providers + 1))
        if [[ "$limit" != "UNLIMITED" ]]; then
            total_daily_limit=$((total_daily_limit + limit))
        fi
        echo "  ✅ $name: Configured (${limit:0:10} requests/day)"
    else
        echo "  ❌ $name: Not configured"
    fi
done

echo ""
echo "📊 Summary:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Configured Providers: $configured_providers/${#providers[@]}"
echo "  Total Daily Limit: $total_daily_limit requests"

# Check if quota monitor is running
if pgrep -f "quota-monitor" > /dev/null || pgrep -f "hades.*forever" > /dev/null; then
    echo "  Status Monitor: ✅ Running"
else
    echo "  Status Monitor: ❌ Not running"
    echo ""
    echo "💡 To start monitoring:"
    echo "  ./bin/hades forever --monitor 30s --verbose"
fi

# Check for existing quota status file
quota_file="/tmp/hades_quota_status.json"
if [[ -f "$quota_file" ]]; then
    echo ""
    echo "📈 Current Usage (from $quota_file):"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if command -v jq > /dev/null; then
        # Pretty print with jq if available
        jq -r '
        to_entries[] | 
        select(.key | contains("provider")) | 
        "\(.key | split("_")[0] | ascii_upcase): \(.value.current_usage)/\(.value.daily_limit) remaining (\(.value.remaining))"'
        "$quota_file" 2>/dev/null || echo "  ⚠️  Could not parse quota file"
    else
        echo "  Install jq for detailed usage breakdown"
    fi
else
    echo ""
    echo "⚠️  No usage data available yet"
    echo "  Start making requests to see usage statistics"
fi

echo ""
echo "🔧 Configuration Instructions:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Set environment variables for each provider:"
echo ""
echo "  export ANTHROPIC_API_KEY='sk-ant-your-key-here'"
echo "  export GEMINI_API_KEY='AIza-your-key-here'"
echo "  export OPENAI_API_KEY='sk-your-key-here'"
echo "  export GROQ_API_KEY='gsk_your-key-here'"
echo "  export TOGETHER_API_KEY='your-key-here'"
echo "  export HUGGINGFACE_API_KEY='hf_your-key-here'"
echo ""
echo "  Or create a .env file with these variables"
echo ""

echo "🚀 Quick Start Commands:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  # Check current quota status"
echo "  ./scripts/quota-status.sh"
echo ""
echo "  # Start forever API with monitoring"
echo "  ./bin/hades forever --monitor 30s --verbose"
echo ""
echo "  # Run with specific request limit"
echo "  ./bin/hades forever --max 50 --interval 1m"
echo ""

echo "📚 Documentation:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  • API Keys Guide: ./API_KEYS_GUIDE.md"
echo "  • Additional Providers: ./ADDITIONAL_FREE_PROVIDERS.md"
echo "  • Forever API: ./docs/forever-api-documentation.md"
echo ""
