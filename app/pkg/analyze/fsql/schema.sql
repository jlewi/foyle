CREATE TABLE IF NOT EXISTS sessions (
    contextID VARCHAR(255) PRIMARY KEY,
    startTime TIMESTAMP,
    endTime TIMESTAMP,
    proto BLOB
);
