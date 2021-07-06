package logger

import "log"

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
	Fatal
)

const defaultLogLevel = Debug

type LogLevel uint

type Logger struct {
	Level LogLevel
}

func NewLogger(level ...LogLevel) *Logger {
	logger := new(Logger)
	if len(level) == 0 {
		logger.Level = defaultLogLevel
	} else {
		logger.Level = level[0]
	}

	return logger
}

func (l *Logger) SetLevel(level LogLevel) {
	l.Level = level
}

func (l *Logger) Debug(arg string) {
	if l.checkLevel(Debug) {
		log.Printf("%c[%d;%d;%dm%v%c[0m\n", 0x1B, 1, 1, 1, arg, 0x1B)
	}
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	if l.checkLevel(Debug) {
		var fargs []interface{}
		fargs = append(fargs, 0x1B, 1, 1, 1)
		fargs = append(fargs, args...)
		fargs = append(fargs, 0x1B)
		log.Printf("%c[%d;%d;%dm"+template+"%c[0m\n", fargs...)
	}
}

func (l *Logger) Info(arg string) {
	if l.checkLevel(Info) {
		log.Printf("%c[%d;%d;%dm%v%c[0m\n", 0x1B, 1, 1, 1, arg, 0x1B)
	}
}

func (l *Logger) Infof(template string, args ...interface{}) {
	if l.checkLevel(Info) {
		var fargs []interface{}
		fargs = append(fargs, 0x1B, 1, 1, 1)
		fargs = append(fargs, args...)
		fargs = append(fargs, 0x1B)
		log.Printf("%c[%d;%d;%dm"+template+"%c[0m\n", fargs...)
	}
}

func (l *Logger) Warn(arg string) {
	if l.checkLevel(Warn) {
		log.Printf("%c[%d;%d;%dm%v%c[0m\n", 0x1B, 1, 1, 33, arg, 0x1B)
	}
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	if l.checkLevel(Warn) {
		var fargs []interface{}
		fargs = append(fargs, 0x1B, 1, 1, 33)
		fargs = append(fargs, args...)
		fargs = append(fargs, 0x1B)
		log.Printf("%c[%d;%d;%dm"+template+"%c[0m\n", fargs...)
	}
}

func (l *Logger) Error(arg string) {
	if l.checkLevel(Error) {
		log.Printf("%c[%d;%d;%dm%v%c[0m\n", 0x1B, 1, 1, 31, arg, 0x1B)
	}
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	if l.checkLevel(Error) {
		var fargs []interface{}
		fargs = append(fargs, 0x1B, 1, 1, 31)
		fargs = append(fargs, args...)
		fargs = append(fargs, 0x1B)
		log.Printf("%c[%d;%d;%dm"+template+"%c[0m\n", fargs...)
	}
}

func (l *Logger) Fatal(arg string) {
	if l.checkLevel(Fatal) {
		log.Panicln(arg)
	}
}

func (l *Logger) Fatalf(template string, args ...interface{}) {
	if l.checkLevel(Error) {
		log.Panicf(template, args...)
	}
}

// checkLevel 判断当前level是否输出该日志
func (l *Logger) checkLevel(level LogLevel) bool {
	return level >= l.Level
}
