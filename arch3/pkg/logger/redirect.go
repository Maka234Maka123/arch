package logger

import (
	"io"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// StdoutRedirector 标准输出重定向器
type StdoutRedirector struct {
	originalStdout *os.File
	originalStderr *os.File
	logWriter      io.Writer
}

// NewStdoutRedirector 创建标准输出重定向器
func NewStdoutRedirector(cfg *Config) (*StdoutRedirector, error) {
	r := &StdoutRedirector{
		originalStdout: os.Stdout,
		originalStderr: os.Stderr,
	}

	// 创建一个多写入器，同时写入标准输出和日志文件
	if cfg.Output == "both" || cfg.Output == "file" {
		fileWriter := getFileWriter(cfg, cfg.Filename)
		if cfg.Output == "both" {
			r.logWriter = io.MultiWriter(os.Stdout, &syncWriter{ws: fileWriter})
		} else {
			r.logWriter = &syncWriter{ws: fileWriter}
		}
	} else {
		r.logWriter = os.Stdout
	}

	return r, nil
}

// syncWriter 包装 zapcore.WriteSyncer 为 io.Writer
type syncWriter struct {
	ws zapcore.WriteSyncer
}

func (w *syncWriter) Write(p []byte) (n int, err error) {
	return w.ws.Write(p)
}

// RedirectStdLog 重定向标准库 log 到 zap
func RedirectStdLog() {
	// 将标准库的 log 重定向到 zap
	log.SetFlags(0)
	log.SetOutput(&zapStdLogWriter{logger: L()})
}

// zapStdLogWriter 实现 io.Writer 接口，将标准库 log 输出到 zap
type zapStdLogWriter struct {
	logger *zap.Logger
}

func (w *zapStdLogWriter) Write(p []byte) (n int, err error) {
	// 移除末尾的换行符
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.logger.Info(msg, zap.String("source", "std_log"))
	return len(p), nil
}

// TeeWriter 创建一个 tee writer，同时写入多个目标
type TeeWriter struct {
	writers []io.Writer
}

// NewTeeWriter 创建一个 tee writer
func NewTeeWriter(writers ...io.Writer) *TeeWriter {
	return &TeeWriter{writers: writers}
}

// Write 实现 io.Writer 接口
func (t *TeeWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			return
		}
	}
	return len(p), nil
}

// AddWriter 添加写入器
func (t *TeeWriter) AddWriter(w io.Writer) {
	t.writers = append(t.writers, w)
}

// GetMultiWriter 获取同时输出到日志和标准输出的 Writer
func GetMultiWriter(cfg *Config) io.Writer {
	if cfg.Output == "both" {
		fileWriter := getFileWriter(cfg, cfg.Filename)
		return io.MultiWriter(os.Stdout, &syncWriter{ws: fileWriter})
	} else if cfg.Output == "file" {
		fileWriter := getFileWriter(cfg, cfg.Filename)
		return &syncWriter{ws: fileWriter}
	}
	return os.Stdout
}
