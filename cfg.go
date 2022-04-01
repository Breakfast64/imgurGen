package main

import (
	"log"
	"os"

	flag "github.com/spf13/pflag"
)

type userConfig struct {
	Length      int
	Amount      int
	Extension   string
	Logger      *log.Logger
	Connections int
}

func parseArgs() (cfg userConfig) {

	var help, shouldLog bool

	flag.BoolVarP(&help, "help", "h", false, "displays this screen")
	flag.IntVarP(&cfg.Length, "length", "l", 5, "length of ID")
	flag.IntVarP(&cfg.Amount, "amount", "n", 1, "how many urls to produce")
	flag.IntVarP(&cfg.Connections, "connections", "c", 256, "how many concurrent connections will be used")
	flag.StringVarP(&cfg.Extension, "extension", "x", ".jpeg", "image's extension")
	flag.BoolVarP(&shouldLog, "verbose", "v", false, "log extra info to stderr")

	panicOnErr(flag.CommandLine.MarkHidden("help"))

	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}

	if shouldLog {
		cfg.Logger = log.New(os.Stderr, "", 0)
	}

	if cfg.Extension[0] != '.' {
		cfg.Extension = "." + cfg.Extension
	}

	return
}
