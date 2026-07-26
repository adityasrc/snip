
ALTER TABLE links ADD COLUMN click_count INT DEFAULT 0;

CREATE TABLE clicks(
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ DEFAULT NOW(),
    user_agent TEXT,
    referer TEXT,
    ip_address TEXT
);

CREATE INDEX idx_clicks_link_id ON clicks(link_id);

