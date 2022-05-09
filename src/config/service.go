package config

import (
	"os"
)

type serviceConfig struct {
	Port string
}

var service serviceConfig

func Service(reload ...bool) *serviceConfig {
	if (serviceConfig{}) != service && !reload[0] {
		return &service
	}
	service = serviceConfig{
		Port: os.Getenv("PORT"),
	}
	return &service
}
