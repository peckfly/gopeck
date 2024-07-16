package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"
)

type (
	Config struct {
		Logger LoggerConfig
	}
	LoggerConfig struct {
		Debug      bool   `mapstructure:"debug"`
		Level      string `mapstructure:"level"` // debug/info/warn/error/panic/fatal
		CallerSkip int    `mapstructure:"caller_skip"`
		File       File   `mapstructure:"file"`
	}

	File struct {
		Enable     bool   `mapstructure:"enable"`
		Path       string `mapstructure:"path"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxBackups int    `mapstructure:"max_backups"`
		MaxAge     int    `mapstructure:"max_age"`
	}

	ZapInfoWriter struct {
		Logger *zap.Logger
	}

	ZapErrorWriter struct {
		Logger *zap.Logger
	}
)

func Setup(cfg *LoggerConfig) (func(), error) {
	var config zap.Config
	if cfg.Debug {
		cfg.Level = "debug"
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	config.Level.SetLevel(level)

	var (
		logger   *zap.Logger
		cleanFns []func()
	)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     TimeEncoder,                   // 自定义时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.File.Enable {
		filename := cfg.File.Path
		err := os.MkdirAll(filepath.Dir(filename), 0777)
		if err != nil {
			return nil, err
		}
		fileWriter := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    cfg.File.MaxSize,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAge,
			Compress:   true,
			LocalTime:  true,
		}

		cleanFns = append(cleanFns, func() {
			_ = fileWriter.Close()
		})

		zc := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(fileWriter),
			config.Level,
		)
		logger = zap.New(zc)
	} else {
		zc := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			config.Level,
		)
		logger = zap.New(zc)
		if err != nil {
			return nil, err
		}
	}

	skip := cfg.CallerSkip
	if skip <= 0 {
		skip = 2
	}

	logger = logger.WithOptions(
		zap.WithCaller(true),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(skip),
	)
	zap.ReplaceGlobals(logger)
	return func() {
		for _, fn := range cleanFns {
			fn()
		}
	}, nil
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339))
}

func NewZapInfoWriter() *ZapInfoWriter {
	return &ZapInfoWriter{
		Logger: Logger(),
	}
}

func NewZapErrorWriter() *ZapErrorWriter {
	return &ZapErrorWriter{
		Logger: Logger(),
	}
}

// Write Info implements io.Writer
func (gw *ZapInfoWriter) Write(p []byte) (n int, err error) {
	gw.Logger.Info(string(p))
	return len(p), nil
}

// Write Error implements io.Writer
func (gw *ZapErrorWriter) Write(p []byte) (n int, err error) {
	gw.Logger.Error(string(p))
	return len(p), nil
}

// Info logs a message at InfoLevel.
func Info(msg string, fields ...zap.Field) {
	Logger().Info(msg, fields...)
}

// Warn logs a message at WarnLevel.
func Warn(msg string, fields ...zap.Field) {
	Logger().Warn(msg, fields...)
}

// Error logs a message at ErrorLevel.
func Error(msg string, fields ...zap.Field) {
	Logger().Error(msg, fields...)
}

// ErrorStack logs the error and the stack trace
func ErrorStack(v ...any) {
	msg := fmt.Sprintf("%s\n%s", fmt.Sprint(v...), string(debug.Stack()))
	Logger().Error(msg)
}

// Must checks if err is nil, otherwise logs the error and exits.
func Must(err error) {
	if err == nil {
		return
	}
	msg := fmt.Sprintf("%+v\n\n%s", err.Error(), string(debug.Stack()))
	Error(msg)
	panic(msg)
}
