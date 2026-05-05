#!/bin/bash
# PostgreSQL setup script for Hades SOC V2.0 Docker initialization
# Initializes database with required schemas and security settings

set -e

echo "🔧 Initializing Hades SOC PostgreSQL database..."

# Create additional schemas if needed
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create extensions required by Hades SOC
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
    
    -- Create monitoring user for health checks (drop if exists first)
    DROP USER IF EXISTS hades_monitor;
    CREATE USER hades_monitor WITH PASSWORD 'monitor_password_2024';
    GRANT CONNECT ON DATABASE $POSTGRES_DB TO hades_monitor;
    GRANT USAGE ON SCHEMA public TO hades_monitor;
    GRANT SELECT ON ALL TABLES IN SCHEMA public TO hades_monitor;
    
    -- Set default privileges
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO hades_monitor;
    
    -- Create initial tables for Hades SOC
    CREATE TABLE IF NOT EXISTS system_status (
        id SERIAL PRIMARY KEY,
        timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        component VARCHAR(100) NOT NULL,
        status VARCHAR(20) NOT NULL,
        details JSONB,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );
    
    CREATE TABLE IF NOT EXISTS threat_events (
        id SERIAL PRIMARY KEY,
        event_type VARCHAR(100) NOT NULL,
        severity INTEGER NOT NULL CHECK (severity >= 1 AND severity <= 10),
        source_ip INET,
        target VARCHAR(255),
        description TEXT,
        metadata JSONB,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        resolved_at TIMESTAMP WITH TIME ZONE
    );
    
    CREATE TABLE IF NOT EXISTS agent_metrics (
        id SERIAL PRIMARY KEY,
        agent_name VARCHAR(100) NOT NULL,
        metric_type VARCHAR(50) NOT NULL,
        metric_value NUMERIC,
        timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );
    
    -- Create indexes for performance
    CREATE INDEX IF NOT EXISTS idx_threat_events_created_at ON threat_events(created_at);
    CREATE INDEX IF NOT EXISTS idx_threat_events_severity ON threat_events(severity);
    CREATE INDEX IF NOT EXISTS idx_agent_metrics_timestamp ON agent_metrics(timestamp);
    CREATE INDEX IF NOT EXISTS idx_system_status_timestamp ON system_status(timestamp);
    
    -- Insert initial system status
    INSERT INTO system_status (component, status, details) VALUES 
    ('database', 'initialized', '{"version": "16", "init_time": "'$(date -Iseconds)'"}'),
    ('sentinel', 'starting', '{"status": "initializing"}');
    
    -- Set row level security for threat events
    ALTER TABLE threat_events ENABLE ROW LEVEL SECURITY;
    
    -- Create policy for threat events
    CREATE POLICY threat_events_policy ON threat_events
        FOR ALL TO hades_monitor
        USING (true);
    
    -- Grant permissions to main user
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $POSTGRES_USER;
    
    -- Configure logging
    ALTER SYSTEM SET log_min_duration_statement = 1000;
    ALTER SYSTEM SET log_checkpoints = on;
    ALTER SYSTEM SET log_connections = on;
    ALTER SYSTEM SET log_disconnections = on;
    ALTER SYSTEM SET log_lock_waits = on;
    
    -- Reload configuration
    SELECT pg_reload_conf();
    
    -- Log initialization completion
    INSERT INTO system_status (component, status, details) VALUES 
    ('database', 'ready', '{"init_complete": true, "tables_created": 3}');
EOSQL

echo "✅ PostgreSQL initialization completed successfully"
echo "📊 Database: $POSTGRES_DB"
echo "👤 User: $POSTGRES_USER"
echo "🔐 Extensions: uuid-ossp, pg_stat_statements"
echo "📈 Tables: system_status, threat_events, agent_metrics"
