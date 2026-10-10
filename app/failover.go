package app

import (
	"bytes"
	"failover/logs"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

type Failover struct {
	Patroni struct {
		RESTApi            string
		Port               string
		FailoverEndpoint   string
		SwitchoverEndpoint string
		ClusterEndpoint    string
		HostInfo           Patroni
		ClusterInfo        PatroniCluster
		Active             bool
	}
	Composition struct {
		Path     string
		StartCmd []string
		Started  bool
	}
	Logs logs.Logs
}

func (m Model) Failover() {
	var f Failover
	m.Logs.Info("Failover signal received")
	m.State = "failover"
	m.SwitchState()
	f.Patroni.RESTApi = "http://localhost:" + m.Config.Patroni.Port
	f.Patroni.FailoverEndpoint = f.Patroni.RESTApi + "/failover"
	f.Patroni.SwitchoverEndpoint = f.Patroni.RESTApi + "/switchover"
	f.Patroni.ClusterEndpoint = f.Patroni.RESTApi + "/cluster"
	f.Composition.Path = m.Config.Composition.CompositionDir
	f.Composition.StartCmd = m.Config.Composition.Scripts.StartCMD

	f.Patroni.HostInfo = f.getHostInfo()
	f.Patroni.ClusterInfo = f.getClusterInfo()
	if m.DryRun {
		m.Logs.Info("Finished")
		return
	}
	if f.patroni().composition() {
		m.State = MASTERSTATE
	} else {
		m.State = DEADSTATE
	}

	m.SwitchState()
	return
}

func (f Failover) patroni() Failover {
	if f.Patroni.HostInfo.Role == "primary" {
		f.Logs.Info("Patroni is already primary")
		return f
	}
	endpoint := f.Patroni.FailoverEndpoint
	body := []byte(`{"candidate":"` + f.Patroni.HostInfo.Patroni.Name + `"}`)
	for _, l := range f.Patroni.ClusterInfo.Members {
		if l.Role == "leader" {
			endpoint = f.Patroni.SwitchoverEndpoint
			body = []byte(`{"leader":"` + l.Name + `","candidate":"` + f.Patroni.HostInfo.Patroni.Name + `"}`)
		}
	}
	f.Logs.Info("Candidate: " + f.Patroni.HostInfo.Patroni.Name)
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		f.Logs.Error(err.Error())
	}
	if resp != nil {
		b, _ := io.ReadAll(resp.Body)
		code := resp.StatusCode
		f.Logs.Info(strconv.Itoa(code) + " " + string(b))
	}
	f.Patroni.HostInfo = f.getHostInfo()
	f.Logs.Info(f.Patroni.HostInfo.Patroni.Name + " is now primary node")
	return f
}

func (f Failover) composition() bool {
	//if f.Patroni.HostInfo.Role != "primary" {
	//	f.Logs.Error("This patroni node is not in primary mode. Rechecking...")
	//	time.Sleep(3 * time.Second)
	//	f.Patroni.HostInfo = f.getHostInfo()
	//	f.composition()
	//
	//}
	f.Logs.Info("Starting composition")
	os.Chdir(f.Composition.Path)
	cmd := exec.Command(f.Composition.StartCmd[0], f.Composition.StartCmd[1:]...)
	out, err := cmd.Output()
	if err != nil {
		f.Logs.Warn("Cannot start composition with log:\n" + string(out))
		return true
	}
	f.Logs.Info("Composition started")
	return true
}

func (f Failover) getHostInfo() Patroni {
	hi := Patroni{}
	getHostInfo, err := http.Get(f.Patroni.RESTApi)
	if err != nil {

		f.Patroni.Active = false
		f.Logs.Error(err.Error())
		return Patroni{}
	}
	hostInfo, _ := io.ReadAll(getHostInfo.Body)
	_ = yaml.Unmarshal(hostInfo, &hi)
	return hi
}

func (f Failover) getClusterInfo() PatroniCluster {
	ci := PatroniCluster{}
	getClusterInfo, err := http.Get(f.Patroni.ClusterEndpoint)
	if err != nil {
		f.Patroni.Active = false
		f.Logs.Error(err.Error())
		return PatroniCluster{}
	}
	clusterInfo, _ := io.ReadAll(getClusterInfo.Body)
	_ = yaml.Unmarshal(clusterInfo, &ci)
	return ci
}
