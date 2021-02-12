package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger is used here to de-couple logging framework
type logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})

	// Fatal must execute os.Exit(1) after logging
	Fatalf(template string, args ...interface{})

	// Panic must execute panic() after logging
	Panicf(template string, args ...interface{})
}

type loggerType int

// Logger Type
const (
	MACHINE loggerType = iota
	HUMAN
)

type Log struct {
	Logger logger
}

type Config struct {
	Mode       loggerType
	OutputFile string
}

// NewLogger generates a new logger based on config
func NewLogger(c *Config) (*Log, error) {
	var l Log
	switch c.Mode {
	case MACHINE:
		z, _ := zap.NewDevelopment()
		l.Logger = z.Sugar()
	case HUMAN:
		writeSyncer := zapcore.AddSync(os.Stdout)
		if c.OutputFile != "" {
			writeSyncer, _ = getFileLogWriter(c.OutputFile)
		}
		encoder := getEncoder()
		core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

		l.Logger = zap.New(core, zap.AddCaller()).Sugar()
	}
	return &l, nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getFileLogWriter(path string) (zapcore.WriteSyncer, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(file), nil
}
