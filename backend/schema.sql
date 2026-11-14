CREATE TABLE IF NOT EXISTS funnels (
    slug VARCHAR(100) PRIMARY KEY,
    product_name VARCHAR(255) NOT NULL,
    target_audience TEXT NOT NULL,
    auto_winner BOOLEAN DEFAULT FALSE,
    variants_json JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS funnel_events (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(100) NOT NULL,
    variant CHAR(1) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    utm_source VARCHAR(100) DEFAULT 'direct',
    revenue_cents BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);