package main

var (
	flags = struct {
		Input           string
		Output          string
		Context         bool
		OutputImageName string
		Cmd             string
		NoOverlay       bool
		Mount           []string
		Overwrite       bool
	}{}
)
