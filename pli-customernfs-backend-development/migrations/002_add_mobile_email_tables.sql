-- Migration 002: Add mobile_change_detail, email_change_detail and version history tables
-- For WF-NFS-006 (Mobile Change) and WF-NFS-007 (Email Change)

-- Policy-customer mapping for PM notification lookups
CREATE TABLE IF NOT EXISTS nfs.policy_customer_mapping (
    policy_number   VARCHAR(30) NOT NULL,
    customer_id     BIGINT      NOT NULL,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (policy_number, customer_id)
);

CREATE INDEX IF NOT EXISTS idx_pcm_customer ON nfs.policy_customer_mapping (customer_id, is_active);

-- Mobile change detail
CREATE TABLE IF NOT EXISTS nfs.mobile_change_detail (
    detail_id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id          UUID        NOT NULL REFERENCES nfs.service_request(request_id),
    old_mobile_number   VARCHAR(15),
    new_mobile_number   VARCHAR(15) NOT NULL,
    aadhaar_txn_id      VARCHAR(100),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by          VARCHAR(100) NOT NULL,
    updated_by          VARCHAR(100),
    version             INTEGER     NOT NULL DEFAULT 1
);

-- Mobile version history
CREATE TABLE IF NOT EXISTS nfs.mobile_version_history (
    version_id      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id     BIGINT      NOT NULL,
    request_id      UUID        REFERENCES nfs.service_request(request_id),
    mobile_number   VARCHAR(15) NOT NULL,
    version_number  INTEGER     NOT NULL DEFAULT 1,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    effective_from  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      VARCHAR(100) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_mobile_version_customer ON nfs.mobile_version_history (customer_id);
CREATE INDEX IF NOT EXISTS idx_mobile_version_active   ON nfs.mobile_version_history (customer_id, is_active);

-- Email change detail
CREATE TABLE IF NOT EXISTS nfs.email_change_detail (
    detail_id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id      UUID        NOT NULL REFERENCES nfs.service_request(request_id),
    old_email       VARCHAR(255),
    new_email       VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      VARCHAR(100) NOT NULL,
    updated_by      VARCHAR(100),
    version         INTEGER     NOT NULL DEFAULT 1
);

-- Email version history
CREATE TABLE IF NOT EXISTS nfs.email_version_history (
    version_id      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id     BIGINT      NOT NULL,
    request_id      UUID        REFERENCES nfs.service_request(request_id),
    email           VARCHAR(255) NOT NULL,
    version_number  INTEGER     NOT NULL DEFAULT 1,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    effective_from  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by      VARCHAR(100) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_email_version_customer ON nfs.email_version_history (customer_id);
CREATE INDEX IF NOT EXISTS idx_email_version_active   ON nfs.email_version_history (customer_id, is_active);
