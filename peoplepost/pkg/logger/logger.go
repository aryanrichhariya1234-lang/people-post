package logger

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() {
	Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(msg string) {
	Logger.Println("[INFO]", msg)
}

func Error(msg string) {
	Logger.Println("[ERROR]", msg)
}

func Fatal(msg string) {
	Logger.Fatal("[FATAL]", msg)
}