package main

import (
	"flag"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-sql-driver/mysql"

	"github.com/itisbryan/oh-my-lazysql/app"
	"github.com/itisbryan/oh-my-lazysql/helpers/logger"
	"github.com/itisbryan/oh-my-lazysql/models"
	"github.com/itisbryan/oh-my-lazysql/ui"
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

	logger.Info("Starting LazySQL...", nil)

	if err := mysql.SetLogger(log.New(io.Discard, "", 0)); err != nil {
		log.Fatalf("Error setting MySQL logger: %v", err)
	}

	if err := app.LoadConfig(*configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	ui.ApplyTheme(app.App.Config().Theme)

	var initModel tea.Model
	initModel = ui.NewRootModel()

	args := flag.Args()
	if len(args) == 1 {
		conn := models.Connection{
			Name:     "CLI Connection",
			URL:      args[0],
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
