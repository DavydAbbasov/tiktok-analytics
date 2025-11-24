package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"ttanalytic/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	user := url.QueryEscape(cfg.SQLDataBase.Username)
	pass := url.QueryEscape(cfg.SQLDataBase.Password)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user,
		pass,
		cfg.SQLDataBase.Server,
		cfg.SQLDataBase.Port,
		cfg.SQLDataBase.Database,
	)

	m, err := migrate.New("file://internal/migrations", dsn)
	if err != nil {
		log.Fatalf("migrate new: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no migrations to apply")
			return
		}
		log.Fatalf("migrate up: %v", err)
	}

	log.Println("migrations applied")
}
