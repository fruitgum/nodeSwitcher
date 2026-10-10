package app

import (
	"failover/config"
	"failover/logs"
	"os"
	"strings"
)

var (
	Action string
	DryRun bool
	State  string
)

func New() (Model, error) {
	var m Model
	c, err := config.New()
	if m.Config.LogLevel == "" {
		m.Config.LogLevel = "info"
	}
	m.Logs = logs.New()
	m.Logs.DesireLevel = strings.ToLower(m.Config.LogLevel)
	m.Logs.DryRun = DryRun
	if err != nil {
		return m, err
	}
	m.Config = c
	m.Action = Action
	m.Master = false
	m.Healthy = true
	m.StateFilePath = "/etc/failoverd/state"
	m.DryRun = DryRun
	m.State = m.StateOnStart(State)
	_, err = os.ReadDir(m.Config.Composition.CompositionDir)
	if err != nil {
		m.Logs.Error(err.Error())
		return m, err
	}

	_, err = os.Stat(strings.Join([]string{m.Config.Composition.CompositionDir, m.Config.Composition.Scripts.StartCMD[len(m.Config.Composition.Scripts.StartCMD)-1]}, "/"))
	if err != nil {
		m.Logs.Error(err.Error())
		return m, err
	}

	_, err = os.Stat(strings.Join([]string{m.Config.Composition.CompositionDir, m.Config.Composition.Scripts.StopCMD[len(m.Config.Composition.Scripts.StopCMD)-1]}, "/"))
	if err != nil {
		m.Logs.Error(err.Error())
		return m, err
	}

	m.SwitchState()
	return m, nil
}
