package main

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"
	"sync"

	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/app"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/config"
	"github.com/cloudfoundry/lager"
)

func main() {
	logger := lager.NewLogger("stackdriver-nozzle")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal("config", err)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 9; i++ {
		logger := lager.NewLogger(fmt.Sprintf("stackdriver-nozzle.%d", i))
		logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.ERROR))
		a := app.New(cfg, logger)

		ctx := context.Background()
		wg.Add(1)
		go func() {
			app.Run(ctx, a)
			defer wg.Done()
		}()
	}

	wg.Wait()
}
