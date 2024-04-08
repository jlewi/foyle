package docs

const (
	BASHLANG = "bash"
	// OUTPUTLANG is the language to give to output code blocks.
	// We want to potentially distinguish output from code blocks because output blocks are nested inside blocks
	// in notebooks. Therefore if we want to be able to convert a markdown document into a document with blocks
	// then having a unique language for output blocks helps us identify them and properly reencode them.
	OUTPUTLANG = "output"
)
