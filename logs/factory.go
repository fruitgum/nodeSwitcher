package logs

import (
	"os"
	"strings"
	"time"
)

const (
	logFilePath = "/var/log/failoverd.log"
)

type Logs struct {
	Timestamp   string
	Level       string
	DesireLevel string
	Message     string
	MapLevel    map[string]int
	Show        int
	DryRun      bool
}

var mapLevel = map[string]int{
	"debug": 4,
	"info":  3,
	"warn":  2,
	"error": 1,
	"fatal": 0,
}

func New() Logs {

	return Logs{
		MapLevel: mapLevel,
	}
}

func (logs Logs) Debug(message string) {
	logs.Message = message
	logs.Level = "debug"
	logs.Print()

}
func (logs Logs) Info(message string) {
	logs.Message = message
	logs.Level = "info"
	logs.Print()
}
func (logs Logs) Warn(message string) {
	logs.Message = message
	logs.Level = "warn"
	logs.Print()
}
func (logs Logs) Error(message string) {
	logs.Message = message
	logs.Level = "error"
	logs.Print()
}
func (logs Logs) Fatal(message string) {
	logs.Message = message
	logs.Level = "fatal"
	logs.Print()
}

func (logs Logs) hide() bool {
	logs.Show = logs.MapLevel[logs.DesireLevel]
	if logs.Show+1 < logs.MapLevel[logs.Level] {
		return false
	}
	return true
}

func (logs Logs) constructor() string {
	delimiter := " - "
	timeStamp := time.Now().Format("2006-01-02 15:04:05")
	level := "[" + strings.ToUpper(logs.Level) + "]"
	message := logs.Message
	if logs.DryRun {
		message = "(dry run) " + message
	}
	entry := timeStamp + delimiter + level + delimiter + message
	return entry
}

func (logs Logs) Print() {
	//if logs.hide() {
	//	return
	//}
	f, _ := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	f.WriteString(logs.constructor() + "\n")
}
