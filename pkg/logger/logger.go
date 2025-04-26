package logger

import "log"

func Info(msg string) {
	log.Printf("[INFO] %s\n", msg)
}

func Warn(msg string) {
	log.Printf("[WARN] %s\n", msg)
}

func Error(msg string) {
	log.Printf("[ERROR] %s\n", msg)
}
