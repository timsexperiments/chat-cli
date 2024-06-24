package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	labstack "github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/timsexperiments/chat-cli/internal/database"
	"github.com/timsexperiments/chat-cli/internal/handlers"
	"github.com/timsexperiments/chat-cli/internal/middleware"
)

func main() {
	e := echo.New()
	e.Use(labstack.Logger())

	e.HTTPErrorHandler = handlers.ErrorHandler

	dbPath := "data/chat.db"
	if _, err := os.Stat(dbPath); err != nil {
		os.MkdirAll("data", os.ModePerm)
		file, err := os.Create(dbPath)
		if err != nil {
			panic(fmt.Errorf("unable to create database file: %w", err))
		}
		file.Close()
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("unable to open database: %w", err))
	}
	defer db.Close()
	sqlite := database.CreateDB(db)
	e.Use(middleware.ContextDB(sqlite))

	handlers.RegisterConversationsHandlers(e)

	e.Logger.Fatal(e.Start(":8080"))
}
