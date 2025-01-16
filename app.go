package gcakit

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"emperror.dev/emperror"
	"emperror.dev/errors/match"
	"github.com/thangvuhoangit/gcakit/logger"
	"github.com/thangvuhoangit/gcakit/logger/recover/handlers"
	"github.com/thangvuhoangit/gcakit/run"
	"logur.dev/logur"
)

type Config struct {
	Name         string
	HiddenBanner bool
	Signals      []os.Signal
}

var DefaultConfig = Config{
	Name:         "my app",
	HiddenBanner: false,
	Signals:      []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT},
}

func defaultConfig(config ...Config) Config {
	if len(config) < 1 {
		return DefaultConfig
	}

	cfg := config[0]
	return cfg
}

type appcontextkey string

var appContextKey = appcontextkey("app")

type App struct {
	mutex        sync.Mutex
	config       Config
	configured   *Config
	ctx          *AppContext
	cancel       context.CancelFunc
	runGroup     *run.Group
	logger       logur.Logger
	errorHandler *handlers.Handler
}

func New(config ...Config) *App {
	cfg := defaultConfig(config...)
	ctx, cancel := signal.NotifyContext(context.Background(), cfg.Signals...)

	app := &App{config: cfg, configured: &cfg, ctx: &AppContext{Context: ctx}, cancel: cancel}

	app.init()

	return app
}

func (a *App) init() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.runGroup == nil {
		a.runGroup = new(run.Group)
	}

	a.logger = logger.NewZapLogger(logger.Config{Level: "debug", Format: "logfmt", Mode: "production", NoColor: false})
	logger.SetStandardLogger(a.logger)

	a.errorHandler = handlers.New(a.logger)
}

func (a *App) WithName(name string) *App {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.configured.Name = name
	return a
}

func (a *App) WithConfig(config Config) *App {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.configured = &config
	return a
}

func (a *App) WithContext(ctx context.Context) *App {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	appCtx := context.WithValue(ctx, appContextKey, a)

	a.ctx = &AppContext{
		Context: appCtx,
		App:     a,
	}

	return a
}

func (a *App) Context() context.Context {
	return a.ctx
}

func (a *App) Logger() logur.Logger {
	return a.logger
}

func (a *App) WithSignal(signals ...os.Signal) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.configured.Signals = signals

	a.runGroup = new(run.Group)

	ctx, cancel := signal.NotifyContext(a.ctx, a.configured.Signals...)
	a.cancel = cancel
	a.ctx = &AppContext{Context: ctx, App: a}

	a.runGroup.Add(run.SignalHandler(a.ctx, a.configured.Signals...))

	return nil
}

func (a *App) AddExecutor(executor *Executor) {
	a.runGroup.Add(func() error {
		err := executor.Execute(a.ctx)
		a.errorHandler.HandleContext(a.ctx, err)

		<-a.ctx.Done()
		return nil
	}, func(err error) {
		executor.Interrupt(err)
	})
}

func (a *App) AddExecuteFunc(handler ExecuteFunc) {
	exec, cleanup := handler()

	a.runGroup.Add(func() error {
		err := exec(a.ctx)
		a.errorHandler.HandleContext(a.ctx, err)

		<-a.ctx.Done()
		return nil
	}, func(err error) {
		cleanup(err)
	})
}

func (a *App) Start() {
	defer emperror.HandleRecover(a.errorHandler)

	if !a.configured.HiddenBanner {
		a.printStartUpMessage()
	}

	a.logger.Info(fmt.Sprintf("%s: starting...", a.configured.Name))

	err := a.runGroup.Run()
	emperror.WithFilter(a.errorHandler, match.As(&run.SignalError{}).MatchError).Handle(err)
}

func (a *App) printStartUpMessage() {
	fmt.Printf(`
   ___   ____   __    _  __  _  _____ 
  |\ _  / /    / /\  | |/ / | |  | | 
  |___| \_\__ /_/--\ |_|\_\ |_|  |_|    v%s, built with Go %s

  `, "1.0.0", runtime.Version())
	fmt.Println()
}

func (a *App) Release() {
	a.cancel()
}

func (a *App) Stop() {
	a.cancel()

	a.runGroup = nil

	a.logger.Info(fmt.Sprintf("%s: stopped", a.configured.Name))
}

func (a *App) WaitDone() <-chan struct{} {
	return a.ctx.Done()
}

func (a *App) StopWithTimeout(timeout time.Duration) {
	a.cancel()

	a.runGroup = nil

	time.Sleep(timeout)

	a.logger.Info("Closed")
}
