package eval

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

// HashExample returns a hash of the example.
// This is intended to be a stable ID used to identify the example across different evaluation datasets.
func HashExample(e *v1alpha1.Example) string {
	// I don't think there is a guarantee that the wire format of protos is deterministic.
	// So we explicitly hash the contents of the blocks to ensure we get a deterministic hash.
	hasher := sha256.New()

	for _, b := range e.Query.Blocks {
		hasher.Write([]byte(b.GetContents()))
	}

	for _, b := range e.Answer {
		hasher.Write([]byte(b.GetContents()))
	}
	hash := hasher.Sum(nil)

	return base64.StdEncoding.EncodeToString(hash)
}
