package app

import (
	"encoding/json"
	"failover/logs"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	emergTimeout = 120
	tickRateSec  = int64(30)
)

type ForceStopCandidates struct {
	Containers []string
	Logs       logs.Logs
}

type Health struct {
	Patroni        bool
	Composition    bool
	Check          bool
	Containers     []string
	ContainerState ContainerState
	Logs           logs.Logs
}

type ContainerState struct {
	Name       string
	isHealthy  bool
	Status     string    `json:"Status"`
	Running    bool      `json:"Running"`
	Paused     bool      `json:"Paused"`
	Restarting bool      `json:"Restarting"`
	OOMKilled  bool      `json:"OOMKilled"`
	Dead       bool      `json:"Dead"`
	Pid        int       `json:"Pid"`
	ExitCode   int       `json:"ExitCode"`
	Error      string    `json:"Error"`
	StartedAt  time.Time `json:"StartedAt"`
	FinishedAt time.Time `json:"FinishedAt"`
	Health     struct {
		Status        string `json:"Status"`
		FailingStreak int    `json:"FailingStreak"`
		Log           []struct {
			Start    time.Time `json:"Start"`
			End      time.Time `json:"End"`
			ExitCode int       `json:"ExitCode"`
			Output   string    `json:"Output"`
		} `json:"Log"`
	} `json:"Health"`
}

func (m Model) HealthCheck() {
	s := m.GetState()

	if s != MASTERSTATE {
		//m.Logs.Warn("Cannot perform health check on " + s + " node")
		return

	}

	if m.Wait() {
		return
	}
	h := Health{
		Composition: true,
		Containers:  m.Config.Composition.Containers,
		Logs:        m.Logs,
	}
	hPatroni := h.patroni()
	hComposition := h.composition()

	if !hPatroni || !hComposition {
		m.Logs.Warn("Node is unhealthy")

		if m.DryRun {
			m.Logs.Info("Finished")
			return
		}

		m.State = UNHEALTHYSTATE
		m.SwitchState()
		m.Healthy = false
		m.shutdown()
		os.Exit(1)
	}
	m.Logs.Info("Node is healthy")
	os.Exit(0)
}

func (m Model) Wait() bool {
	if m.DryRun {
		return false
	}
	f, _ := os.Stat(m.StateFilePath)
	mdate := f.ModTime()
	duration := time.Now().Sub(mdate).Round(time.Second).Seconds()
	timeLeft := m.Config.Cooldown - duration
	if timeLeft > 0 {
		m.Logs.Info(fmt.Sprintf("Cooldown - %v seconds left", int(timeLeft)))
		return true
	}
	return false
}

func (h Health) patroni() bool {
	cmd := exec.Command("systemctl", "is-active", "patroni.service")
	out, _ := cmd.Output()
	if string(out) != "active\n" {
		h.Patroni = false
		return false
	}
	return true
}

func (h Health) composition() bool {
	h.Logs.Info("Composition healthcheck")

	if len(h.Containers) == 0 {
		h.Logs.Warn("No containers defined in config. Healthcheck aborted")
		return true
	}

	var wg sync.WaitGroup
	wg.Add(len(h.Containers))
	for _, container := range h.Containers {
		go func() {
			defer wg.Done()
			var c ContainerState
			c.Name = container
			c.isHealthy = true
			hc, hs := c.containerHealth()
			if !hc {
				h.Composition = false
			}
			if hs != "" {
				h.Logs.Warn(hs)
			}
		}()
	}
	wg.Wait()
	return h.Composition
}

func (c ContainerState) containerHealth() (bool, string) {
	command := exec.Command("sh", "-c", "docker inspect "+c.Name+" --format '{{json .State}}' 2>&1")
	out, err := command.CombinedOutput()
	if err != nil {
		return true, strings.Trim(string(out), "\n")
	}

	err = json.Unmarshal(out, &c)
	if err != nil {
		return true, err.Error()
	}

	isHealthy := c.isHealthy

	if c.Status != "running" {
		isHealthy = false
	}
	if c.Health.Status != "" && c.Health.Status != "healthy" {
		isHealthy = false
	}
	if c.ExitCode != 0 {
		isHealthy = false
	}
	if !isHealthy {
		c.isHealthy = false
		status := c.Name + " is " + c.Health.Status + ". Exit code " + strconv.Itoa(c.ExitCode)
		if len(c.Health.Log) != 0 {
			status += " (error message: " + strings.Trim(c.Health.Log[0].Output, "\n") + ")."
		}
		status += " Status: "
		switch true {
		case c.Restarting:
			return false, status + "restarting"
		case c.OOMKilled:
			return false, status + "oom killed"
		case c.Dead:
			return false, status + "dead"
		case c.Running:
			return false, status + "not running"
		default:
			return false, status + " " + c.Status
		}
	}
	return true, ""
}

func (m Model) shutdown() {
	m.Logs.Info("Shutting down composition")
	_ = os.Chdir(m.Config.Composition.CompositionDir)

	f := ForceStopCandidates{
		Containers: m.Config.Composition.Containers,
		Logs:       m.Logs,
	}

	var wg sync.WaitGroup
	isDown := make(chan bool)
	wg.Add(2)
	go func() {
		defer wg.Done()
		isDown <- false
		cmdS := m.Config.Composition.Scripts.StopCMD
		cmd := exec.Command(cmdS[0], cmdS[1:]...)
		_, err := cmd.Output()
		if err != nil {
			m.Logs.Warn(err.Error())
		}
		isDown <- true
	}()
	go func() {
		defer wg.Done()
		for {
			if !<-isDown {
				time.Sleep(time.Duration(tickRateSec) * time.Second)
				emergTimeout = emergTimeout - int(tickRateSec)
				m.Logs.Info("Waiting for graceful shutdown - " + strconv.Itoa(emergTimeout) + " seconds")
				if emergTimeout < 1 {
					m.Logs.Warn("Composition didn't shutdown gracefully. Forcing to shutdown")
					isDown <- f.forceStop()
				}
			} else {
				break
			}
		}
	}()
	wg.Wait()

	m.Logs.Info("Done")
}

func (f ForceStopCandidates) forceStop() bool {
	var wg sync.WaitGroup
	wg.Add(len(f.Containers))
	for _, container := range f.Containers {
		go func() {
			defer wg.Done()
			f.Logs.Info("Shutting down " + container)
			cmd := exec.Command("sh", "-c", "docker rm --force "+container)
			_ = cmd.Run()
			f.Logs.Info(container + " is down")
		}()
	}
	wg.Wait()
	return true
}
