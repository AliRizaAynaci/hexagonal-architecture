package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Environment'a göre (Dev/Prod) farklı formatta log basar.
func NewLogger() *zap.Logger {
	// 1. Encoder Ayarları (Zaman formatı vs.)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"                   // "ts" yerine "timestamp" yazar
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder // Okunabilir tarih formatı (2026-02-08T...)

	// 2. Genel Config
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel), // Default INFO
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false, // Hata olduğunda stacktrace basar (Production için true olabilir)
		Sampling:          nil,
		Encoding:          "json", // Logları JSON basar (Elasticsearch dostu)
		EncoderConfig:     encoderCfg,
		OutputPaths:       []string{"stderr"}, // Standart hata çıktısına basar
		ErrorOutputPaths:  []string{"stderr"},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(), // Hangi process çalışıyor (Debug için süper)
		},
	}

	// 3. Logger'ı inşa et
	logger := zap.Must(config.Build())

	// 4. Global logger'ı da değiştir (Opsiyonel ama faydalı)
	// Böylece projenin bir yerinde yanlışlıkla zap.L() çağrılırsa senin ayarlarınla çalışır.
	zap.ReplaceGlobals(logger)

	return logger
}
