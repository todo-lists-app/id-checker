// Package main run the app
package main

import (
	"fmt"

	gc "github.com/keloran/go-config"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/todo-lists-app/id-checker/internal/service"
)

var (
	// BuildVersion is the version of the app
	BuildVersion = "dev"
	// BuildHash is the git-hash of the app
	BuildHash = "unknown"
	// ServiceName is the name of the service
	ServiceName = "base-service"
)

func main() {
	logs.Local().Info(fmt.Sprintf("Starting %s", ServiceName))
	logs.Local().Info(fmt.Sprintf("Version: %s, Hash: %s", BuildVersion, BuildHash))

	cfg, err := gc.Build(gc.Keycloak, gc.Local)
	if err != nil {
		_ = logs.Errorf("config: %v", err)
		return
	}

	s := &service.Service{
		Config: cfg,
	}

	if err := s.Start(); err != nil {
		_ = logs.Errorf("start service: %v", err)
		return
	}
}
