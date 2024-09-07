-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY startTime;