package logger

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	verbose = false
	quiet   = false
)


var (
	infoColor    = color.New(color.FgCyan, color.Bold)
	warnColor    = color.New(color.FgYellow, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	successColor = color.New(color.FgGreen, color.Bold)
	debugColor   = color.New(color.FgMagenta)
)

func SetVerbose(v bool) {
	verbose = v
}

func SetQuiet(q bool) {
	quiet = q
}

func Info(format string, args ...any) {
	if quiet {
		return
	}
	infoColor.Print("[INFO] ")
	fmt.Printf(format+"\n", args...)
}

func Warn(format string, args ...any) {
	if quiet {
		return
	}
	warnColor.Print("[WARN] ")
	fmt.Printf(format+"\n", args...)
}

func Error(format string, args ...any) {
	errorColor.Print("[ERROR] ")
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func Success(format string, args ...any) {
	if quiet {
		return
	}
	successColor.Print("[SUCCESS] ")
	fmt.Printf(format+"\n", args...)
}

func Debug(format string, args ...any) {
	if !verbose || quiet {
		return
	}
	debugColor.Print("[DEBUG] ")
	fmt.Printf(format+"\n", args...)
}

func Fatal(format string, args ...any) {
	Error(format, args...)
	os.Exit(1)
}
