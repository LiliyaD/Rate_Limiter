package http

import (
	"net"
	"time"

	"github.com/LiliyaD/Rate_Limiter/config"
)

type timeUsage int8

const (
	TimeLimitEnd timeUsage = iota
	CooldownEnd
)

type rateLimiterCounter struct {
	counter   int
	endTime   time.Time
	timeUsage timeUsage
}

type configRateLimiter struct {
	host            string
	mask            net.IPMask
	timeCooldownSec int
	rateLimits      config.RateLimits
}
