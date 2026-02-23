-- name: CreateAlert :one
INSERT INTO alerts (device_id, alert_type, severity, message, fingerprint, context)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ResolveAlertByFingerprint :exec
UPDATE alerts
SET active = false, resolved_at = now()
WHERE fingerprint = $1 AND active = true;

-- name: AckAlert :exec
UPDATE alerts
SET acknowledged_at = now()
WHERE id = $1 AND acknowledged_at IS NULL;

-- name: ListAlerts :many
SELECT * FROM alerts
WHERE ($1::bool IS NULL OR active = $1)
  AND ($2::text IS NULL OR severity = $2)
  AND ($3::uuid IS NULL OR device_id = $3)
ORDER BY triggered_at DESC
LIMIT $4 OFFSET $5;