package app

func (m Model) Router() {
	switch m.Action {
	case "healthcheck":
		m.HealthCheck()
	case "failover":
		m.Failover()
	default:
		m.Logs.Error("Invalid Action")
		return
	}
}
