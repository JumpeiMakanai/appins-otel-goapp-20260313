package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-app/internal/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	ctx := context.Background()

	tp, err := telemetry.InitTracer(ctx, "profile-service")
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/healthz", otelhttp.NewHandler(http.HandlerFunc(healthHandler), "health_handler"))
	mux.Handle("/profile", otelhttp.NewHandler(http.HandlerFunc(profileHandler), "profile_handler"))

	port := getEnv("SERVER_PORT", "8081")

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("profile service started on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	waitForShutdown(server, tp)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_ = simulateBusiness(ctx)

	id := r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":    1,
		"name":  "taro",
		"team":  "platform",
		"rawId": id,
	})
}

func simulateBusiness(ctx context.Context) error {
	tracer := otel.Tracer("profile-service/internal/service")

	_, span := tracer.Start(ctx, "load_profile")
	defer span.End()

	span.SetAttributes(
		attribute.String("profile.source", "mock"),
	)

	time.Sleep(120 * time.Millisecond)
	return nil
}

func waitForShutdown(server *http.Server, tp interface{ Shutdown(context.Context) error }) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)
	_ = tp.Shutdown(ctx)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}