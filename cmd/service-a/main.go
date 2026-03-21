package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-app/internal/handler"
	"go-app/internal/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func main() {
	ctx := context.Background()

	tp, err := telemetry.InitTracer(ctx, "go-api")
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	mux := http.NewServeMux()

	// GET
	mux.Handle("/", otelhttp.NewHandler(http.HandlerFunc(handler.HelloHandler), "hello_handler"))
	mux.Handle("/healthz", otelhttp.NewHandler(http.HandlerFunc(handler.HealthHandler), "health_handler"))
	mux.Handle("/users", otelhttp.NewHandler(http.HandlerFunc(handler.UsersHandler), "users_handler"))

	// POST
	mux.Handle("/echo", otelhttp.NewHandler(http.HandlerFunc(handler.EchoHandler), "echo_handler"))
	mux.Handle("/items", otelhttp.NewHandler(http.HandlerFunc(handler.CreateItemHandler), "create_item_handler"))

	port := getEnv("SERVER_PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("server started on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	waitForShutdown(server, tp)
}

// func withCORS(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "https://YOUR_STATIC_SITE.z13.web.core.windows.net")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
// 		w.Header().Set("Access-Control-Allow-Headers", "traceparent, tracestate, content-type, request-id, request-context")

// 		if r.Method == http.MethodOptions {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

func waitForShutdown(server *http.Server, tp interface{ Shutdown(context.Context) error }) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	if err := tp.Shutdown(ctx); err != nil {
		log.Printf("tracer shutdown error: %v", err)
	}

	log.Println("server exited")
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}
