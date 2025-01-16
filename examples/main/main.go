package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/thangvuhoangit/gcakit"
	"logur.dev/logur"
)

func NewHttpExecutor(name string, addr string, logger logur.Logger) *gcakit.Executor {
	e := gcakit.Executor{}
	e.Name = name

	server := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}

	e.Execute = func(ctx context.Context) error {
		logger.Info(fmt.Sprintf("Starting http server on %s", addr))
		return server.ListenAndServe()
	}

	e.Interrupt = func(err error) {
		logger.Info(fmt.Sprintf("Stopping http server on %s: %v", addr, err))
		server.Shutdown(context.Background())
	}

	return &e
}

func main() {
	myApp := gcakit.New(gcakit.Config{Name: "My App"}).WithConfig(gcakit.Config{Name: "My App Changed"})
	logger := myApp.Logger()

	myApp.WithSignal(os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	myApp.AddExecuteFunc(func() (exec func(ctx context.Context) error, cleanup func(err error)) {
		return func(ctx context.Context) error {
				logger.Info("Hello")
				return nil
			}, func(err error) {
				logger.Info("Goodbye")
			}
	})

	myApp.AddExecutor(gcakit.NewExecutor(
		"MyExecutor1",
		func(ctx context.Context) error {
			logger.Info("myExecutor1 executing... ")
			return nil
		},
		func(err error) {
			logger.Info("myExecutor1 interupting... ")
		},
	))

	myApp.AddExecutor(gcakit.NewExecutor(
		"MyExecutor2",
		func(ctx context.Context) error {
			logger.Info("myExecutor2 executing... ")
			return errors.New("myExecutor2 error")
		},
		func(err error) {
			logger.Info("myExecutor2 interupting... ")
		},
	))

	myApp.AddExecutor(NewHttpExecutor("my http server", ":8080", logger))

	myApp.Start()

	<-myApp.WaitDone()

	myApp.Stop()
}
