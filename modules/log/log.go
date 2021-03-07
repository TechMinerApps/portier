package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is used here to de-couple logging framework
type Logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})

	// Fatal must execute os.Exit(1) after logging
	Fatalf(template string, args ...interface{})

	// Panic must execute panic() after logging
	Panicf(template string, args ...interface{})
}

type LoggerType int

// Logger Type
const (
	MACHINE LoggerType = iota
	HUMAN
)

type log struct {
	Logger Logger
}

// Config is used to config a logger
type Config struct {
	Mode       LoggerType
	OutputFile string
}

// NewLogger generates a new logger based on config
func NewLogger(c *Config) (Logger, error) {
	var l log
	switch c.Mode {
	case MACHINE:
		z, _ := zap.NewDevelopment()
		l.Logger = z.Sugar()
	case HUMAN:
		var err error

		// Write to stdout by default
		writeSyncer := zapcore.AddSync(os.Stdout)

		// Write to file if OutputFile is set
		if c.OutputFile != "" {
			writeSyncer, err = getFileLogWriter(c.OutputFile)
			if err != nil {
				return nil, err
			}
		}

		// Custonmize zap logger
		encoder := getEncoder()
		core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

		l.Logger = zap.New(core, zap.AddCaller()).Sugar()
	}
	return l.Logger, nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getFileLogWriter(path string) (zapcore.WriteSyncer, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(file), nil
}
