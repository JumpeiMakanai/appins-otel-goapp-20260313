package main

import (
	"context"
	"database/sql" // 追加
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-app/internal/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"           // 追加
	"go.opentelemetry.io/otel/attribute" // 追加
)

// グローバル変数として定義（newDBで初期化）
var db *sql.DB

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// トレーサーの初期化
	tp, err := telemetry.InitTracer(ctx, "go-logic")
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("tracer shutdown error: %v", err)
		}
	}()

	// MySQL データベース接続を初期化 (db.go で定義された newDB を呼び出す)
	db, err = newDB(ctx)
	if err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}
	defer db.Close()

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
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.URL.Query().Get("id")
	if id == "" {
		id = "1"
	}

	// ビジネスロジック用スパンを開始
	tracer := otel.Tracer("go-logic")
	ctx, span := tracer.Start(ctx, "business_logic")
	defer span.End()
	span.SetAttributes(attribute.String("profile.id", id))

	// DB クエリ
	var profile struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Team string `json:"team"`
	}

	// db.go で otelsql を使ってラップされた db を使用
	err := db.QueryRowContext(ctx,
		"SELECT id, name, team FROM profiles WHERE id = ?", id).
		Scan(&profile.ID, &profile.Name, &profile.Team)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "profile not found", http.StatusNotFound)
		} else {
			span.RecordError(err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// --- 補助関数 (healthHandler, writeJSON, getEnv) は既存のまま ---

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
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
