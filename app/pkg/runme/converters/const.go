package converters

const (
	// See this thread https://discord.com/channels/1102639988832735374/1218835142962053193/1278863895813165128

	// RunmeIdField is the field name for the ID that Runme uses in memory
	RunmeIdField = "runme.dev/id"
	// IdField is the field name for the ID field when it is serialized
	IdField = "id"

	// GhostKeyField is the field name that indicates whether a cell is a ghost cell or not.
	// Keep it in sync with https://github.com/stateful/vscode-runme/blob/000e08b9523ac264cfda85a8b18427953ef59aac/src/extension/ai/ghost.ts#L19
	GhostKeyField = "_ghostCell"
)
