package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Arguments struct {
	ConfigFile string
	Port       uint
	CacheFile  string
}

func ParseCLIArguments() *Arguments {
	args := Arguments{
		Port:       8765,
		ConfigFile: "./config.yaml",
		CacheFile:  "./cache.json",
	}

	errors := make([]string, 0)
	envPort := os.Getenv("PORT")
	if envPort != "" {
		port, err := strconv.ParseUint(envPort, 10, 16)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Invalid Environment Value 'PORT': %s", err.Error()))
		} else {
			args.Port = uint(port)
		}
	}

	flag.UintVar(&args.Port, "port", args.Port, "Port to listen on (if not set, taken from env variable PORT if available, otherwise uses default)")
	flag.StringVar(&args.ConfigFile, "config", args.ConfigFile, "Configuration file")
	flag.StringVar(&args.CacheFile, "cache", args.CacheFile, "Cache file for results")
	flag.BoolVar(&DebugMode, "debug", DebugMode, "Enable debug mode (live-frontend and logging to stdout)")
	showHelp := flag.Bool("help", false, "Show this help")
	flag.Parse()

	_, err := os.Stat(args.ConfigFile)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Configuration file cannot be found at \"%s\"\n", args.ConfigFile))
	}

	if *showHelp {
		flag.PrintDefaults()
		os.Exit(EXIT_OK)
	}

	if len(errors) > 0 {
		for _, e := range errors {
			outError(e)
		}
		os.Exit(EXIT_CLI_ARGS)
	}

	return &args
}
