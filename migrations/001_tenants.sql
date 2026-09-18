-- ============================================================
--  HORIZON MULTI-TENANT MIGRATION
--  اضافه کردن tenant_id به همه جداول + جدول tenants
-- ============================================================

-- 1) جدول tenants
CREATE TABLE IF NOT EXISTS tenants (
    id              SERIAL PRIMARY KEY,
    tenant_id       VARCHAR(64) UNIQUE NOT NULL,
    name            VARCHAR(255) NOT NULL,
    type            VARCHAR(32) NOT NULL, -- bank | refinery | powerplant | gas | other
    contact_email   VARCHAR(255),
    contact_phone   VARCHAR(64),
    deployment_mode VARCHAR(32) DEFAULT 'cloud', -- cloud | on-premise | hybrid
    agent_url       VARCHAR(255), -- آدرس سوئیچ مشتری (اگر on-premise)
    agent_token     TEXT,         -- توکن احراز هویت agent
    last_heartbeat  TIMESTAMP,
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenants_tenant_id ON tenants(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenants_type ON tenants(type);

-- 2) اضافه کردن tenant_id به جداول موجود (SQLite: قابل انجام با ALTER TABLE)
-- نکته: در SQLite نمی‌توان ستون با constraint اضافه کرد، پس جداگانه در کد اضافه می‌شود.
