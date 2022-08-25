//go:build tools
// +build tools

package tools

import (
	// linter specifically for TF plugins
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlint"
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlintx"
)
