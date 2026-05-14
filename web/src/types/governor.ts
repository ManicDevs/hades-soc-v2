// Governor Stats Interface - matches Go struct tags exactly (snake_case)
export const GovernorStatsSchema = {
  approved_last_hour: 'number',
  approved_last_24h: 'number', 
  blocked_last_24h: 'number',
  manual_ack_last_24h: 'number',
  total_last_24h: 'number',
  current_block_count: 'number',
  remaining_blocks: 'number',
  max_blocks_per_hour: 'number',
  manual_ack_required: 'boolean',
  time_until_reset: 'string',
  time_until_reset_seconds: 'number',
  database_error: 'boolean?',
  error_message: 'string?'
}

// Governor Status Interface
export const GovernorStatusSchema = {
  current_block_count: 'number',
  remaining_blocks: 'number',
  max_blocks_per_hour: 'number',
  manual_ack_required: 'boolean',
  database_error: 'boolean?',
  error_message: 'string?'
}

// Governor Action Interface
export const GovernorActionSchema = {
  id: 'number',
  action_id: 'string',
  action_name: 'string',
  target: 'string',
  reasoning: 'string?',
  requester: 'string',
  status: 'string',
  requires_approval: 'boolean',
  approved: 'boolean',
  requires_manual_ack: 'boolean',
  block_reason: 'string?',
  execution_time: 'number',
  metadata: 'object?',
  created_at: 'string',
  updated_at: 'string'
}

// WebSocket GOVERNOR_INTERCEPT Message Interface
export const GovernorInterceptSchema = {
  type: 'GOVERNOR_INTERCEPT',
  data: {
    action_name: 'string',
    target: 'string',
    reasoning: 'string',
    requires_approval: 'boolean',
    requester: 'string',
    timestamp: 'number',
    block_reason: 'string',
    source: 'string'
  },
  timestamp: 'string'
}

// Runtime validation function for governor stats
export const validateGovernorStats = (data) => {
  const required = ['approved_last_hour', 'approved_last_24h', 'blocked_last_24h', 'manual_ack_last_24h', 'total_last_24h']
  const missing = required.filter(key => !(key in data))
  
  if (missing.length > 0) {
    throw new Error(`Missing required governor stats fields: ${missing.join(', ')}`)
  }
  
  // Validate types
  if (typeof data.approved_last_hour !== 'number') {
    throw new Error('approved_last_hour must be a number')
  }
  
  if (typeof data.current_block_count !== 'number') {
    throw new Error('current_block_count must be a number')
  }
  
  return true
}

// Runtime validation function for governor intercept messages
export const validateGovernorIntercept = (message) => {
  if (!message.type || message.type !== 'GOVERNOR_INTERCEPT') {
    throw new Error('Invalid message type, expected GOVERNOR_INTERCEPT')
  }
  
  if (!message.data || typeof message.data !== 'object') {
    throw new Error('Message data is required and must be an object')
  }
  
  const required = ['action_name', 'target', 'requires_approval', 'requester', 'timestamp']
  const missing = required.filter(key => !(key in message.data))
  
  if (missing.length > 0) {
    throw new Error(`Missing required intercept fields: ${missing.join(', ')}`)
  }
  
  return true
}
