// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package logger defines gorm logger
package db

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/lunarianss/Hurricane/pkg/log"
	gormlogger "gorm.io/gorm/logger"
)

// Define colors.
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

// Define gorm log levels.
const (
	Silent gormlogger.LogLevel = iota + 1
	Error
	Warn
	Info
)

// Writer log writer interface.
type Writer interface {
	Printf(string, ...interface{})
}

// Config defines a gorm logger configuration.
type Config struct {
	SlowThreshold time.Duration
	Colorful      bool
	LogLevel      gormlogger.LogLevel
}

// New create a gorm logger instance.
func NewGormLogger(level int) gormlogger.Interface {
	var (
		infoStr      = "%s[info] "
		warnStr      = "%s[warn] "
		errStr       = "%s[error] "
		traceStr     = "[%s][%.3fms] [rows:%d] %s"
		traceWarnStr = "%s %s[%.3fms] [rows:%d] %s"
		traceErrStr  = "%s %s[%.3fms] [rows:%d] %s"
	)

	config := Config{
		SlowThreshold: 200 * time.Millisecond,
		Colorful:      false,
		LogLevel:      gormlogger.LogLevel(level),
	}

	if config.Colorful {
		infoStr = Green + "%s " + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "%s " + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "%s " + Reset + Red + "[error] " + Reset
		traceStr = Green + "%s " + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "%s " + Yellow + "%s " + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "%s " + MagentaBold + "%s " + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	}
	return &gormLogger{
		Writer:       log.StdInfoLogger(),
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type gormLogger struct {
	Writer
	Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode.
func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level

	return &newlogger
}

// Info print info.
func (l gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Info {
		l.Printf(l.infoStr+msg, append([]interface{}{fileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages.
func (l gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Warn {
		l.Printf(l.warnStr+msg, append([]interface{}{fileWithLineNum()}, data...)...)
	}
}

// Error print error messages.
func (l gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Error {
		l.Printf(l.errStr+msg, append([]interface{}{fileWithLineNum()}, data...)...)
	}
}

// Trace print sql message.
func (l gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= Error:
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceErrStr, fileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceErrStr, fileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf(l.traceWarnStr, fileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceWarnStr, fileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel >= Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceStr, fileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceStr, fileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func fileWithLineNum() string {
	for i := 4; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)

		// if ok && (!strings.HasPrefix(file, gormSourceDir) || strings.HasSuffix(file, "_test.go")) {
		if ok && !strings.HasSuffix(file, "_test.go") {
			dir, f := filepath.Split(file)

			return filepath.Join(filepath.Base(dir), f) + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}
