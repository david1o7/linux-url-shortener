
CREATE TABLE IF NOT EXISTS urls (
    id            BIGSERIAL PRIMARY KEY,
    originalurl   TEXT        NOT NULL,
    shortcode     VARCHAR(16) UNIQUE NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed TIMESTAMPTZ,
    clicks        BIGINT      NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_urls_shortcode ON urls (shortcode);

CREATE INDEX IF NOT EXISTS idx_urls_originalurl ON urls (originalurl);