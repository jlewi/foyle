-- name: ListSessions :many
-- TODO(jeremy): How would we paginate this? Since contextIds are ULIDs we could rely on them being lexicographically
-- ordered and use the last contextId as the start of the next page.
SELECT * FROM sessions
ORDER BY startTime desc limit 25;

-- name: ListSessionsForExamples :many
-- This queries for sessions that we will use for Examples.
--- This should be sessions that have a selectedKind of Code
SELECT * FROM sessions
WHERE (:cursor = '' OR contextId < :cursor) and selectedKind = 'CELL_KIND_CODE'
ORDER BY contextId DESC
    LIMIT :page_size;

-- name: GetSession :one
SELECT * FROM sessions
WHERE contextID = ?;

-- name: UpdateSession :exec
INSERT OR REPLACE INTO sessions 
(contextID, startTime, endTime, selectedId, selectedKind, proto)
VALUES 
(?, ?, ?, ?, ?, ?);