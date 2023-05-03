package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LiliyaD/Rate_Limiter/config"
	"github.com/LiliyaD/Rate_Limiter/internal/http"
	"github.com/LiliyaD/Rate_Limiter/internal/static"
)

const (
	waitDuration = 3 * time.Second
)

type app struct {
	cfg    *config.Config
	doneCh chan struct{}
}

func NewApp(cfg *config.Config) *app {
	return &app{cfg: cfg, doneCh: make(chan struct{})}
}

func (a *app) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	staticContent := static.NewStaticContent()

	chHttpErr := make(chan error)
	go func() {
		chHttpErr <- http.NewHttpServer(a.cfg, staticContent).Run()
	}()

	select {
	case err := <-chHttpErr:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		go a.waitShootDown(waitDuration)
	}

	<-a.doneCh
	log.Println("Exited")
	return nil
}

func (a *app) waitShootDown(duration time.Duration) {
	time.Sleep(duration)
	a.doneCh <- struct{}{}
}
