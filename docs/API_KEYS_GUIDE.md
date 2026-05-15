# 🔑 API Keys Acquisition Guide

## 📋 Where to Get Free Tier API Keys

### 1. Anthropic Claude (1000 requests/day free)
- **Website**: https://console.anthropic.com/
- **Steps**:
  1. Sign up for free account
  2. Go to API Keys section
  3. Create new API key
  4. Copy key (starts with `sk-ant-`)
- **Free Tier**: 1000 requests per day
- **Documentation**: https://docs.anthropic.com/

### 2. Google Gemini (20 requests/day free)
- **Website**: https://aistudio.google.com/app/apikey
- **Steps**:
  1. Sign in with Google account
  2. Go to "API Keys" section
  3. Create new API key
  4. Copy key (starts with `AIza`)
- **Free Tier**: 20 requests per day
- **Documentation**: https://ai.google.dev/docs

### 3. OpenAI GPT-4 (100 requests/day free)
- **Website**: https://platform.openai.com/api-keys
- **Steps**:
  1. Sign up for free account
  2. Go to API Keys section
  3. Create new API key
  4. Copy key (starts with `sk-`)
- **Free Tier**: $5 free credit (~100 GPT-4 requests)
- **Documentation**: https://platform.openai.com/docs

## 🛡️ Security Best Practices

### Environment Setup
```bash
# Method 1: Environment variables
export ANTHROPIC_API_KEY='sk-ant-your-key-here'
export GEMINI_API_KEY='AIza-your-key-here'
export OPENAI_API_KEY='sk-your-key-here'

# Method 2: .env file
echo "ANTHROPIC_API_KEY=sk-ant-your-key-here" >> .env
echo "GEMINI_API_KEY=AIza-your-key-here" >> .env
echo "OPENAI_API_KEY=sk-your-key-here" >> .env
```

### Security Rules
- ✅ **DO**: Store keys in environment variables or .env files
- ✅ **DO**: Use separate keys for development/production
- ✅ **DO**: Rotate keys regularly
- ❌ **DON'T**: Commit keys to git repositories
- ❌ **DON'T**: Share keys publicly
- ❌ **DON'T**: Log keys in application output

## 🚀 Quick Start

1. **Get your keys** using the links above
2. **Configure them**:
   ```bash
   export ANTHROPIC_API_KEY='your-key-here'
   export GEMINI_API_KEY='your-key-here'
   export OPENAI_API_KEY='your-key-here'
   ```
3. **Test the system**:
   ```bash
   ./scripts/verify-forever-api.sh
   ```
4. **Run forever**:
   ```bash
   ./scripts/forever-api.sh -m 10 -i 30s -v
   ```

## 📊 Daily Quota Summary

| Provider | Free Tier | Requests/Day | Key Prefix |
|-----------|------------|---------------|-------------|
| Anthropic | Claude | 1,000 | `sk-ant-` |
| Google | Gemini | 20 | `AIza` |
| OpenAI | GPT-4 | ~100 | `sk-` |
| **Total** | **1,120** | **per day** |

## 🔍 Verification

After setting up keys, run:
```bash
# Verify system works
./scripts/verify-forever-api.sh

# Check quota status
./scripts/api-quota-check.sh

# Start maximizing free tier
./scripts/forever-api.sh
```

## ⚠️ Important Notes

- Free tiers have usage limits and may rate limit
- Keys are tied to your account - keep them secure
- Some providers require credit card for verification (but won't charge)
- Quota resets at midnight UTC for most providers
- Monitor usage to avoid unexpected charges
