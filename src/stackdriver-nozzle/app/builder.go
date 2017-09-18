package app

import (
	"context"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/cloudfoundry"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/config"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/heartbeat"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/nozzle"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/stackdriver"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/version"
	"github.com/cloudfoundry/lager"
)

type App struct {
	logger      lager.Logger
	c           *config.Config
	cfConfig    *cfclient.Config
	cfClient    *cfclient.Client
	labelMaker  nozzle.LabelMaker
	heartbeater heartbeat.Heartbeater
	bufferEmpty func() bool
}

func New(c *config.Config, logger lager.Logger) *App {
	logger.Info("version", lager.Data{"name": version.Name, "release": version.Release(), "user_agent": version.UserAgent()})
	logger.Info("arguments", c.ToData())

	metricClient, err := stackdriver.NewMetricClient()
	if err != nil {
		logger.Fatal("metricClient", err)
	}

	// Create a metricAdapter that will be used by the heartbeater
	// to send heartbeat metrics to Stackdriver. This metricAdapter
	// has its own heartbeater (with its own trigger) that writes to a logger.
	trigger := time.NewTicker(time.Duration(c.HeartbeatRate) * time.Second).C
	adapterHeartbeater := heartbeat.NewHeartbeater(logger, trigger)
	adapterHeartbeater.Start()
	metricAdapter, err := stackdriver.NewMetricAdapter(c.ProjectID, metricClient, adapterHeartbeater)
	if err != nil {
		logger.Error("metricAdapter", err)
	}

	// Create a heartbeater that will write heartbeat events to Stackdriver
	// logging and monitoring. It uses the metricAdapter created previously
	// to write to Stackdriver.
	metricHandler := heartbeat.NewMetricHandler(metricAdapter, logger, c.NozzleId, c.NozzleName, c.NozzleZone)
	trigger2 := time.NewTicker(time.Duration(c.HeartbeatRate) * time.Second).C
	heartbeater := heartbeat.NewLoggerMetricHeartbeater(metricHandler, logger, trigger2)

	cfConfig := &cfclient.Config{
		ApiAddress:        c.APIEndpoint,
		Username:          c.Username,
		Password:          c.Password,
		SkipSslValidation: c.SkipSSL}
	cfClient, err := cfclient.NewClient(cfConfig)
	if err != nil {
		logger.Error("cfClient", err)
	}

	var appInfoRepository cloudfoundry.AppInfoRepository
	if c.ResolveAppMetadata {
		appInfoRepository = cloudfoundry.NewAppInfoRepository(cfClient)
	} else {
		appInfoRepository = cloudfoundry.NullAppInfoRepository()
	}
	labelMaker := nozzle.NewLabelMaker(appInfoRepository)

	return &App{
		logger:      logger,
		c:           c,
		cfConfig:    cfConfig,
		cfClient:    cfClient,
		labelMaker:  labelMaker,
		heartbeater: heartbeater,
	}
}

func (a *App) newProducer() cloudfoundry.Firehose {
	return cloudfoundry.NewFirehose(a.cfConfig, a.cfClient, a.c.SubscriptionID)
}

func (a *App) newConsumer(ctx context.Context) (*nozzle.Nozzle, error) {
	loggingEvents := strings.Split(a.c.LoggingEvents, ",")
	metricEvents := strings.Split(a.c.MonitoringEvents, ",")

	logSink, err := nozzle.NewFilterSink(loggingEvents, nozzle.NewSinkBuffer(a.newLogSink()))
	if err != nil {
		return nil, err
	}

	metricSink, err := nozzle.NewFilterSink(metricEvents, nozzle.NewSinkBuffer(a.newMetricSink(ctx)))
	if err != nil {
		return nil, err
	}

	return &nozzle.Nozzle{
		LogSink:     logSink,
		MetricSink:  metricSink,
		Heartbeater: a.heartbeater,
	}, nil
}

func (a *App) newLogSink() nozzle.Sink {
	logAdapter, logErrs := a.newLogAdapter()
	go func() {
		err := <-logErrs
		a.logger.Error("logAdapter", err)
	}()

	return nozzle.NewLogSink(a.labelMaker, logAdapter, a.c.NewlineToken)
}

func (a *App) newLogAdapter() (stackdriver.LogAdapter, <-chan error) {
	return stackdriver.NewLogAdapter(
		a.c.ProjectID,
		a.c.BatchCount,
		time.Duration(a.c.BatchDuration)*time.Second,
		a.heartbeater,
	)
}

func (a *App) newMetricSink(ctx context.Context) nozzle.Sink {
	metricClient, err := stackdriver.NewMetricClient()
	if err != nil {
		a.logger.Fatal("metricClient", err)
	}

	metricAdapter, err := stackdriver.NewMetricAdapter(a.c.ProjectID, metricClient, a.heartbeater)
	if err != nil {
		a.logger.Error("metricAdapter", err)
	}

	metricBuffer, errs := stackdriver.NewAutoCulledMetricsBuffer(ctx, a.logger, time.Duration(a.c.MetricsBufferDuration)*time.Second, a.c.MetricsBufferSize, metricAdapter)
	a.bufferEmpty = metricBuffer.IsEmpty
	go func() {
		for err = range errs {
			a.logger.Error("metricsBuffer", err)
		}
	}()

	return nozzle.NewMetricSink(a.labelMaker, metricBuffer, nozzle.NewUnitParser())
}
