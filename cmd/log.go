package main

import (
	"fmt"
	"os"
)

// / Logger/Output handling
func out(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format, args...)
}

func outDebug(format string, args ...any) {
	if DebugMode {
		fmt.Fprintf(os.Stdout, format, args...)
	}
}

func outError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func outFatal(exitCode int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(exitCode)
}
