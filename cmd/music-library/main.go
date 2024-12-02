package main

import (
	"context"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/nosikmy/music-library/internal/app/handler"
	"github.com/nosikmy/music-library/internal/app/repository"
	"github.com/nosikmy/music-library/internal/app/server"
	"github.com/nosikmy/music-library/internal/app/services"
	"github.com/nosikmy/music-library/logger"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

//	@title			Music Library
//	@version		1.0
//	@description	API Server for Music Library Service

// @host		localhost:8080
// @BasePath	/
func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalln("Error loading .env file: " + err.Error())
	}

	myLogger := logger.SetUpLogger(os.Getenv("LOGGER_TYPE"))

	cfgDB := repository.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := repository.NewPostgresDB(cfgDB)

	if err != nil {
		myLogger.Error("Error occured while init DB: " + err.Error())
		return
	}

	libraryRepository := repository.NewLibraryRepository(db)
	songRepository := repository.NewSongRepository(db)
	songChangerRepository := repository.NewSongChangerRepository(db)
	versesRepository := repository.NewVersesRepository(db)

	libraryService := services.NewLibraryService(myLogger, libraryRepository)
	songService := services.NewSongService(myLogger, songRepository, songChangerRepository, versesRepository)

	handlers := handler.NewHandler(myLogger, libraryService, songService)

	srv := new(server.Server)
	bindAddr := os.Getenv("BIND_ADDR")

	go func() {
		if err := srv.Run(bindAddr, handlers.InitRoutes()); err != nil {
			myLogger.Error("Error while running server" + err.Error())
			return
		}
		myLogger.Info("Server shuts down")
	}()

	myLogger.Info("Server started", slog.String("port", bindAddr))

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-quitSignal

	if err := srv.Shutdown(context.Background()); err != nil {
		myLogger.Error("Can't terminate server: %s" + err.Error())
	}
	if err := db.Close(); err != nil {
		myLogger.Error("Can't close DB connection: %s" + err.Error())
	}
}
