package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/attribute"
)

func newDB(ctx context.Context) (*sql.DB, error) {
	cfg := mysql.Config{
		User:                 getEnv("MYSQL_USER", "appuser"),
		Passwd:               getEnv("MYSQL_PASSWORD", "apppass"),
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", getEnv("MYSQL_HOST", "127.0.0.1"), getEnv("MYSQL_PORT", "3306")),
		DBName:               getEnv("MYSQL_DATABASE", "appdb"),
		ParseTime:            true,
		AllowNativePasswords: true,
	}

	driverName, err := otelsql.Register(
		"mysql",
		otelsql.WithAttributes(
			attribute.String("db.system", "mysql"),
			attribute.String("db.system.name", "mysql"),
			attribute.String("server.address", getEnv("MYSQL_HOST", "mysql-dev")),
			attribute.Int("server.port", 3306),
			attribute.String("db.namespace", cfg.DBName),
		),
		// otelsql.WithSpanNameFormatter(func(_ context.Context, method otelsql.Method, query string) string {
		// 	switch method {
		// 	case otelsql.MethodConnQuery, otelsql.MethodStmtQuery:
		// 		q := strings.TrimSpace(strings.ToUpper(query))
		// 		if strings.HasPrefix(q, "SELECT") {
		// 			return "mysql SELECT profiles"
		// 		}
		// 		return "mysql query"
		// 	case otelsql.MethodConnExec, otelsql.MethodStmtExec:
		// 		return "mysql exec"
		// 	default:
		// 		return string(method)
		// 	}
		// }),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip: true,
			// OmitConnResetSession: true,
			// OmitConnectorConnect: true,
			// OmitConnPrepare:      true,
		}),
		// otelsql.WithDisableSkipErrMeasurement(true),
	)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
