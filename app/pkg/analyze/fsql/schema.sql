-- TODO(jeremy): Should we add the notebookURI as a column so we can query for
-- sesions associated with a particular notebook. We could also add the selectedCellId as a column.
-- N.B. Column names should be snake case. This is because sqlc will map snake case to camel case in the generated Go c
-- code. So we will get camelCase member names consistent with GoLang style.
-- See: https://docs.sqlc.dev/en/stable/howto/rename.html
CREATE TABLE IF NOT EXISTS sessions (
    contextID VARCHAR(255) PRIMARY KEY,
    -- protobufs can't have null timestamps so no point allowing nulls
    startTime TIMESTAMP NOT NULL,
    endTime TIMESTAMP NOT NULL,
    -- The selectedId is the id of the cell that is selected in the session.
    selectedID VARCHAR(255) NOT NULL,
    -- SelectedKind is the kind of the selected cell; i.e. Markdown or Code.
    selectedKind VARCHAR(255) NOT NULL,

    -- Total number of input and output tokens
    total_input_tokens INT NOT NULL,
    total_output_tokens INT NOT NULL,

    -- Number of generate traces is the number of generations for this particular session

    num_generate_traces INT NOT NULL,
    -- TODO(jeremy): Should we store the proto in JSON format so that we can run SQL queries on values in it?
    proto BLOB
);
