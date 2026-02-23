-- name: InsertReading :one
INSERT INTO readings (device_id, ts, temperature_c, pressure_kpa, rpm, vibration, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListReadingsForDevice :many
SELECT * FROM readings
WHERE device_id = $1 AND ts >= $2
ORDER BY ts DESC
LIMIT $3;