package logger

import (
	"github.com/fatih/color"
	"os"
)

// colorWrapper wrapper function that returns function for printing with the given color
func colorWrapper(colorAttr color.Attribute) func(msg string, a ...interface{}) {
	return func(msg string, a ...interface{}) {
		color.New(colorAttr).PrintfFunc()(msg, a...)
	}
}

// Simple for printing regular text
func Simple(msg string, a ...interface{}) {
	colorWrapper(color.FgWhite)(msg, a...)
}

// Simpleln for printing regular text with new line
func Simpleln(msg string, a ...interface{}) {
	Simple(msg+"\n", a...)
}

// Info for printing info text
func Info(msg string, a ...interface{}) {
	colorWrapper(color.FgBlue)(msg, a...)
}

// Infoln for printing info text with new line
func Infoln(msg string, a ...interface{}) {
	Info(msg+"\n", a...)
}

// Header for printing header text
func Header(msg string, a ...interface{}) {
	colorWrapper(color.FgMagenta)(msg+"\n", a...)
}

// Success for printing success message
func Success(msg string, a ...interface{}) {
	colorWrapper(color.FgGreen)(msg+"\n", a...)
}

// Warning for printing warning message
func Warning(msg string, a ...interface{}) {
	colorWrapper(color.FgYellow)(msg+"\n", a...)
}

// Error for printing error message
func Error(msg string, a ...interface{}) {
	colorWrapper(color.FgRed)(msg, a...)
}

// ErrorWithExit for printing error message with the following stopping execution
func ErrorWithExit(msg string, a ...interface{}) {
	Error(msg+"\n", a...)
	os.Exit(1)
}
