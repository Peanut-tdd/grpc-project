package zap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ZapInterceptor() *zap.Logger {

	return InitLogger()
}

const (
	FilePath   = "./log/"
	FileSuffix = ".log"
)

var ZapLogger *zap.Logger
var fileWriter = &lazyLogWriter{}

func InitLogger() *zap.Logger {

	core := zapcore.NewCore(GetEncoder(), GetLogWriter(), zapcore.DebugLevel)

	ZapLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	//grpc_zap.ReplaceGrpcLoggerV2(ZapLogger)

	return ZapLogger

}

func GetEncoder() zapcore.Encoder {

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	return zapcore.NewJSONEncoder(encoderConfig)
}

func GetLogWriter() zapcore.WriteSyncer {
	return fileWriter
}

type lazyLogWriter struct {
	mu         sync.Mutex
	file       *os.File
	currentKey string
}

func (w *lazyLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureFile(); err != nil {
		return 0, err
	}
	// if _, err := os.Stdout.Write(p); err != nil {
	// 	return 0, err
	// }
	return w.file.Write(p)
}

func (w *lazyLogWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

func (w *lazyLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.currentKey = ""
	return err
}

func (w *lazyLogWriter) ensureFile() error {
	day := time.Now()
	dayPath := day.Format("2006-01-02")
	hourPath := fmt.Sprintf("%02d", day.Hour())
	key := fmt.Sprintf("%s/%s", dayPath, hourPath)
	if w.file != nil && w.currentKey == key {
		return nil
	}

	str := filepath.Join(dayPath, fmt.Sprintf("%s%s", hourPath, FileSuffix))
	logPath := filepath.Join(FilePath, str)

	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return err
	}

	// 先打开新文件，再关闭旧文件，避免切换时写丢失。
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	w.file = file
	w.currentKey = key
	return nil
}

func Errorf(format string, args ...interface{}) {
	ZapLogger.Sugar().Errorf(format, args...)
}

func CtxInfof(ctx context.Context, format string, args ...interface{}) {
	logger := ZapLogger.Sugar().With(traceFields(ctx)...)
	logger.Infof(format, args...)
}

func CtxDebugf(ctx context.Context, format string, args ...interface{}) {
	logger := ZapLogger.Sugar().With(traceFields(ctx)...)
	logger.Debugf(format, args...)
}

func CtxErrorf(ctx context.Context, format string, args ...interface{}) {
	logger := ZapLogger.Sugar().With(traceFields(ctx)...)
	logger.Errorf(format, args...)
}

func traceFields(ctx context.Context) []interface{} {
	if ctx == nil {
		return nil
	}
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return nil
	}
	return []interface{}{
		"trace_id", spanCtx.TraceID().String(),
		"span_id", spanCtx.SpanID().String(),
	}
}

func CloseLogger() error {
	if ZapLogger != nil {
		_ = ZapLogger.Sync()
	}
	return fileWriter.Close()
}
