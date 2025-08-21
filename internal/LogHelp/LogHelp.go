package LogHelp

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"time"
)

// AlwaysQueue Is a flag to be set if a alternative log delivery system shall be used.
var AlwaysQueue = false
var AlternativeWriter *io.Writer
var PrintableLogLevel LogLevelID

func init() {
	flag.IntVar(
		(*int)(&PrintableLogLevel),
		"log-level",
		2,
		"Set the level of logging\n0 = fatal\t1 = error\t2 = warning\t3 = info\t4 = debug",
	)
}

// LogLevelID Is s a simple enum for Log levels in a machine processable and human-readable format, at the same time.
type LogLevelID int

func (l LogLevelID) String() string {
	switch l {
	case Fatal:
		return "fatal"
	case Error:
		return "error"
	case Warn:
		return "warn"
	case Info:
		return "info"
	case Debug:
		return "debug"
	default:
		return "unknown"
	}
}

// LogEntry is a help structure for logs, It provides a rigid yet complete structure for logging.
type LogEntry struct {
	LogTime     time.Time   `json:"timestamp"`
	LogLevelInt LogLevelID  `json:"log_level_int"`
	LogMessage  string      `json:"log_message"`
	LogContext  interface{} `json:"log_context"`
}

func (e LogEntry) String() string {
	str, err := json.Marshal(e)
	if err != nil {
		fmt.Println("LOG ENTRY JSON ERROR:", e.LogTime, e.LogMessage)
		log.Fatal("FATAL; CANNOT RETURN LOG ENTRY", err)
	}
	return string(str)
}

func (e LogEntry) Log() {
	if AlternativeWriter != nil {
		written, err := io.WriteString(*AlternativeWriter, e.String())
		if err != nil {
			log.Fatal("FATAL; CANNOT WRITE LOG ENTRY", err, e)
		}
		if written < 5 {
			log.Fatal("writing the log has failed without error,", e.String(), written)
		}
	}
	if e.LogLevelInt >= PrintableLogLevel {
		log.Println(e.String())
	}
}

const (
	Fatal LogLevelID = iota
	Error
	Warn
	Info
	Debug
)

// The LogQueue is used for potential different log delivery systems (stdout, stderr, webhook, telegram etc).
var LogQueue = make(chan *LogEntry, 100)

func NewLog(id LogLevelID, msg string, ctx interface{}) *LogEntry {
	le := &LogEntry{
		LogTime:     time.Now(),
		LogLevelInt: id,
		LogMessage:  msg,
		LogContext:  ctx,
	}

	if AlwaysQueue {
		LogQueue <- le
	}
	return le
}
