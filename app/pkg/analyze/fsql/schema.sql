CREATE TABLE IF NOT EXISTS sessions (
    contextID VARCHAR(255) PRIMARY KEY,
    -- protobufs can't have null timestamps so no point allowing nulls
    startTime TIMESTAMP NOT NULL,
    endTime TIMESTAMP NOT NULL,
    proto BLOB
);
