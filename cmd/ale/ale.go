package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/alde/ale/config"
	"github.com/alde/ale/db"
	"github.com/alde/ale/server"
	"github.com/alde/ale/version"

	"cloud.google.com/go/datastore"
	"github.com/sirupsen/logrus"
	"github.com/braintree/manners"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Printf(version.Version)
		os.Exit(0)
	}

	go catchInterrupt()

	cfg := config.Initialize("") // TODO: --config <file> to read file path from commandline
	setupLogging(cfg)
	ctx := context.Background()
	database := setupDatabase(ctx, cfg)

	bind := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	logrus.WithFields(logrus.Fields{
		"version": version.Version,
		"address": cfg.Address,
		"port":    cfg.Port,
	}).Info("Launching ALE")
	router := server.NewRouter(cfg, database)
	if err := manners.ListenAndServe(bind, router); err != nil {
		logrus.WithError(err).Fatal("Unrecoverable error!")
	}
}

func setupDatabase(ctx context.Context, cfg *config.Config) db.Database {
	if cfg.Database.Type == "datastore" {
		logrus.WithFields(logrus.Fields{
			"namespace": cfg.Database.Namespace,
			"project":   cfg.Database.Project,
			"type":      cfg.Database.Type,
		}).Info("configuring database connection")
		ds, err := datastore.NewClient(ctx, cfg.Database.Project)
		if err != nil {
			logrus.WithError(err).Fatal("unable to initialize datastore library")
		}
		database, err := db.NewDatastore(ctx, cfg, ds)
		if err != nil {
			logrus.WithError(err).Fatal("unable to create datastore client")
		}
		if _, err := database.Has("0"); err != nil {
			logrus.WithError(err).Fatal("unable to check connection to the database")
		}
		return database
	}
	logrus.Info("setting up filesystem pretend database")
	database, err := db.NewFilestore(ctx, cfg)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create fake database")
	}
	return database
}

func setupLogging(cfg *config.Config) {
	if cfg.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
}

func catchInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	if s != os.Interrupt && s != os.Kill {
		return
	}
	logrus.Info("Shutting down.")
	os.Exit(0)
}
