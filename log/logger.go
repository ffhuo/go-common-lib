package logger

import (
	"go.uber.org/zap"
	"net/http"
	"sync"
)

var (
	log  *Logger
	once sync.Once
)

func FindOrCreateLoggerInstance(config LoggerConfig) *Logger {
	once.Do(func() {
		newConfig := NewLoggerConfig(config)
		atomicLevel := waitSetLevel(newConfig.Level, newConfig.HasHTTPNet)
		log = initLogger(newConfig, atomicLevel)
	})
	return log
}

func waitSetLevel(level string, hasHTTP bool) zap.AtomicLevel {
	alevel := zap.NewAtomicLevelAt(getLoggerLevel(level))
	if hasHTTP {
		http.HandleFunc("/log/level", alevel.ServeHTTP)
		go func() {
			if err := http.ListenAndServe(":9090", nil); err != nil {
				panic(err)
			}
		}()
	}
	return alevel
}
