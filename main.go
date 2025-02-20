/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/hinterland-software/openv/cmd"
	"github.com/hinterland-software/openv/internal/version"
)

func main() {
	cmd.SetVersion(version.Info())
	cmd.Execute()
}
