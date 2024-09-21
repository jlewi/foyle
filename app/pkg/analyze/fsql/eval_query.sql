-- name: UpdateResult :exec
INSERT OR REPLACE INTO results
(id, time, proto_json)
VALUES
(?, ?, ?);

-- name: GetResult :one
SELECT * FROM results
WHERE id = ?;