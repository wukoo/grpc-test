package logger

import (
	"fmt"
	"os"
	"time"

	rotatelogs "chainmaker.org/chainmaker/common/v2/log/file-rotatelogs"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	SHOWLINE    = false
	LEVEL_DEBUG = "DEBUG"
	LEVEL_INFO  = "INFO"
	LEVEL_WARN  = "WARN"
	LEVEL_ERROR = "ERROR"
)

var (
	logPathFromConfig          string
	displayInConsoleFromConfig bool
)

func NewLogger(name, level string) *zap.SugaredLogger {

	logPathFromConfig = fmt.Sprintf("%s.log", name)
	displayInConsoleFromConfig = true

	encoder := getEncoder()
	writeSyncer := getLogWriter()

	var logLevel zapcore.Level

	switch level {
	case LEVEL_DEBUG:
		logLevel = zap.DebugLevel
	case LEVEL_INFO:
		logLevel = zap.InfoLevel
	case LEVEL_WARN:
		logLevel = zap.WarnLevel
	case LEVEL_ERROR:
		logLevel = zap.ErrorLevel
	default:
		logLevel = zap.InfoLevel
	}

	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		logLevel,
	)

	logger := zap.New(core).Named(name)
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			return
		}
	}(logger)

	if SHOWLINE {
		logger = logger.WithOptions(zap.AddCaller())
	}

	sugarLogger := logger.Sugar()

	return sugarLogger
}

func getEncoder() zapcore.Encoder {

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    CustomLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {

	hook, _ := rotatelogs.New(
		logPathFromConfig+".%Y%m%d%H",
		rotatelogs.WithRotationTime(time.Hour*time.Duration(1)),
		rotatelogs.WithLinkName(logPathFromConfig),
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(365)),
	)

	var syncer zapcore.WriteSyncer
	if displayInConsoleFromConfig {
		syncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(hook))
	} else {
		syncer = zapcore.AddSync(hook)
	}

	return syncer
}

func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}
