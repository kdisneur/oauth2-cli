// +build tools

package main

import (
	_ "github.com/spf13/cobra"
	_ "golang.org/x/lint/golint"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
