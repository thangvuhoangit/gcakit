package main

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/thangvuhoangit/contrib/executor_fiber"
	"github.com/thangvuhoangit/gcakit"
)

func main() {
	myApp := gcakit.New(gcakit.Config{Name: "My App"})
	// logger := myApp.Logger()

	myApp.AddExecutor(executor_fiber.NewFiberExecutor(executor_fiber.FiberExecutorConfig{
		Name: "MyFiberExecutor",
		Addr: ":5000",
		FiberConfig: fiber.Config{
			DisableStartupMessage: true,
		},
		PreHooks: []func(ctx context.Context){
			func(ctx context.Context) {
				fmt.Println("Prehook 1")
			},
		},
	}))

	myApp.Start()

	<-myApp.WaitDone()

	myApp.Stop()
}
