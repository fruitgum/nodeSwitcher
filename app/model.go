package app

import (
	"failover/config"
	"failover/logs"
)

type Model struct {
	Config         config.Config
	Patroni        Patroni
	Healthy        bool
	Action         string
	Master         bool
	State          string
	StateFilePath  string
	CooldownActive bool
	DryRun         bool
	Logs           logs.Logs
}

type Patroni struct {
	State               string `json:"state"`
	PostmasterStartTime string `json:"postmaster_start_time"`
	Role                string `json:"role"`
	ServerVersion       int    `json:"server_version"`
	Xlog                struct {
		Location int64 `json:"location"`
	} `json:"xlog"`
	Timeline    int `json:"timeline"`
	Replication []struct {
		Usename         string `json:"usename"`
		ApplicationName string `json:"application_name"`
		ClientAddr      string `json:"client_addr"`
		State           string `json:"state"`
		SyncState       string `json:"sync_state"`
		SyncPriority    int    `json:"sync_priority"`
	} `json:"replication"`
	DcsLastSeen              int    `json:"dcs_last_seen"`
	DatabaseSystemIdentifier string `json:"database_system_identifier"`
	Patroni                  struct {
		Version string `json:"version"`
		Scope   string `json:"scope"`
		Name    string `json:"name"`
	} `json:"patroni"`
}

type PatroniCluster struct {
	Members []struct {
		Name     string `json:"name"`
		Role     string `json:"role"`
		State    string `json:"state"`
		ApiUrl   string `json:"api_url"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Timeline int    `json:"timeline"`
		Lag      int64  `json:"lag,omitempty"`
	} `json:"members"`
	Scope string `json:"scope"`
}
