package listener

import (
	"log"
	"os"
	"runtime"
	"time"
)

const (
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_DEBUG = "debug"
)

var logger = log.New(os.Stdout, "", 0)

func Log(level, msg string) {
	ts := time.Now().UTC().Format(time.RFC3339Nano)

	_, file, line, ok := runtime.Caller(1)
	caller := "unknown"
	if ok {
		caller = Sprintf("%s:%d", file, line)
	}

	logger.Printf("level=%s ts=%s caller=%s msg=%q",
		level, ts, pathBase(caller), msg)
}

func pathBase(fp string) string {
	lastslash := -1
	for id := len(fp) - 1; id >= 0; id-- {
		if fp[id] == '/' {
			lastslash = id
			break
		}
	}
	if lastslash >= 0 && lastslash+1 < len(fp) {
		return fp[lastslash+1:]
	}
	return fp
}

func TrimSpace(b []byte) []byte {
	start, end := 0, len(b)
	for start < end && (b[start] == ' ' || b[start] == '\n' || b[start] == '\t' || b[start] == '\r') {
		start++
	}
	for end > start && (b[end-1] == ' ' || b[end-1] == '\n' || b[end-1] == '\t' || b[end-1] == '\r') {
		end--
	}
	return b[start:end]
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		d := byte(n % 10)
		digits = append([]byte{d + '0'}, digits...)
		n /= 10
	}
	return sign + string(digits)
}

// alternative for fmt.Sprintf() method
func Sprintf(format string, args ...interface{}) string {
	out := []byte{}
	argIndex := 0
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) {
			i++
			switch format[i] {
			case 's':
				if v, ok := args[argIndex].(string); ok {
					out = append(out, v...)
				}
				argIndex++
			case 'd':
				if v, ok := args[argIndex].(int); ok {
					out = append(out, itoa(v)...)
				}
				argIndex++
			default:
				out = append(out, '%', format[i])
			}
		} else {
			out = append(out, format[i])
		}
	}
	return string(out)
}
