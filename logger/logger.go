package logger

import (
	"kortho/config"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger
var SugarLogger *zap.SugaredLogger

func InitLogger(cfg *config.LogConfigInfo) (err error) {
	encoder := getEncoder()
	syncWriter := getLogWriter(cfg.FileName, cfg.MaxAge, cfg.MaxSize, cfg.MaxBackups)

	level := new(zapcore.Level)
	err = level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		log.Panic(err)
		return
	}

	core := zapcore.NewCore(encoder, syncWriter, zapcore.DebugLevel)
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	SugarLogger = Logger.Sugar()
	return
}

func getEncoder() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

func getLogWriter(filename string, maxAge, maxSize, maxBackups int) zapcore.WriteSyncer {
	umberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxAge:     maxAge,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
	}
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(umberJackLogger), zapcore.AddSync(os.Stdout))
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}
