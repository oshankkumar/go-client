package goclient

import "github.com/afex/hystrix-go/hystrix"

type HystrixConfig struct {
	hystrix.CommandConfig
	Name string
}

func NewHystrixConfig(commandName string, config hystrix.CommandConfig) *HystrixConfig {
	return &HystrixConfig{Name: commandName, CommandConfig: config}
}
