package main

import (
	"embed"
	"io/fs"
	"os"
	"os/signal"
	"syscall"
)

var DebugMode = false // Global, set via CLI-Argument

//go:embed frontend/*
var frontend embed.FS

func main() {
	args := ParseCLIArguments()

	config, err := ReadConfiguration(args.ConfigFile)
	if err != nil {
		outFatal(EXIT_PARSE_CONFIG, "Cannot read/parse configuration file %s: %s\n", args.ConfigFile, err.Error())
	}

	var fFs fs.ReadFileFS = frontend
	if DebugMode {
		fFs = NewFrontendFS("frontend/")
	}

	server := NewServer(args.Port, fFs, args.CacheFile, config)

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-cancelChan
		out("Got Signal, shutting down\n")
		server.Active = false
	}()

	refreshChan := make(chan os.Signal, 1)
	signal.Notify(refreshChan, syscall.SIGUSR1)
	go func() {
		for {
			<-refreshChan
			out("Reloading configuration...\n")
			config, err := ReadConfiguration(args.ConfigFile)
			if err != nil {
				outError("Cannot reload configuration file %s: %s\n", args.ConfigFile, err.Error())
			} else {
				server.setConfiguration(config)
				out("Configuration reloaded\n")
			}
		}
	}()

	_ = server.Run() // Runs until server.Active is set to false
}
