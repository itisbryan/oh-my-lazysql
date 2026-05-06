//go:build bubbletea
// +build bubbletea

package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/helpers/logger"
	"github.com/jorgerojas26/lazysql/models"
	"github.com/jorgerojas26/lazysql/ui"
)

var version = "dev"

func main() {
	defaultConfigPath, err := app.DefaultConfigFile()
	if err != nil {
		log.Fatalf("Error getting default config file: %v", err)
	}
	configFile := flag.String("config", defaultConfigPath, "config file to use")
	printVersion := flag.Bool("version", false, "Show version")
	logLevel := flag.String("loglevel", "info", "Log level")
	logFile := flag.String("logfile", "", "Log file")
	readOnly := flag.Bool("read-only", false, "Connect in read-only mode")
	flag.Parse()

	if *printVersion {
		println("LazySQL version: ", version)
		os.Exit(0)
	}

	slogLevel, err := logger.ParseLogLevel(*logLevel)
	if err != nil {
		log.Fatalf("Error parsing log level: %v", err)
	}
	logger.SetLevel(slogLevel)

	if *logFile != "" {
		if err := logger.SetFile(*logFile); err != nil {
			log.Fatalf("Error setting log file: %v", err)
		}
	}

	logger.Info("Starting LazySQL (Bubbletea)...", nil)

	if err := mysql.SetLogger(log.New(io.Discard, "", 0)); err != nil {
		log.Fatalf("Error setting MySQL logger: %v", err)
	}

	if err := app.LoadConfig(*configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	var initModel tea.Model
	initModel = ui.NewRootModel()

	args := flag.Args()
	if len(args) == 1 {
		conn := models.Connection{
			Name:    "CLI Connection",
			URL:     args[0],
			ReadOnly: *readOnly,
		}
		initModel = ui.NewHomeModel(conn)
	} else if len(args) > 1 {
		log.Fatal("Only a single connection is allowed")
	}

	p := tea.NewProgram(initModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}