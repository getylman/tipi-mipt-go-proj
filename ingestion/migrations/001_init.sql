-- Невалидные метрики — пишет Ingestion при провале валидации
CREATE TABLE IF NOT EXISTS invalid_metrics (
    id           SERIAL      PRIMARY KEY,
    raw_payload  JSONB       NOT NULL,
    error_reason VARCHAR(500) NOT NULL,
    received_at  TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invalid_metrics_received_at ON invalid_metrics(received_at);
