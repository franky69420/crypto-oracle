package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger est une interface de journalisation personnalisée
type Logger struct {
	zap *zap.Logger
}

// NewLogger crée un nouveau logger avec le niveau spécifié
func NewLogger(level string) *Logger {
	// Configurer le niveau de log
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configuration de l'encodeur
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Configurer l'écriture des logs
	// Sortie JSON sur stdout
	jsonCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	// Créer le logger
	logger := zap.New(jsonCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		zap: logger,
	}
}

// Info enregistre un message de niveau info
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		l.zap.Info(msg, convertFields(fields[0])...)
	} else {
		l.zap.Info(msg)
	}
}

// Debug enregistre un message de niveau debug
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		l.zap.Debug(msg, convertFields(fields[0])...)
	} else {
		l.zap.Debug(msg)
	}
}

// Warning enregistre un message de niveau warning
func (l *Logger) Warning(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		l.zap.Warn(msg, convertFields(fields[0])...)
	} else {
		l.zap.Warn(msg)
	}
}

// Error enregistre un message de niveau error
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		zapFields := convertToZapFields(fields[0])
		zapFields = append([]zap.Field{zap.Error(err)}, zapFields...)
		l.zap.Error(msg, zapFields...)
	} else {
		l.zap.Error(msg, zap.Error(err))
	}
}

// Fatal enregistre un message de niveau fatal puis quitte l'application
func (l *Logger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		zapFields := convertToZapFields(fields[0])
		zapFields = append([]zap.Field{zap.Error(err)}, zapFields...)
		l.zap.Fatal(msg, zapFields...)
	} else {
		l.zap.Fatal(msg, zap.Error(err))
	}
}

// TimeTrack permet de mesurer le temps d'exécution d'une fonction
func (l *Logger) TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	l.Info(
		"Execution time",
		map[string]interface{}{
			"operation": name,
			"duration":  elapsed.String(),
		},
	)
}

// WithContext crée un nouveau logger avec un contexte
func (l *Logger) WithContext(context map[string]interface{}) *Logger {
	fields := convertToZapFields(context)
	newZap := l.zap.With(fields...)
	return &Logger{
		zap: newZap,
	}
}

// convertFields convertit une map en champs zap
func convertFields(fields map[string]interface{}) []zap.Field {
	return convertToZapFields(fields)
}

// convertToZapFields convertit une map en champs zap
func convertToZapFields(fields map[string]interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

// Sync vide les tampons et ferme le logger
func (l *Logger) Sync() {
	_ = l.zap.Sync()
}

// Sugar retourne un logger de type zap.SugaredLogger
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.zap.Sugar()
} 