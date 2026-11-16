package entity

type SubsystemHealth struct {
	Name    string
	Healthy bool
}

type HealthCheck struct {
	Subsystems []SubsystemHealth
}

func (h *HealthCheck) Healthy() bool {
	for _, subsystem := range h.Subsystems {
		if !subsystem.Healthy {
			return false
		}
	}
	return true
}
