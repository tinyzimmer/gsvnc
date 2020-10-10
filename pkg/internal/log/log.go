package log

import (
	"fmt"
	glog "log"
	"os"
	"path"
	"runtime"

	"github.com/tinyzimmer/gsvnc/pkg/config"
)

var infoLogger, warningLogger, errorLogger, debugLogger *glog.Logger

func init() {
	infoLogger = glog.New(os.Stderr, "INFO: ", glog.Ldate|glog.Ltime)
	warningLogger = glog.New(os.Stderr, "WARNING: ", glog.Ldate|glog.Ltime)
	errorLogger = glog.New(os.Stderr, "ERROR: ", glog.Ldate|glog.Ltime)
	debugLogger = glog.New(os.Stderr, "DEBUG: ", glog.Ldate|glog.Ltime)
}

func formatNormal(args ...interface{}) string {
	_, file, line, _ := runtime.Caller(2)
	out := fmt.Sprintf("%s:%d: ", path.Base(file), line)
	out += fmt.Sprint(args...)
	return out
}

func formatFormat(fstr string, args ...interface{}) string {
	_, file, line, _ := runtime.Caller(2)
	out := fmt.Sprintf("%s:%d: ", path.Base(file), line)
	out += fmt.Sprintf(fstr, args...)
	return out
}

// Info is the equivalent of a log.Println on the info logger.
func Info(args ...interface{}) {
	infoLogger.Println(formatNormal(args...))
}

// Infof is the equivalent of a log.Printf on the info logger.
func Infof(fstr string, args ...interface{}) {
	infoLogger.Println(formatFormat(fstr, args...))
}

// Warning is the equivalent of a log.Println on the warning logger.
func Warning(args ...interface{}) {
	warningLogger.Println(formatNormal(args...))
}

// Warningf is the equivalent of a log.Printf on the warning logger.
func Warningf(fstr string, args ...interface{}) {
	warningLogger.Println(formatFormat(fstr, args...))
}

// Error is the equivalent of a log.Println on the error logger.
func Error(args ...interface{}) {
	errorLogger.Println(formatNormal(args...))
}

// Errorf is the equivalent of a log.Printf on the error logger.
func Errorf(fstr string, args ...interface{}) {
	errorLogger.Println(formatFormat(fstr, args...))
}

// Debug is the equivalent of a log.Println on the debug logger.
func Debug(args ...interface{}) {
	if config.Debug {
		debugLogger.Println(formatNormal(args...))
	}
}

// Debugf is the equivalent of a log.Printf on the debug logger.
func Debugf(fstr string, args ...interface{}) {
	if config.Debug {
		debugLogger.Println(formatFormat(fstr, args...))
	}
}
