package log

import (
	"fmt"
	"os"
	"sync"
	"time"
)

//----------------------------------------------------------------------------

const (
	LEVEL_FATAL = iota
	LEVEL_ERROR
	LEVEL_WARNING
	LEVEL_INFO
	LEVEL_DEBUG
)

//----------------------------------------------------------------------------

type LogNode struct {
	Level     int
	Timestamp time.Time
	Message   string
}

func (n *LogNode) ToString() string {
	s := "[" + n.Timestamp.String() + "]["
	switch n.Level {
	case LEVEL_FATAL:
		s += "FATAL"
	case LEVEL_ERROR:
		s += "ERROR"
	case LEVEL_WARNING:
		s += "WARNING"
	case LEVEL_INFO:
		s += "INFO"
	case LEVEL_DEBUG:
		s += "DEBUG"
	}
	return s + "] " + n.Message
}

//----------------------------------------------------------------------------

type Logger struct {
	baseDir    string
	id         string
	level      int
	useLogFile bool
	f          *os.File
	yearDay    int
	in         chan *LogNode
}

//----------------------------------------------------------------------------

var globalLoggers = make(map[string]*Logger)
var globalLock *sync.RWMutex = new(sync.RWMutex)

//----------------------------------------------------------------------------

func GetLogger(baseDir string, id string, level int) *Logger {
	if len(id) > 0 {
		// Check whether such a logger exist or not.
		globalLock.RLock()
		l, okay := globalLoggers[id]
		globalLock.RUnlock()

		if okay {
			return l
		}
	}

	globalLock.Lock()
	defer globalLock.Unlock()

	// Create a new logger.
	l := new(Logger)

	l.baseDir = baseDir
	l.id = id
	l.level = level
	l.useLogFile = ((len(id) > 0) && (len(baseDir) > 0))
	l.f = nil
	l.yearDay = -1
	l.in = make(chan *LogNode, 10240)
	go (func() {
		for {
			select {
			case n, okay := <-l.in:
				if !okay {
					// The in channel had been closed.
					break
				}
				l.process(n)
			}
		}
	})()

	if l.useLogFile {
		// Check disk directories.
		err := os.MkdirAll(baseDir, 0644)
		if err != nil {
			fmt.Println(l.timedString(err.Error()))
			l.useLogFile = false
		}
	}

	globalLoggers[id] = l
	return l
}

//----------------------------------------------------------------------------

func (l *Logger) timedString(s string) string {
	return "[" + time.Now().String() + "] " + s
}

//----------------------------------------------------------------------------

func (l *Logger) push(s string, level int) {
	n := new(LogNode)
	n.Message = s
	n.Level = level
	n.Timestamp = time.Now()

	l.in <- n
}

//----------------------------------------------------------------------------

func (l *Logger) process(n *LogNode) {
	if !l.useLogFile {
		fmt.Println(n.ToString())
		return
	}

	// Check whether it needs to create a new disk file.
	if n.Timestamp.YearDay() > l.yearDay || l.f == nil {
		var err error = nil

		// Save current disk file.
		if l.f != nil {
			err = l.f.Close()
			if err != nil {
				fmt.Println(l.timedString("CloseLogFile: " + err.Error()))
			}
			l.f = nil
		}

		// Construct a new file name.
		name := l.baseDir + "/" + l.id + "." + n.Timestamp.Format("2006-01-02") + ".log"
		l.f, err = os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0644)
		if err != nil {
			fmt.Println(l.timedString("OpenLogFile: " + err.Error()))
			l.f = nil
		}
	}

	if l.f != nil {
		fmt.Fprintln(l.f, n.ToString())
	}
}

//----------------------------------------------------------------------------

func (l *Logger) Close() {
	if l.f != nil {
		l.f.Close()
		l.f = nil
	}

	delete(globalLoggers, l.id)
}

//----------------------------------------------------------------------------

func (l *Logger) SetLevel(n int) {
	if n > LEVEL_DEBUG {
		l.level = LEVEL_DEBUG
	} else if n < LEVEL_FATAL {
		l.level = LEVEL_FATAL
	} else {
		l.level = n
	}
}

//----------------------------------------------------------------------------

func (l *Logger) GetLevel() int {
	return l.level
}

//----------------------------------------------------------------------------

func (l *Logger) Debug(s string) {
	if l.level >= LEVEL_DEBUG {
		l.push(s, LEVEL_DEBUG)
	}
}

func (l *Logger) Info(s string) {
	if l.level >= LEVEL_INFO {
		l.push(s, LEVEL_INFO)
	}
}

func (l *Logger) Warning(s string) {
	if l.level >= LEVEL_WARNING {
		l.push(s, LEVEL_WARNING)
	}
}

func (l *Logger) Error(s string) {
	if l.level >= LEVEL_ERROR {
		l.push(s, LEVEL_ERROR)
	}
}

func (l *Logger) Fatal(s string) {
	if l.level >= LEVEL_FATAL {
		l.push(s, LEVEL_FATAL)
	}
}

//----------------------------------------------------------------------------
