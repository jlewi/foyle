package analyze

// LogEntry represents a log entry.
// Any fields you want to decode from the log entry should be added here.
// Per https://stackoverflow.com/questions/33436730/unmarshal-json-with-some-known-and-some-unknown-field-names if
// we wanted to automatically decode unknown fields to map[string]interface{} we'd probably need to do this in two
// passes.
type LogEntry struct {
	Severity string   `json:"severity"`
	Time     float64  `json:"time"`
	Message  string   `json:"message"`
	TraceID  string   `json:"traceId"`
	BlockIds []string `json:"blockIds"`
}
