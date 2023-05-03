package http

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/LiliyaD/Rate_Limiter/config"
	"github.com/LiliyaD/Rate_Limiter/internal/static"
	"github.com/valyala/fasthttp"
)

type httpServer struct {
	config          configRateLimiter
	staticContent   static.StaticContentI
	requestsCounter map[uint32]rateLimiterCounter // key - subnet
	mux             sync.Mutex
}

func NewHttpServer(cfg *config.Config, staticContent static.StaticContentI) *httpServer {
	return &httpServer{
		config: configRateLimiter{
			host:            cfg.RateLimiterHost,
			mask:            net.CIDRMask(cfg.SubnetPrefixLength, 32),
			timeCooldownSec: cfg.TimeCooldownSec,
			rateLimits:      cfg.RateLimits,
		},
		staticContent:   staticContent,
		requestsCounter: make(map[uint32]rateLimiterCounter),
		mux:             sync.Mutex{},
	}
}

func (h *httpServer) Run() error {
	server := fasthttp.Server{
		Name:    "Rate Limiter",
		Handler: h.getStaticContent,
	}

	err := server.ListenAndServe(h.config.host)
	if err != nil {
		return fmt.Errorf("ListenAndServe Rate Limiter is failed, error: %v", err)
	}
	return nil
}

func (h *httpServer) getStaticContent(ctx *fasthttp.RequestCtx) {
	if !bytes.Equal(ctx.Request.Header.Method(), []byte(fasthttp.MethodGet)) {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	xForwardedFor := ctx.Request.Header.Peek("X-Forwarded-For")
	ip := net.ParseIP(string(xForwardedFor))
	if len(ip) == 0 {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	ip = ip.Mask(h.config.mask)

	h.mux.Lock()
	defer h.mux.Unlock()

	status := fasthttp.StatusOK

	// endian doesn't matter, number just should be converted by same method every time
	ipAsNum := binary.BigEndian.Uint32(ip)
	rpsCounter, ok := h.requestsCounter[ipAsNum]
	if !ok {
		h.requestsCounter[ipAsNum] = rateLimiterCounter{
			counter:   1,
			endTime:   time.Now().Add(time.Duration(h.config.rateLimits.TimeSec) * time.Second),
			timeUsage: TimeLimitEnd,
		}
	} else {
		if rpsCounter.timeUsage == TimeLimitEnd {
			if rpsCounter.endTime.Before(time.Now()) {
				h.requestsCounter[ipAsNum] = rateLimiterCounter{
					counter:   1,
					endTime:   time.Now().Add(time.Duration(h.config.rateLimits.TimeSec) * time.Second),
					timeUsage: TimeLimitEnd,
				}
			} else if rpsCounter.counter < h.config.rateLimits.RequestsCount {
				h.requestsCounter[ipAsNum] = rateLimiterCounter{
					counter:   rpsCounter.counter + 1,
					endTime:   rpsCounter.endTime,
					timeUsage: TimeLimitEnd,
				}
			} else {
				h.requestsCounter[ipAsNum] = rateLimiterCounter{
					counter:   0,
					endTime:   time.Now().Add(time.Duration(h.config.timeCooldownSec) * time.Second),
					timeUsage: CooldownEnd,
				}
				h.sendResponseTooManyRequests(ctx, time.Duration(h.config.timeCooldownSec))
				return
			}
		} else {
			if rpsCounter.endTime.After(time.Now()) {
				h.sendResponseTooManyRequests(ctx, time.Until(rpsCounter.endTime)/time.Second)
				return
			} else {
				h.requestsCounter[ipAsNum] = rateLimiterCounter{
					counter:   1,
					endTime:   time.Now().Add(time.Duration(h.config.rateLimits.TimeSec) * time.Second),
					timeUsage: TimeLimitEnd,
				}
			}
		}
	}

	body, err := h.staticContent.Get()
	if err != nil {
		status = fasthttp.StatusUnprocessableEntity
	}

	ctx.SetContentType("application/json")
	ctx.Response.SetBody(body)
	ctx.Response.SetStatusCode(status)
}

func (h *httpServer) sendResponseTooManyRequests(ctx *fasthttp.RequestCtx, sec time.Duration) {
	ctx.SetContentType("text/plain")
	ctx.Response.SetBody([]byte(fmt.Sprintf("Too many requests. Retry after %d seconds", sec)))
	ctx.Response.SetStatusCode(fasthttp.StatusTooManyRequests)
	ctx.Response.Header.Add("Retry-After", strconv.FormatInt(int64(sec), 10))
}
