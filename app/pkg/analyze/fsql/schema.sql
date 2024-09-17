-- TODO(jeremy): Should we add the notebookURI as a column so we can query for
-- sesions associated with a particular notebook. We could also add the selectedCellId as a column.
CREATE TABLE IF NOT EXISTS sessions (
    contextID VARCHAR(255) PRIMARY KEY,
    -- protobufs can't have null timestamps so no point allowing nulls
    startTime TIMESTAMP NOT NULL,
    endTime TIMESTAMP NOT NULL,
    -- The selectedId is the id of the cell that is selected in the session.
    selectedID VARCHAR(255) NOT NULL,
    -- SelectedKind is the kind of the selected cell; i.e. Markdown or Code.
    selectedKind VARCHAR(255) NOT NULL,
    proto BLOB
);
