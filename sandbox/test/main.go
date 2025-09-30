package main

import (
	"log"
	"os"
	"runtime"
	"time"
)

var logger = log.New(os.Stdout, "", 0)

const (
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_DEBUG = "debug"
)

func main() {
	Log(LOG_INFO, "Testing")
}

func Log(level, msg string) {
	ts := time.Now().UTC().Format(time.RFC3339Nano)

	_, file, line, ok := runtime.Caller(1)
	caller := "unknown"
	if ok {
		caller = sprintf("%s:%d", file, line)
	}

	logger.Printf("level=%s ts=%s caller=%s msg=%q",
		level, ts, pathBase(caller), msg)
}

func sprintf(format string, args ...interface{}) string {
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
