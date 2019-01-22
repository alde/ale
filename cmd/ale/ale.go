package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/alde/ale/config"
	"github.com/alde/ale/server"
	"github.com/alde/ale/version"

	"github.com/Sirupsen/logrus"
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

	bind := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	logrus.WithFields(logrus.Fields{
		"version": version.Version,
		"address": cfg.Address,
		"port":    cfg.Port,
	}).Info("Launching ALE")
	router := server.NewRouter(cfg)
	if err := manners.ListenAndServe(bind, router); err != nil {
		logrus.WithError(err).Fatal("Unrecoverable error!")
	}
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
