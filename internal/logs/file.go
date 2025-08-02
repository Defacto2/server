package logs

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	// ServerLog is the filename of the Error, Panic and Fatal level log.
	ServerLog = "defacto2_server_panics.log"
	// InfoLog is the filename of the Warn and Info level log.
	InfoLog = "defacto2_server_info.log"
	// MaxSizeMB is the maximum file size in megabytes before a log rotation is triggered.
	MaxSizeMB = 100
	// MaxBackups is the maximum number of rotated logs to keep, older logs are deleted.
	MaxBackups = 5
	// MaxDays is the maximum days a log is kept before a log rotation.
	MaxDays = 45
)

type LogFile struct {
	file *os.File
}

func NewFile(name string) (*LogFile, error) {
	const flag = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	const perm = 0666
	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &LogFile{file: file}, nil
}

func (f *LogFile) Write(r slog.Record) error {
	s := fmt.Sprintf("%s %s: %s\n", r.Time, r.Level, r.Message)
	_, err := f.file.WriteString(s)
	return err
}

func (f *LogFile) Close() error {
	return f.file.Close()
}
