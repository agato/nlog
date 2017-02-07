package nlog

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
	DebugOn       = 1
)

//log write variable
var Logger = New(os.Stderr, "", LstdFlags)

type Logging struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	prefix    string     // prefix to write at beginning of each line
	flag      int        // properties
	out       io.Writer  // destination for output
	buf       []byte     // for accumulating text to write
	debugFlag int        // 0 or 1 1 is debug
	filePath  string
}

// log initial set
func New(out io.Writer, prefix string, flag int) *Logging {
	return &Logging{out: out, prefix: prefix, flag: flag}
}

//write lock
func (l *Logging) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

//write
func (l *Logging) Output(calldepth int, s string) error {

	now := time.Now()
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)

	setOutputFile(l.buf)

	return err
}

//ファイルに対して書き込みます
func setOutputFile(buf []byte) {

	if Logger.filePath != "" {
		f, err := os.OpenFile(Logger.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("error opening file :", err.Error())
		}
		defer f.Close()
		writer := bufio.NewWriter(f)
		bw := bufio.NewWriter(writer)
		bw.Write(buf)
		bw.Flush()

		//std.SetOutput(f)
	}
}


//出力フォーマットを設定します
func SetFlags(flag int) {
	Logger.SetFlags(flag)
}

func GetFlags() int {
	return Logger.flag
}

//デバックの出力を設定します
//1 = debug on
func SetDebugFlags(debugFlag int) {
	Logger.SetDebugFlags(debugFlag)
}

func (l *Logging) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

func (l *Logging) SetDebugFlags(debugFlag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugFlag = debugFlag
}


func (l *Logging) Info(v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = "[INFOaaaaaaa]"
	l.Output(2, fmt.Sprint(v...))
}


func Info(v ...interface{}) {
	Logger.prefix = "[INFObbbbb]"
	Logger.Output(2, fmt.Sprint(v...))
}

func Debug(v ...interface{}) {
	Logger.prefix = "[DEBUG]"
	if Logger.debugFlag != DebugOn {
		return
	}
	Logger.Output(2, fmt.Sprint(v...))
}

func Error(v ...interface{}) {
	Logger.prefix = "[ERROR]"
	Logger.Output(2, fmt.Sprint(v...))
}

func Fatal(v ...interface{}) {
	Logger.prefix = "[FATAL]"
	Logger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

//ファイル出力します
func SetFilePath(filePath string) *Logging {
	Logger.filePath = filePath
	return Logger
}

//---privete funcs---

//フォーマットを設定します
func (l *Logging) formatHeader(buf *[]byte, t time.Time, file string, line int) {
	*buf = append(*buf, l.prefix...)
	if l.flag&LUTC != 0 {
		t = t.UTC()
	}
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
