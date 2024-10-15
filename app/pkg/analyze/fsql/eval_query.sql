-- name: UpdateResult :exec
INSERT OR REPLACE INTO results
(id, time, proto_json)
VALUES
(?, ?, ?);

-- name: GetResult :one
SELECT * FROM results
WHERE id = ?;


-- name: ListResults :many
-- This queries for results.
-- Results are listed in descending order of time (most recent first) because the primary use is for resuming
-- in the evaluator
SELECT * FROM results
WHERE (:cursor = '' OR time < :cursor)
ORDER BY time DESC
    LIMIT :page_size;

-- name: CountResults :one
-- Count the total number of results
SELECT COUNT(*) FROM results;

-- name: CountErrors :one
SELECT COUNT(*) FROM results WHERE json_extract(proto_json, '$.error') IS NOT NULL;

-- name: CountByCellsMatchResult :many
SELECT json_extract(proto_json, '$.cellsMatchResult') as match_result, COUNT(*) as count FROM results GROUP BY match_result;