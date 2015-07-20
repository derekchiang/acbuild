package main

var (
	flags = struct {
		Input           string
		Output          string
		Context         bool
		OutputImageName string
		Cmd             string
		NoOverlay       bool
		Split           bool
		Mount           []string
		AllButLast      bool
		Overwrite       bool
	}{}
)
