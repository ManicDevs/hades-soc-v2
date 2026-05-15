# Forever API - Free Tier Maximizer Documentation

## Overview

The Forever API script is a sophisticated tool designed to maximize free tier usage across multiple AI API providers. It automatically rotates through providers based on availability, quota limits, and performance metrics.

## Features

### 🚀 Core Features
- **Multi-Provider Support**: Supports 10+ API providers (Anthropic, OpenAI, Gemini, Groq, Together AI, Hugging Face, Replicate, Cohere, Perplexity, Mistral, AI21)
- **Intelligent Rotation**: Automatically switches providers based on quota exhaustion, error rates, and response times
- **Adaptive Rate Limiting**: Dynamically adjusts request intervals based on provider performance
- **Real-time Monitoring**: Tracks success rates, response times, and provider health
- **Auto-Recovery**: Automatically restarts failed processes and recovers from errors
- **Log Management**: Automatic log rotation with configurable backup retention

### 📊 Advanced Monitoring
- **Live Dashboard**: Real-time statistics display with auto-refresh
- **Performance Metrics**: Tracks average response times and provider performance
- **Usage Analytics**: Detailed request counts and success rates per provider
- **Health Monitoring**: Continuous provider health checks with configurable intervals

### 🛡️ Safety & Reliability
- **PID Management**: Prevents multiple instances and ensures clean shutdown
- **Graceful Shutdown**: Proper cleanup of resources and background processes
- **Error Handling**: Comprehensive error detection with automatic recovery
- **Signal Handling**: Responds to SIGINT/SIGTERM for graceful termination

## Provider Configuration

### Unlimited Providers (Priority 1-5)
| Provider | Model | Rate Limit | Daily Limit | Priority |
|-----------|--------|-------------|-------------|----------|
| Groq | llama3-70b-8192 | 30/min | 1 (Highest) |
| Together AI | Meta-Llama-3.1-8B | 60/min | 2 |
| Hugging Face | Meta-Llama-3.1-8B | 300/hour | 3 |
| Replicate | meta-llama-3-70b | 100/min | 4 |
| Cohere | command | 100/min | 5 |

### Limited Providers (Priority 6-8)
| Provider | Model | Daily Limit | Priority |
|-----------|--------|-------------|----------|
| Anthropic | claude-3-5-sonnet | 1,000/day | 6 |
| OpenAI | gpt-4 | 100/day | 7 |
| Gemini | gemini-1.5-flash | 20/day | 8 (Lowest) |

### Additional Providers
| Provider | Model | Daily Limit | Priority |
|-----------|--------|-------------|----------|
| Perplexity | llama-3-sonar-small | ~167/day | 9 |
| Mistral | mistral-large | ~33/day | 10 |
| AI21 | j2-ultra | ~33/day | 11 |

## Installation & Setup

### Prerequisites
- Go 1.21+ installed
- API keys configured in `.env` file
- Bash 4.0+ for advanced features

### Environment Variables
```bash
# Core Providers
ANTHROPIC_API_KEY=sk-ant-api03-...
GEMINI_API_KEY=AIzaSy...
OPENAI_API_KEY=sk-proj-...

# Unlimited Providers
GROQ_API_KEY=gsk_cRNG...
TOGETHER_API_KEY=tgp_v1_...
HUGGINGFACE_API_KEY=hf_JUUpA...
REPLICATE_API_KEY=r8_UN5u...
COHERE_API_KEY=cohere_ZLYJ...

# Optional Providers
PERPLEXITY_API_KEY=your_perplexity_api_key_here
MISTRAL_API_KEY=your_mistral_api_key_here
AI21_API_KEY=your_ai21_api_key_here
```

## Usage

### Basic Usage
```bash
# Run with default settings (30s intervals)
./scripts/forever-api.sh

# Custom interval and verbose logging
./scripts/forever-api.sh -i 10s -v

# Limited requests with custom prompts
./scripts/forever-api.sh -m 100 -i 1m -p my_prompts.txt

# Enable logging
./scripts/forever-api.sh -l api.log -v
```

### Advanced Options
```bash
# Real-time dashboard mode
./scripts/forever-api.sh --dashboard

# Disable log rotation
./scripts/forever-api.sh --no-log-rotation

# Custom health check interval
./scripts/forever-api.sh --health-check 1m

# Disable auto-recovery
./scripts/forever-api.sh --no-auto-recovery

# Fixed rate limiting (no adaptive)
./scripts/forever-api.sh --fixed-rate
```

### Command Line Options
| Option | Description | Default |
|--------|-------------|---------|
| `-i, --interval TIME` | Request interval | 30s |
| `-m, --max NUM` | Maximum requests (0=infinite) | 0 |
| `-p, --prompts FILE` | Custom prompts file | - |
| `-l, --log FILE` | Log to file | stdout |
| `-v, --verbose` | Verbose logging | false |
| `--dashboard` | Show real-time dashboard | false |
| `--no-log-rotation` | Disable log rotation | false |
| `--health-check TIME` | Health check interval | 5m |
| `--no-auto-recovery` | Disable auto-recovery | false |
| `--fixed-rate` | Disable adaptive rate limiting | false |
| `-h, --help` | Show help | - |

## Monitoring & Dashboard

### Real-time Dashboard Features
- **Live Statistics**: Request count, success rate, error count
- **Provider Usage**: Per-provider request distribution
- **Performance Metrics**: Average response times and current provider
- **System Status**: Uptime, health checks, and process information
- **Auto-refresh**: Updates every 10 seconds

### Metrics Files
- `.forever-api.metrics`: JSON-formatted real-time metrics
- `.forever-api.stats`: Final statistics summary
- `logs/forever-api.log`: Main application log
- `logs/forever-api.1.log`, `.2.log`, etc.: Rotated logs

### Performance Tracking
```json
{
  "timestamp": 1715678900,
  "request_count": 150,
  "success_count": 142,
  "error_count": 8,
  "success_rate": 94.67,
  "current_provider": "Groq",
  "average_response_time": 1250,
  "uptime_seconds": 3600,
  "providers": {
    "Groq_requests": 45,
    "Together_requests": 38,
    "Anthropic_requests": 25,
    "OpenAI_requests": 20,
    "Gemini_requests": 12,
    "HuggingFace_requests": 8,
    "Replicate_requests": 2
  }
}
```

## Operation Modes

### Standard Mode
- Starts API process with monitoring
- Handles provider rotation automatically
- Performs health checks at configured intervals
- Logs all activities and errors

### Dashboard Mode
- Displays real-time statistics and metrics
- Auto-refreshes every 10 seconds
- Shows provider performance and usage distribution
- Can be exited with Ctrl+C

## Troubleshooting

### Common Issues

#### Process Already Running
```bash
Error: Forever API is already running (PID: 12345)
Solution: kill 12345 or ./scripts/forever-api.sh --stop
```

#### Missing API Keys
```bash
Warning: Missing API keys: Perplexity, Mistral, AI21
Solution: Add keys to .env file or environment variables
```

#### Build Failures
```bash
Error: Build failed!
Solution: Check Go installation and dependencies
```

#### Permission Issues
```bash
Error: Permission denied
Solution: chmod +x scripts/forever-api.sh
```

### Debug Mode
Enable verbose logging for detailed troubleshooting:
```bash
./scripts/forever-api.sh -v -l debug.log
```

### Log Analysis
```bash
# View recent errors
grep -i error logs/forever-api.log

# Monitor provider switches
grep "Switched providers" logs/forever-api.log

# Check performance metrics
tail -f .forever-api.metrics
```

## Architecture

### Component Overview
```
┌─────────────────────────────────────────┐
│         forever-api.sh (Launcher)    │
├─────────────────────────────────────────┤
│         bin/forever-api (Go Binary) │
├─────────────────────────────────────────┤
│  internal/api/ (Core Logic)        │
│  ├── quota_orchestrator.go         │
│  ├── client.go                   │
│  └── unlimited_providers.go      │
├─────────────────────────────────────────┤
│  .env (Configuration)               │
│  logs/ (Output)                   │
└─────────────────────────────────────────┘
```

### Data Flow
1. **Shell Script** parses arguments and validates environment
2. **Go Binary** initializes providers and manages rotation
3. **API Client** handles HTTP requests and rate limiting
4. **Quota Orchestrator** tracks usage and switches providers
5. **Monitoring System** logs metrics and provides real-time dashboard

### Provider Selection Algorithm
1. Check if current provider has available quota
2. Calculate provider scores based on:
   - Remaining quota percentage
   - Error count and health status
   - Response time performance
   - Priority weighting
3. Select highest-scoring available provider
4. Switch provider if different from current
5. Log rotation and update statistics

## Security Considerations

### API Key Protection
- Keys loaded from `.env` file with proper permissions
- No keys logged or exposed in output
- Secure file permissions recommended (600)

### Process Isolation
- PID file prevents multiple instances
- Graceful shutdown prevents resource leaks
- Background processes properly managed

### Rate Limiting
- Respects provider rate limits
- Adaptive throttling based on response times
- Automatic backoff on errors

## Performance Optimization

### Request Efficiency
- Intelligent provider selection minimizes latency
- Connection pooling for repeated requests
- Adaptive intervals based on success rates

### Resource Management
- Minimal memory footprint
- Efficient log rotation
- Background monitoring with low overhead

## Integration Examples

### With Hades V2
```bash
# Integrate with Hades monitoring
./scripts/forever-api.sh --monitor-interval 30s --hades-integration

# Export metrics to Hades database
./scripts/forever-api.sh --export-metrics --hades-endpoint http://localhost:8443
```

### Custom Workflows
```bash
# Development mode with fast requests
./scripts/forever-api.sh -i 2s -m 50 -v --dashboard

# Production mode with logging
./scripts/forever-api.sh -i 30s -l production.log --no-auto-recovery

# Testing with specific providers
./scripts/forever-api.sh --providers-only "Groq,Together" --test-mode
```

## Best Practices

### Production Deployment
1. Use log rotation for long-running processes
2. Enable auto-recovery for reliability
3. Monitor dashboard for performance insights
4. Set appropriate intervals for provider limits
5. Regular backup of metrics files

### Development
1. Use verbose mode for debugging
2. Test with small request counts first
3. Monitor provider rotation behavior
4. Validate API key configuration
5. Check performance metrics regularly

### Maintenance
1. Regular cleanup of old log files
2. Monitor provider API changes and limits
3. Update script for new providers
4. Backup configuration files
5. Review performance metrics periodically

## FAQ

### Q: How are providers prioritized?
A: Unlimited providers (Groq, Together AI, etc.) are prioritized highest (1-5), followed by limited providers based on quota size.

### Q: What happens when all providers are exhausted?
A: The system waits for cooldown periods and automatically retries when providers become available.

### Q: How accurate are the quota limits?
A: Limits are based on provider documentation and real-world testing. Actual limits may vary.

### Q: Can I add custom providers?
A: Yes, by modifying the provider configuration in the Go source code.

### Q: How do I monitor costs?
A: The script tracks request counts per provider. Multiply by provider pricing for cost estimation.

## Support & Contributing

### Getting Help
```bash
./scripts/forever-api.sh --help
```

### Reporting Issues
Include:
- OS and version information
- Error messages and logs
- Provider configuration
- Steps to reproduce

### Contributing
1. Fork the repository
2. Test changes thoroughly
3. Update documentation
4. Submit pull requests
5. Follow coding standards

---

*Last Updated: 2026-05-13*
*Version: 2.0 Enhanced*
*Documentation Version: 1.0*
