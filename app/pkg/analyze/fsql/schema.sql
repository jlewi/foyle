-- TODO(jeremy): Should we add the notebookURI as a column so we can query for
-- sesions associated with a particular notebook. We could also add the selectedCellId as a column.
CREATE TABLE IF NOT EXISTS sessions (
    contextID VARCHAR(255) PRIMARY KEY,
    -- protobufs can't have null timestamps so no point allowing nulls
    startTime TIMESTAMP NOT NULL,
    endTime TIMESTAMP NOT NULL,
    proto BLOB
);
