package logger

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/iconnor-code/cogo/core"
	configimpl "github.com/iconnor-code/cogo/core/impl/config"
)

func TestLoggerWritesToStdoutWithoutFilePath(t *testing.T) {
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = originalStdout
		_ = reader.Close()
		_ = writer.Close()
	})

	conf := &configimpl.Config{
		Config: core.Config{
			Logger: core.LoggerConfig{},
		},
	}
	logger, err := NewLogger(conf)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("stdout-only")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), `"msg":"stdout-only"`) {
		t.Fatalf("stdout output = %q, want log message", output)
	}
}
