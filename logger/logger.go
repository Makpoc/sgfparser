package logger

import (
	"log"
)

const (
	OFF LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
)

type LogLevel int

var currentLevel = INFO

func SetLogLevel(level LogLevel) {
	currentLevel = level
}

func LogError(msg string) {
	if currentLevel >= ERROR {
		log.Printf("ERROR: %s", msg)
	}
}

func LogWarn(msg string) {
	if currentLevel >= WARN {
		log.Printf("WARN: %s", msg)
	}
}

func LogInfo(msg string) {
	if currentLevel >= INFO {
		log.Printf("INFO: %s", msg)
	}
}

func LogDebug(msg string) {
	if currentLevel >= DEBUG {
		log.Printf("DEBUG: %s", msg)
	}
}
