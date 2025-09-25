package main

// This is a http server that serves the broker API. It is used in the integration tests.

import (
	"log/slog"
	"net/http"
	"os"

	"code.cloudfoundry.org/brokerapi/v13"
	"code.cloudfoundry.org/brokerapi/v13/fakes"
	"code.cloudfoundry.org/brokerapi/v13/handlers"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("Starting fake broker api server")

	serviceBroker := &fakes.FakeServiceBroker{
		ProvisionedInstances: make(map[string]brokerapi.ProvisionDetails),
		BoundBindings:        make(map[string]brokerapi.BindDetails),
	}

	apiHandler := handlers.NewApiHandler(serviceBroker, logger)
	r := http.NewServeMux()
	r.HandleFunc("GET /v2/catalog", apiHandler.Catalog)

	r.HandleFunc("PUT /v2/service_instances/{instance_id}", apiHandler.Provision)
	r.HandleFunc("GET /v2/service_instances/{instance_id}", apiHandler.GetInstance)
	r.HandleFunc("PATCH /v2/service_instances/{instance_id}", apiHandler.Update)
	r.HandleFunc("DELETE /v2/service_instances/{instance_id}", apiHandler.Deprovision)

	r.HandleFunc("GET /v2/service_instances/{instance_id}/last_operation", apiHandler.LastOperation)

	r.HandleFunc("PUT /v2/service_instances/{instance_id}/service_bindings/{binding_id}", apiHandler.Bind)
	r.HandleFunc("GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}", apiHandler.GetBinding)
	r.HandleFunc("DELETE /v2/service_instances/{instance_id}/service_bindings/{binding_id}", apiHandler.Unbind)

	r.HandleFunc("GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation", apiHandler.LastBindingOperation)

	logger.Info("Listening on :8080")

	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Error("Failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
	}

}
