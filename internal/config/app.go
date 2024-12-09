package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type App struct {
	Name       string
	Server     *Server
	Repository []string
}

func NewAppConfig(file string) (*App, error) {
	var conf App
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
