package main

import "github.com/shenyb/solo-workspace/cli/go/cmd"

// version is injected at build time via -ldflags "-X main.version=vX.Y.Z"
var version = "dev"

func main() {
	cmd.Execute(version)
}
