package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/alde/ale/config"
	"github.com/alde/ale/db"
	"github.com/alde/ale/db/postgres"
	"github.com/alde/ale/server"
	"github.com/alde/ale/version"

	"cloud.google.com/go/datastore"

	"github.com/braintree/manners"
	"github.com/kardianos/osext"
	"github.com/sirupsen/logrus"
)

func main() {
	configFile := flag.String("config", "", "Specify a config.toml file")
	flag.Parse()
	go catchInterrupt()

	cfg := config.Initialize(*configFile)
	setupLogging(cfg)
	ctx := context.Background()
	database := setupDatabase(ctx, cfg)

	bind := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	logrus.WithFields(logrus.Fields{
		"version": version.Version,
		"address": cfg.Server.Address,
		"port":    cfg.Server.Port,
	}).Info("Launching ALE")
	router := server.NewRouter(cfg, database)
	if err := manners.ListenAndServe(bind, router); err != nil {
		logrus.WithError(err).Fatal("Unrecoverable error!")
	}
}

func setupDatabase(ctx context.Context, cfg *config.Config) db.Database {
	if (config.SQLConf{}) != cfg.PostgreSQL {
		logrus.WithFields(logrus.Fields{
			"host":     cfg.PostgreSQL.Host,
			"port":     cfg.PostgreSQL.Port,
			"username": cfg.PostgreSQL.Username,
		}).Info("configuring postgres connection")
		db, err := postgres.New(cfg)
		if err != nil {
			logrus.WithError(err).Fatal("unable to create postgres client")
		}
		return db
	}

	if (config.DatastoreConf{}) != cfg.GoogleCloudDatastore {
		logrus.WithFields(logrus.Fields{
			"namespace": cfg.GoogleCloudDatastore.Namespace,
			"project":   cfg.GoogleCloudDatastore.Project,
		}).Info("configuring datastore connection")
		ds, err := datastore.NewClient(ctx, cfg.GoogleCloudDatastore.Project)
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
	folder, err := osext.ExecutableFolder()
	if err != nil {
		logrus.WithError(err).Fatal("unable to create fake database")
	}
	database, err := db.NewFilestore(folder)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create fake database")
	}
	return database
}

func setupLogging(cfg *config.Config) {
	if cfg.Logging.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyLevel: "severity",
			},
		})
	}
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetOutput(os.Stdout)
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
