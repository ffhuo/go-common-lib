package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerConfig struct {
	ModuleName string
	Level      string
	ErrorPath  string
	LogPath    string
	HasHTTPNet bool
}
type Logger struct {
	log    *zap.Logger
	config LoggerConfig
}

func NewLogInstance(moduleName, level, errorPath, logPath string) *Logger {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(getLoggerLevel(level))
	return initLogger(NewLoggerConfig(LoggerConfig{}), atomicLevel)
}

func NewLoggerConfig(config LoggerConfig) LoggerConfig {
	if config.Level == "" {
		config.Level = "debug"
	}

	if config.ErrorPath == "" {
		config.ErrorPath = "./logs/error.log"
	}

	if config.LogPath == "" {
		config.LogPath = "./logs/log.log"
	}

	if config.ModuleName != "" {
		config.ModuleName += ": "
	}

	return config
}

func (logger *Logger) Sync() {
	logger.log.Sync()
}

func (logger *Logger) Error(msg string, err error, fields ...zap.Field) {
	if err != nil {
		logger.log.Error(logger.config.ModuleName+msg+err.Error(), fields...)
	} else {
		logger.log.Error(logger.config.ModuleName+msg, fields...)
	}
}

func (logger *Logger) Info(msg string, fields ...zap.Field) {
	logger.log.Info(logger.config.ModuleName+msg, fields...)
}

func (logger *Logger) Debug(msg string, fields ...zap.Field) {
	logger.log.Debug(logger.config.ModuleName+msg, fields...)
}

func (logger *Logger) Panic(msg string, fields ...zap.Field) {
	logger.log.Panic(logger.config.ModuleName+msg, fields...)
}

func (logger *Logger) Binary(key string, val []byte) zap.Field {
	return zap.Binary(key, val)
}

func (logger *Logger) Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func (logger *Logger) ByteString(key string, val []byte) zap.Field {
	return zap.ByteString(key, val)
}

func (logger *Logger) Complex128(key string, val complex128) zap.Field {
	return zap.Complex128(key, val)
}

func (logger *Logger) Complex64(key string, val complex64) zap.Field {
	return zap.Complex64(key, val)
}

func (logger *Logger) Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

func (logger *Logger) Float32(key string, val float32) zap.Field {
	return zap.Float32(key, val)
}

func (logger *Logger) Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func (logger *Logger) Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func (logger *Logger) Int32(key string, val int32) zap.Field {
	return zap.Int32(key, val)
}

func (logger *Logger) Int16(key string, val int16) zap.Field {
	return zap.Int16(key, val)
}

func (logger *Logger) Int8(key string, val int8) zap.Field {
	return zap.Int8(key, val)
}

func (logger *Logger) String(key string, val string) zap.Field {
	return zap.String(key, val)
}

func (logger *Logger) Uint(key string, val uint) zap.Field {
	return zap.Uint(key, val)
}

func (logger *Logger) Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)
}

func (logger *Logger) Uint32(key string, val uint32) zap.Field {
	return zap.Uint32(key, val)
}

func (logger *Logger) Uint16(key string, val uint16) zap.Field {
	return zap.Uint16(key, val)
}

func (logger *Logger) Uint8(key string, val uint8) zap.Field {
	return zap.Uint8(key, val)
}

func (logger *Logger) Uintptr(key string, val uintptr) zap.Field {
	return zap.Uintptr(key, val)
}

func initLogger(config LoggerConfig, logLevel zap.AtomicLevel) *Logger {
	// Error及以上日志
	highw, _ := newHookLogger(config.ErrorPath, zap.ErrorLevel)
	// 设置级别及以上日志
	loww, _ := newHookLogger(config.LogPath, logLevel.Level())

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 控制台输出
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleDebugging, logLevel.Level()),
		zapcore.NewCore(consoleEncoder, loww, logLevel.Level()),
		zapcore.NewCore(consoleEncoder, highw, zap.ErrorLevel),
	)

	logger := zap.New(core)
	logger.Info(config.ModuleName + " logger init success")
	defer logger.Sync()

	return &Logger{log: logger, config: config}
}

func newHookLogger(logpath string, logLevel zapcore.Level) (zapcore.WriteSyncer, error) {
	hook := lumberjack.Logger{
		Filename:   logpath, // 日志文件路径
		MaxSize:    16,      // megabytes
		MaxBackups: 0,       // 最多保留3个备份
		MaxAge:     30,      // days
		Compress:   false,   // 是否压缩 disabled by default
		LocalTime:  true,    // 是否使用服务器本地时间，默认UTC时间
	}

	// priority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
	// 	return lvl >= logLevel
	// })

	w := zapcore.AddSync(&hook)
	return w, nil
}

func getLoggerLevel(lvl string) zapcore.Level {
	var levelMap = map[string]zapcore.Level{
		"debug":  zapcore.DebugLevel,
		"info":   zapcore.InfoLevel,
		"warn":   zapcore.WarnLevel,
		"error":  zapcore.ErrorLevel,
		"dpanic": zapcore.DPanicLevel,
		"panic":  zapcore.PanicLevel,
		"fatal":  zapcore.FatalLevel,
	}

	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}
