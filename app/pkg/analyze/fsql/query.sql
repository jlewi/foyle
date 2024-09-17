-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY startTime desc limit 25;

-- name: GetSession :one
SELECT * FROM sessions
WHERE contextID = ?;

-- name: UpdateSession :exec
INSERT OR REPLACE INTO sessions 
(contextID, startTime, endTime, selectedId, selectedKind, proto)
VALUES 
(?, ?, ?, ?, ?, ?);