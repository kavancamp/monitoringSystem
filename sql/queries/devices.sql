-- name: CreateDevice :one
INSERT INTO devices (id, name, site, device_type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDevice :one
SELECT * FROM devices WHERE id = $1;

-- name: ListDevices :many
SELECT * FROM devices
WHERE (sqlc.narg(site)::text IS NULL OR site = sqlc.narg(site))
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status))
ORDER BY created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: UpdateDeviceLastSeen :exec
UPDATE devices SET last_seen_at = $2 WHERE id = $1;

-- name: SetDeviceStatus :exec
UPDATE devices SET status = $2 WHERE id = $1;

-- name: ListStaleDevices :many
SELECT * FROM devices
WHERE last_seen_at IS NULL OR last_seen_at < $1;