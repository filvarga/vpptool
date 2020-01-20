package common

import (
	"fmt"
	"io"
	"os"
)

func log(w io.Writer, format string, args ...interface{}) (i int, err error) {
	return fmt.Fprintf(w, "%s: %s\n", os.Args[0], fmt.Sprintf(format, args...))
}

func LogDebug(format string, args ...interface{}) (i int, err error) {
	return fmt.Fprintf(os.Stderr, "%s (debug):\n%s", os.Args[0], fmt.Sprintf(format, args...))
}

func LogInfo(format string, args ...interface{}) {
	log(os.Stdout, format, args...)
}

func LogError(format string, args ...interface{}) {
	log(os.Stderr, format, args...)
}

/* vim: set ts=2: */
