package pool

import (
	"context"
	"sync"
	"time"
)

var GlobalConfig = Configure()

var defaultGo *Pool

func init() {
	defaultGo = New(context.Background())
}

type Config struct {
	Concurrent  int
	IdleTimeout time.Duration
}

func (c *Config) Workers(max int) *Config {
	c.Concurrent = max
	return c
}

func (c *Config) Idle(time time.Duration) *Config {
	c.IdleTimeout = time
	return c
}

func Configure() *Config {
	return &Config{
		Concurrent:  1000,
		IdleTimeout: 60 * time.Second,
	}
}

type Pool struct {
	Cfg *Config

	ctx     context.Context
	cancel  context.CancelFunc
	pending chan func(ctx context.Context)
	workers chan struct{}

	mux    sync.RWMutex
	wg     sync.WaitGroup
	closed bool
}

func (g *Pool) execute(f func(ctx context.Context)) {
	f(g.ctx)
}

func (g *Pool) Do(f func(context.Context)) *Pool {
	select {
	case g.pending <- f: // block if workers are busy
	case g.workers <- struct{}{}:
		g.wg.Add(1)
		go g.loop(f)
	}
	return g
}

func (g *Pool) loop(f func(context.Context)) {
	defer g.wg.Done()
	defer func() { <-g.workers }()

	timer := time.NewTimer(g.Cfg.IdleTimeout)
	defer timer.Stop()
	for {
		g.execute(f)

		select {
		case <-timer.C:
			return
		case f = <-g.pending:
			if f == nil {
				return
			}
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(g.Cfg.IdleTimeout)
		}
	}
}

func (g *Pool) Close(grace bool) {
	g.mux.Lock()
	if g.closed {
		g.mux.Unlock()
		return
	}
	g.closed = true
	g.mux.Unlock()

	close(g.pending)
	close(g.workers)
	g.cancel()
	if grace {
		g.wg.Wait()
	}
}

func (g *Pool) Done() {
	g.mux.Lock()
	if g.closed {
		g.mux.Unlock()
		return
	}
	g.closed = true
	g.mux.Unlock()

	close(g.pending)
	close(g.workers)
	g.wg.Wait()
}

func New(ctx context.Context, cfgs ...*Config) *Pool {
	ctx, cancel := context.WithCancel(ctx)
	if len(cfgs) == 0 {
		cfgs = append(cfgs, GlobalConfig)
	}
	cfg := cfgs[0]
	gr := &Pool{
		Cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
		pending: make(chan func(context.Context)),
		workers: make(chan struct{}, cfg.Concurrent),
	}
	return gr
}

func Go(f func(ctx context.Context)) {
	defaultGo.Do(f)
}

func CloseAndWait() {
	defaultGo.Close(true)
}
