package log

import (
	"fmt"
	"log"
	"os"

	"github.com/mgutz/ansi"
)

type LogLevel int
type Colour int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

const (
	BLACK Colour = iota
	BLUE
	RED
	GREEN
	GREY
	YELLOW
	MAGENTA
	CYAN
	WHITE
	LIGHTBLACK
	LIGHTRED
	LIGHTGREEN
	LIGHTYELLOW
	LIGHTBLUE
	LIGHTMAGENTA
	LIGHTCYAN
	LIGHTWHITE
)

var coloursMap = map[Colour]string{
	BLACK:        ansi.ColorCode("black"),
	RED:          ansi.ColorCode("red"),
	GREEN:        ansi.ColorCode("green"),
	GREY:         string([]byte{'\033', '[', '3', '2', ';', '1', 'm'}),
	YELLOW:       ansi.ColorCode("yellow"),
	BLUE:         ansi.ColorCode("blue"),
	MAGENTA:      ansi.ColorCode("magenta"),
	CYAN:         ansi.ColorCode("cyan"),
	WHITE:        ansi.ColorCode("white"),
	LIGHTBLACK:   ansi.ColorCode("black+h"),
	LIGHTRED:     ansi.ColorCode("red+h"),
	LIGHTGREEN:   ansi.ColorCode("green+h"),
	LIGHTYELLOW:  ansi.ColorCode("yellow+h"),
	LIGHTBLUE:    ansi.ColorCode("blue+h"),
	LIGHTMAGENTA: ansi.ColorCode("magenta+h"),
	LIGHTCYAN:    ansi.ColorCode("cyan+h"),
	LIGHTWHITE:   ansi.ColorCode("white+h"),
}

// Logging facility
type ParityLogger struct {
	log.Logger
	Level LogLevel
}

func init() {
	log.SetFlags(0) // No timestamps
}

func NewLogger() *ParityLogger {
	return &ParityLogger{Level: INFO}
}

var std = NewLogger()

func (m *ParityLogger) Trace(format string, v ...interface{}) {
	m.Log(TRACE, format, v...)
}

func (m *ParityLogger) Debug(format string, v ...interface{}) {
	m.Log(DEBUG, format, v...)
}

func (m *ParityLogger) Info(format string, v ...interface{}) {
	m.Log(INFO, format, v...)
}

func (m *ParityLogger) Stage(format string, v ...interface{}) {
	log.Printf("Stage : "+format+"\n", v...)
}

func (m *ParityLogger) Step(format string, v ...interface{}) {
	log.Printf(" ---> "+format+"\n", v...)
}

func (m *ParityLogger) Warn(format string, v ...interface{}) {
	m.Log(WARN, format, v...)
}

func (m *ParityLogger) Error(format string, v ...interface{}) {
	m.Log(ERROR, format, v...)
}

func (m *ParityLogger) Fatal(v ...interface{}) {
	s := fmt.Sprint(v...)
	m.Log(FATAL, s)
	os.Exit(1)
}

func (m *ParityLogger) Fatalf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	m.Log(FATAL, s)
	os.Exit(1)
}

func (m *ParityLogger) Log(l LogLevel, format string, v ...interface{}) {
	if l >= m.Level {
		var level string
		var colorFormat = ""
		switch l {
		case TRACE:
			level = "TRACE"
		case DEBUG:
			level = "DEBUG"
		case INFO:
			level = "INFO"
		case WARN:
			level = "WARN"
		case ERROR:
			level = "ERROR"
			colorFormat = coloursMap[LIGHTRED]
		case FATAL:
			level = "FATAL"
			colorFormat = coloursMap[LIGHTRED]
		}

		log.Printf("      "+"["+level+"] "+colorFormat+format+ansi.Reset+"\n", v...)
		// log.Printf(colorFormat+format+ansi.Reset+"\n", v...)
		// log.Printf("["+level+"]\t\t"+colorFormat+format+ansi.Reset+"\n", v...)
	}
}

func (m *ParityLogger) SetLevel(l LogLevel) {
	m.Level = l
}

func Colorize(colour Colour, format string) string {
	return fmt.Sprintf("%s%s%s", coloursMap[colour], format, ansi.Reset)
}

func Trace(format string, v ...interface{}) {
	std.Log(TRACE, format, v...)
}

func Debug(format string, v ...interface{}) {
	std.Log(DEBUG, format, v...)
}

func Info(format string, v ...interface{}) {
	std.Log(INFO, format, v...)
}

func Stage(format string, v ...interface{}) {
	std.Stage(format, v...)
}

func Step(format string, v ...interface{}) {
	std.Step(format, v...)
}

func Banner(s string) {
	fmt.Print(s)
}

func Warn(format string, v ...interface{}) {
	std.Log(WARN, format, v...)
}

func Error(format string, v ...interface{}) {
	std.Log(ERROR, format, v...)
}

func Fatal(v ...interface{}) {
	std.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}

func Log(l LogLevel, format string, v ...interface{}) {
	std.Log(l, format, v...)
}

func SetLevel(l LogLevel) {
	std.SetLevel(l)
}
