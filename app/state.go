package app

import (
	"os"
	"strings"
)

var (
	MASTERSTATE    = "active"
	STANDBYSTATE   = "standby"
	DEADSTATE      = "dead"
	UNHEALTHYSTATE = "unhealthy"
)

func (m Model) SwitchState() {
	if s := m.GetState(); s == m.State {
		return
	}
	m.Logs.Info("State: " + m.State)
	os.WriteFile(m.StateFilePath, []byte(m.State), 0644)
}

func (m Model) GetState() string {
	if state, err := os.ReadFile(m.StateFilePath); err != nil {
		m.State = STANDBYSTATE
		os.WriteFile(m.StateFilePath, []byte(m.State), 0644)
		return m.State
	} else {
		return strings.Trim(string(state), "\n")
	}
}

func (m Model) StateOnStart(s string) string {
	if state, err := os.ReadFile(m.StateFilePath); err != nil {
		m.State = s
		os.WriteFile(m.StateFilePath, []byte(m.State), 0644)
		return m.State
	} else {
		return strings.Trim(string(state), "\n")
	}
}
