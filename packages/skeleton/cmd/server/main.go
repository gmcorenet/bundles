package main

import (
	"log"
	"os"

	"github.com/gmcorenet/framework/internal/config"
	"github.com/gmcorenet/framework/internal/container"
	"github.com/gmcorenet/framework/internal/kernel"
	"github.com/gmcorenet/framework/internal/router"
	"github.com/gmcorenet/framework/pkg"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "resources/config"
	}

	cfg := config.GetInstance().Load(configPath)

	c := container.New()
	c.Set("Config", cfg)

	r := router.New()
	registerRoutes(r)

	app := kernel.New(r, c)

	request := pkg.NewRequest()
	response := app.Handle(request)

	log.Printf("Response: %s (status: %d)", response.Body(), response.StatusCode())
}

func registerRoutes(r *router.Router) {
	r.Get("/", func(req *pkg.Request) *pkg.Response {
		return pkg.NewResponse("Hello, GMCore!", 200)
	})

	r.Get("/health", func(req *pkg.Request) *pkg.Response {
		return pkg.NewResponse(`{"status":"ok"}`, 200).WithHeader("Content-Type", "application/json")
	})
}