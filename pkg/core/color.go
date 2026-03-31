package core

import (
	"fmt"
	"os"
	"runtime"
)

type Color int

const (
	Reset  Color = 0
	Red    Color = 31
	Green  Color = 32
	Yellow Color = 33
	Blue   Color = 34
	Purple Color = 35
	Cyan   Color = 36
	White  Color = 37
)

var (
	supportsColor = checkColorSupport()
)

func checkColorSupport() bool {
	if runtime.GOOS == "windows" {
		return os.Getenv("TERM") != "dumb"
	}
	return os.Getenv("TERM") != "dumb"
}

func colorize(color Color, format string, args ...interface{}) string {
	if !supportsColor {
		return fmt.Sprintf(format, args...)
	}
	return fmt.Sprintf("\033[%dm%s\033[0m", color, fmt.Sprintf(format, args...))
}

func Greenf(format string, args ...interface{}) string {
	return colorize(Green, format, args...)
}

func Redf(format string, args ...interface{}) string {
	return colorize(Red, format, args...)
}

func Yellowf(format string, args ...interface{}) string {
	return colorize(Yellow, format, args...)
}

func Bluef(format string, args ...interface{}) string {
	return colorize(Blue, format, args...)
}

func Cyanf(format string, args ...interface{}) string {
	return colorize(Cyan, format, args...)
}

func Boldf(format string, args ...interface{}) string {
	if !supportsColor {
		return fmt.Sprintf(format, args...)
	}
	return fmt.Sprintf("\033[1m%s\033[0m", fmt.Sprintf(format, args...))
}
