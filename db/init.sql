CREATE TABLE account
(
    id         SERIAL PRIMARY KEY,                    -- Auto-incrementing internal DB ID
    account_id VARCHAR(64) NOT NULL UNIQUE,           -- App-level public ID (e.g., 'acct_xxx'), must be unique
    name       TEXT        NOT NULL,                  -- Display name (e.g., "Main Wallet")
    type       VARCHAR(20) NOT NULL DEFAULT 'WALLET', -- Account type: WALLET, CASA
    currency   CHAR(3)     NOT NULL DEFAULT 'MYR',    -- ISO 4217 currency code (USD, MYR, etc.)
    balance    BIGINT      NOT NULL DEFAULT 0,        -- Current balance in minor units (e.g., cents)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),    -- Creation time
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()     -- Last updated time
);

CREATE TABLE transfer
(
    id                        BIGSERIAL PRIMARY KEY,              -- Auto-incrementing internal DB ID
    type                      VARCHAR(36)  NOT NULL DEFAULT '',   -- Type of the transfer
    tx_type                   VARCHAR(36)  NOT NULL DEFAULT '',   -- Purpose type
    user_id                   VARCHAR(36),                        -- Customer who initiates the transfer
    transaction_id            VARCHAR(36)  NOT NULL DEFAULT '',   -- Internal transaction tracking ID
    reference_id              VARCHAR(36)  NOT NULL DEFAULT '',   -- Idempotency key
    status                    VARCHAR(36)  NOT NULL DEFAULT '',   -- Status of the transaction
    amount                    BIGINT       NOT NULL,              -- Amount in minor unit
    currency                  VARCHAR(3)   NOT NULL DEFAULT '',   -- ISO currency code
    source_account_id         VARCHAR(36),                        -- Source account ID
    source_account            JSONB        NOT NULL DEFAULT '{}', -- Source account details
    destination_account_id    VARCHAR(36),                        -- Destination account ID
    destination_account       JSONB        NOT NULL DEFAULT '{}', -- Destination account details
    status_reason             VARCHAR(255)          DEFAULT '',   -- Status reason code
    status_reason_description VARCHAR(500)          DEFAULT '',   -- Status reason description
    note                      VARCHAR(255) NOT NULL DEFAULT '',   -- Transfer note/remark
    properties                JSONB                 DEFAULT '{}', -- Metadata per transfer type
    created_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    valued_at                 TIMESTAMPTZ,
    updated_at                TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_transaction_id UNIQUE (transaction_id),
    CONSTRAINT uk_reference_id UNIQUE (reference_id)
);

CREATE TABLE transaction
(
    id         SERIAL PRIMARY KEY,                                       -- Auto-incrementing ID
    account_id VARCHAR(64) NOT NULL,                                     -- Account this transaction belongs to
    type       VARCHAR(10) NOT NULL CHECK (type IN ('credit', 'debit')), -- 'credit' or 'debit'
    amount     BIGINT      NOT NULL CHECK (amount >= 0),                 -- Minor units (e.g., cents)
    currency   CHAR(3)     NOT NULL DEFAULT 'MYR',                       -- ISO 4217 currency code
    timestamp  TIMESTAMPTZ NOT NULL DEFAULT NOW(),                       -- When transaction occurred
    valued_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),                       -- When it takes effect in balance
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),                       -- Last update time
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),                       -- Insert time
    note       TEXT,                                                     -- Optional description
    properties JSONB                DEFAULT '{}'                         -- Metadata, tags, channel info, etc.
);

CREATE
OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at
= CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE
    ON account
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();