package socks5

import (
	"context"
	"fmt"
	"github.com/qdm12/gluetun/internal/socks5/lib"
	"net"
	"sync"
	"time"

	"github.com/qdm12/gluetun/internal/configuration/settings"
	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/models"
)

type Loop struct {
	state state
	// Other objects
	logger Logger
	// Internal channels and locks
	loopLock      sync.Mutex
	running       chan models.LoopStatus
	stop, stopped chan struct{}
	start         chan struct{}
	backoffTime   time.Duration
}

func (l *Loop) logAndWait(ctx context.Context, err error) {
	if err != nil {
		l.logger.Error(err.Error())
	}
	l.logger.Info("retrying in " + l.backoffTime.String())
	timer := time.NewTimer(l.backoffTime)
	l.backoffTime *= 2
	select {
	case <-timer.C:
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
	}
}

const defaultBackoffTime = 10 * time.Second

func NewLoop(settings settings.Socks5, logger Logger) *Loop {
	return &Loop{
		state: state{
			status:   constants.Stopped,
			settings: settings,
		},
		logger:      logger,
		start:       make(chan struct{}),
		running:     make(chan models.LoopStatus),
		stop:        make(chan struct{}),
		stopped:     make(chan struct{}),
		backoffTime: defaultBackoffTime,
	}
}

func (l *Loop) Run(ctx context.Context, done chan<- struct{}) {
	defer close(done)

	crashed := false

	if *l.GetSettings().Enabled {
		go func() {
			_, _ = l.SetStatus(ctx, constants.Running)
		}()
	}

	select {
	case <-l.start:
	case <-ctx.Done():
		return
	}

	for ctx.Err() == nil {
		settings := l.GetSettings()

		addr, _ := net.ResolveTCPAddr("tcp", *settings.Address)
		server := lib.NewServer()
		server.EnableUDP()
		server.SetAuthHandle(func(username, password string) bool {
			return password == *settings.Password
		})
		server.OnConnected(func(network, address string, port int) {
			l.logger.Info(fmt.Sprintf("%s connection to: %s:%d", network, address, port))
		})

		socks5Ctx, socks5Cancel := context.WithCancel(ctx)

		waitError := make(chan error)
		go func() {
			l.logger.Info(fmt.Sprintf("listening on %s", addr.String()))
			waitError <- server.Run(socks5Ctx, addr)
		}()

		isStableTimer := time.NewTimer(time.Second)

		stayHere := true
		for stayHere {
			select {
			case <-ctx.Done():
				socks5Cancel()
				<-waitError
				close(waitError)
				return
			case <-isStableTimer.C:
				if !crashed {
					l.running <- constants.Running
					crashed = false
				} else {
					l.backoffTime = defaultBackoffTime
					l.state.setStatusWithLock(constants.Running)
				}
			case <-l.start:
				l.logger.Info("starting")
				socks5Cancel()
				<-waitError
				close(waitError)
				stayHere = false
			case <-l.stop:
				l.logger.Info("stopping")
				socks5Cancel()
				<-waitError
				close(waitError)
				l.stopped <- struct{}{}
			case err := <-waitError: // unexpected error
				socks5Cancel()
				close(waitError)
				if ctx.Err() != nil {
					return
				}
				l.state.setStatusWithLock(constants.Crashed)
				l.logAndWait(ctx, err)
				crashed = true
				stayHere = false
			}
		}
		socks5Cancel() // repetition for linter only
		if !isStableTimer.Stop() {
			<-isStableTimer.C
		}
	}
}
