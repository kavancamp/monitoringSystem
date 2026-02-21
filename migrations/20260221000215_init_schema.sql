-- +goose Up
CREATE TABLE devices (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  site TEXT NOT NULL,
  device_type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'OK',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_seen_at TIMESTAMPTZ
);

CREATE INDEX idx_devices_site ON devices(site);
CREATE INDEX idx_devices_status ON devices(status);
CREATE INDEX idx_devices_last_seen ON devices(last_seen_at);

CREATE TABLE readings (
  id BIGSERIAL PRIMARY KEY,
  device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  ts TIMESTAMPTZ NOT NULL DEFAULT now(),
  temperature_c DOUBLE PRECISION,
  pressure_kpa DOUBLE PRECISION,
  rpm DOUBLE PRECISION,
  vibration DOUBLE PRECISION,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_readings_device_ts ON readings(device_id, ts DESC);
CREATE INDEX idx_readings_ts ON readings(ts DESC);

CREATE TABLE alerts (
  id BIGSERIAL PRIMARY KEY,
  device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
  alert_type TEXT NOT NULL,
  severity TEXT NOT NULL,
  message TEXT NOT NULL,
  triggered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  acknowledged_at TIMESTAMPTZ,
  resolved_at TIMESTAMPTZ,
  active BOOLEAN NOT NULL DEFAULT true,
  fingerprint TEXT NOT NULL,
  context JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_alerts_device_active ON alerts(device_id, active);
CREATE INDEX idx_alerts_active_severity ON alerts(active, severity);
CREATE INDEX idx_alerts_triggered_at ON alerts(triggered_at DESC);

CREATE UNIQUE INDEX uniq_active_alert_fingerprint
ON alerts(fingerprint)
WHERE active = true;

-- +goose Down
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS readings;
DROP TABLE IF EXISTS devices;