CREATE TABLE IF NOT EXISTS sessions (
    contextID VARCHAR(255),
    startTime TIMESTAMP,
    endTime TIMESTAMP,
    proto BLOB
);