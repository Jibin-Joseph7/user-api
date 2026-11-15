package main

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/Jibin-Joseph7/user-api/config"
	sqlc "github.com/Jibin-Joseph7/user-api/db/sqlc"
	"github.com/Jibin-Joseph7/user-api/internal/handler"
	"github.com/Jibin-Joseph7/user-api/internal/logger"
	"github.com/Jibin-Joseph7/user-api/internal/middleware"
	"github.com/Jibin-Joseph7/user-api/internal/repository"
	"github.com/Jibin-Joseph7/user-api/internal/routes"
	"github.com/Jibin-Joseph7/user-api/internal/service"
)

func main() {
	cfg := config.Load()

	log, err := logger.New()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	db, err := sql.Open("mysql", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("cannot open db", zap.Error(err))
	}
	defer db.Close()
	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal("cannot connect to db", zap.Error(err))
	}

	repo := repository.New(sqlc.New(db))
	svc := service.New(repo, log)
	h := handler.New(svc, log)

	app := fiber.New()
	app.Use(middleware.RequestID(), middleware.RequestLogger(log))
	routes.Register(app, h)

	log.Info("server starting", zap.String("port", cfg.Port))
	var pingErr error

for i := 1; i <= 10; i++ {
    pingErr = db.PingContext(context.Background())

    if pingErr == nil {
        break
    }

    log.Warn("waiting for db",
        zap.Int("attempt", i),
        zap.Error(pingErr),
    )

    time.Sleep(2 * time.Second)
}

if pingErr != nil {
    log.Fatal("cannot connect to db", zap.Error(pingErr))
}
}