package testutil

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
)

var (
	BlockComparer = cmpopts.IgnoreUnexported(v1alpha1.Block{}, v1alpha1.BlockOutput{}, v1alpha1.BlockOutputItem{})
)
