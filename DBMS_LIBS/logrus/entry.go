package logrus

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var bufferPool *sync.Pool

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
}

// Defines the key when adding errors using WithError.
var ErrorKey = "error"

// An entry is the final or intermediate Logrus logging entry. It contains all
// the fields passed with WithField{,s}. It's finally logged when Debug, Info,
// Warn, Error, Fatal or Panic is called on it. These objects can be reused and
// passed around as much as you wish to avoid field duplication.
type Entry struct {
	Logger *Logger

	// Contains all the fields set by the user.
	Data Fields

	// Time at which the log entry was created
	Time time.Time

	// Level the log entry was logged at: Debug, Info, Warn, Error, Fatal or Panic
	Level Level

	// Message passed to Debug, Info, Warn, Error, Fatal or Panic
	Message string

	// When formatter is called in entry.log(), an Buffer may be set to entry
	Buffer *bytes.Buffer
	Header string
}

func NewEntry(logger *Logger) *Entry {
	et := &Entry{
		Logger: logger,
		// Default is three fields, give a little extra room
		Data: make(Fields, 5),
	}
	return et
}

// Returns the string representation from the reader and ultimately the
// formatter.
func (entry *Entry) String() (string, error) {
	serialized, err := entry.Logger.Formatter.Format(entry)
	if err != nil {
		return "", err
	}
	str := string(serialized)
	return str, nil
}

// Add an error as single field (using the key defined in ErrorKey) to the Entry.
func (entry *Entry) WithError(err error) *Entry {
	return entry.WithField(ErrorKey, err)
}

// Add a single field to the Entry.
func (entry *Entry) WithField(key string, value interface{}) *Entry {
	return entry.WithFields(Fields{key: value})
}

// Add a map of fields to the Entry.
func (entry *Entry) WithFields(fields Fields) *Entry {
	data := make(Fields, len(entry.Data)+len(fields))
	for k, v := range entry.Data {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	return &Entry{Logger: entry.Logger, Data: data}
}

func FileLine(s bool) string {
	pc, file, line, ok := runtime.Caller(6) // called in log is 6, called in Info is 5
	if !ok {
		file = "???"
		line = 0
	}
	if s {
		file = filepath.Base(file)
	}

	funcname := runtime.FuncForPC(pc).Name()
	funcname = filepath.Base(funcname)
	return fmt.Sprintf("%s:%s:%d", file, funcname, line)
}

// !!!! log func
// This function is not declared with a pointer value because otherwise
// race conditions will occur when using multiple goroutines
func (entry Entry) log(level Level, msg string) {
	var buffer *bytes.Buffer
	entry.Time = time.Now()
	entry.Level = level
	entry.Message = msg
	entry.Header = FileLine(true)

	if err := entry.Logger.Hooks.Fire(level, &entry); err != nil {
		entry.Logger.mu.Lock()
		fmt.Fprintf(os.Stderr, "Failed to fire hook: %v\n", err)
		entry.Logger.mu.Unlock()
	}
	buffer = bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer bufferPool.Put(buffer)
	entry.Buffer = buffer
	serialized, err := entry.Logger.Formatter.Format(&entry)
	entry.Buffer = nil
	if err != nil {
		entry.Logger.mu.Lock()
		fmt.Fprintf(os.Stderr, "Failed to obtain reader, %v\n", err)
		entry.Logger.mu.Unlock()
	} else {
		_, err := entry.Logger.Out.Write(serialized)
		if err != nil {
			entry.Logger.mu.Lock()
			fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
			entry.Logger.mu.Unlock()
		}
	}

	// To avoid Entry#log() returning a value that only would make sense for
	// panic() to use in Entry#Panic(), we avoid the allocation by checking
	// directly here.
	if level <= PanicLevel {
		panic(&entry)
	}
}

func (entry *Entry) Debug(args ...interface{}) {
	if entry.Logger.level() >= DebugLevel {
		entry.log(DebugLevel, fmt.Sprint(args...))
	}
}

func (entry *Entry) Print(args ...interface{}) {
	entry.Info(args...)
}

func (entry *Entry) Info(args ...interface{}) {
	if entry.Logger.level() >= InfoLevel {
		entry.log(InfoLevel, fmt.Sprint(args...))
	}
}

func (entry *Entry) Warn(args ...interface{}) {
	if entry.Logger.level() >= WarnLevel {
		entry.log(WarnLevel, fmt.Sprint(args...))
	}
}

func (entry *Entry) Warning(args ...interface{}) {
	entry.Warn(args...)
}

func (entry *Entry) Error(args ...interface{}) {
	if entry.Logger.level() >= ErrorLevel {
		entry.log(ErrorLevel, fmt.Sprint(args...))
	}
}

func (entry *Entry) Fatal(args ...interface{}) {
	if entry.Logger.level() >= FatalLevel {
		entry.log(FatalLevel, fmt.Sprint(args...))
	}
	Exit(1)
}

func (entry *Entry) Panic(args ...interface{}) {
	if entry.Logger.level() >= PanicLevel {
		entry.log(PanicLevel, fmt.Sprint(args...))
	}
	panic(fmt.Sprint(args...))
}

// Entry Printf family functions

func (entry *Entry) Debugf(format string, args ...interface{}) {
	if entry.Logger.level() >= DebugLevel {
		entry.Debug(fmt.Sprintf(format, args...))
	}
}

func (entry *Entry) Infof(format string, args ...interface{}) {
	if entry.Logger.level() >= InfoLevel {
		entry.Info(fmt.Sprintf(format, args...))
	}
}

func (entry *Entry) Printf(format string, args ...interface{}) {
	entry.Infof(format, args...)
}

func (entry *Entry) Warnf(format string, args ...interface{}) {
	if entry.Logger.level() >= WarnLevel {
		entry.Warn(fmt.Sprintf(format, args...))
	}
}

func (entry *Entry) Warningf(format string, args ...interface{}) {
	entry.Warnf(format, args...)
}

func (entry *Entry) Errorf(format string, args ...interface{}) {
	if entry.Logger.level() >= ErrorLevel {
		entry.Error(fmt.Sprintf(format, args...))
	}
}

func (entry *Entry) Fatalf(format string, args ...interface{}) {
	if entry.Logger.level() >= FatalLevel {
		entry.Fatal(fmt.Sprintf(format, args...))
	}
	Exit(1)
}

func (entry *Entry) Panicf(format string, args ...interface{}) {
	if entry.Logger.level() >= PanicLevel {
		entry.Panic(fmt.Sprintf(format, args...))
	}
}

// Entry Println family functions

func (entry *Entry) Debugln(args ...interface{}) {
	if entry.Logger.level() >= DebugLevel {
		entry.Debug(entry.sprintlnn(args...))
	}
}

func (entry *Entry) Infoln(args ...interface{}) {
	if entry.Logger.level() >= InfoLevel {
		entry.Info(entry.sprintlnn(args...))
	}
}

func (entry *Entry) Println(args ...interface{}) {
	entry.Infoln(args...)
}

func (entry *Entry) Warnln(args ...interface{}) {
	if entry.Logger.level() >= WarnLevel {
		entry.Warn(entry.sprintlnn(args...))
	}
}

func (entry *Entry) Warningln(args ...interface{}) {
	entry.Warnln(args...)
}

func (entry *Entry) Errorln(args ...interface{}) {
	if entry.Logger.level() >= ErrorLevel {
		entry.Error(entry.sprintlnn(args...))
	}
}

func (entry *Entry) Fatalln(args ...interface{}) {
	if entry.Logger.level() >= FatalLevel {
		entry.Fatal(entry.sprintlnn(args...))
	}
	Exit(1)
}

func (entry *Entry) Panicln(args ...interface{}) {
	if entry.Logger.level() >= PanicLevel {
		entry.Panic(entry.sprintlnn(args...))
	}
}

// Sprintlnn => Sprint no newline. This is to get the behavior of how
// fmt.Sprintln where spaces are always added between operands, regardless of
// their type. Instead of vendoring the Sprintln implementation to spare a
// string allocation, we do the simplest thing.
func (entry *Entry) sprintlnn(args ...interface{}) string {
	msg := fmt.Sprintln(args...)
	return msg[:len(msg)-1]
}
