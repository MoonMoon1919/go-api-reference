package server

import (
	"time"

	"github.com/moonmoon1919/go-api-reference/internal/config"
)

const ProcessChannelsBufferSize = 1

type Timeouts struct {
	Read  time.Duration
	Write time.Duration
	Idle  time.Duration
}

type Config struct {
	Port      config.Configurator
	Profiling config.Configurator
	Timeouts  Timeouts
}
