package logger

import (
	"os"
	"time"
)

type LogLevel int

const (
	InfoLevel LogLevel = iota
	DebugLevel
	CriticalLevel
	NoneLevel
)

const (
	WORKSPACE_PKR_DIR  = ".PKr"
	LOGS_PKR_FILE_PATH = WORKSPACE_PKR_DIR + "\\logs.txt"
	USER_LOG_FILE_PATH = "\\logs.txt"
)

var logLevelName = map[LogLevel]string{
	InfoLevel:     "INFO",
	DebugLevel:    "DEBUG",
	CriticalLevel: "CRITICAL",
}

func (ll LogLevel) String() string {
	return logLevelName[ll]
}

func IntToLog(level int) LogLevel {
	var logLevel LogLevel
	switch level {
	case 0:
		logLevel = InfoLevel
	case 1:
		logLevel = DebugLevel
	case 2:
		logLevel = CriticalLevel
	default:
		logLevel = NoneLevel
	}

	return logLevel
}

func getDateAndTime() string {
	currentTime := time.Now()

	layout := "02/01/2006 15:04:05"

	return currentTime.Format(layout)
}

// [ ] Find a Way to Avoid Frequent Read and Writes... IDK How ? Maybe its not possible ?

func logToFile(filepath, log string) error {
	// Opens or Creates the Log File
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	dateTime := getDateAndTime()
	defer file.Close()
	if _, err := file.Write([]byte("[" + dateTime + "]" + log + "\n")); err != nil {
		return err
	}

	return nil
}
