package main

import (
	"fmt"
	"taws/internal/plugin"
)

type ExamplePlugin struct {
	name    string
	version string
	running bool
}

func (p *ExamplePlugin) Name() string {
	return p.name
}

func (p *ExamplePlugin) Version() string {
	return p.version
}

func (p *ExamplePlugin) Description() string {
	return "An example TAWS plugin for demonstration purposes"
}

func (p *ExamplePlugin) Author() string {
	return "TAWS Team"
}

func (p *ExamplePlugin) Init(configDir string) error {
	fmt.Printf("ExamplePlugin: Initializing with config dir: %s\n", configDir)
	return nil
}

func (p *ExamplePlugin) Start() error {
	p.running = true
	fmt.Println("ExamplePlugin: Started")
	return nil
}

func (p *ExamplePlugin) Stop() error {
	p.running = false
	fmt.Println("ExamplePlugin: Stopped")
	return nil
}

func (p *ExamplePlugin) Status() string {
	if p.running {
		return "running"
	}
	return "stopped"
}

var Plugin plugin.Plugin = &ExamplePlugin{
	name:    "example",
	version: "1.0.0",
}

func main() {}
