package common

type HealthCheckResponse struct {
	// Boolean indicating if the service is healthy
	Healthy bool
	// Subsystem name -> boolean indicating if the subsystem is healthy
	Subsystems map[string]bool
}
