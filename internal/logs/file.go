package logs

import (
	"fmt"
	"os"
)

const (
	// ServerLog is the filename of the Error, Panic and Fatal level log.
	ServerLog = "defacto2_server_panics.log"
	// InfoLog is the filename of the Warn and Info level log.
	InfoLog = "defacto2_server_info.log"
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

// Write(p []byte) (n int, err error) {
func (f *LogFile) Write(p []byte) (n int, err error) {
	if f == nil {
		return 0, nil
	}
	println("write anything?", string(p))
	fmt.Printf("log file: %+v\n", f)
	// s := fmt.Sprintf("%s %s: %s\n", r.Time, r.Level, r.Message)
	// _, err := f.file.WriteString(s)
	return 0, nil
}

func (f *LogFile) Close() error {
	return f.file.Close()
}
