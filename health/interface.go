package health

import "context"

// The RegisteredHealthService interface is used by the HealthService to ping any registered services
type RegisteredHealthService interface {
	Ping(context.Context) error
}
