package witness

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"

	"github.com/bnb-chain/zkbnb/service/witness/config"
	"github.com/bnb-chain/zkbnb/service/witness/witness"
)

const GracefulShutdownTimeout = 5 * time.Second

var (
	generateBlockWitnessTimeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "witness_generate_time",
		Help:      "witness generate time",
	})
	scheduleNextBlockWitnessTimeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "witness_reschedule_time",
	})
)

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	err2 := registerMetrics()
	if err2 != nil {
		return err2
	}

	w, err := witness.NewWitness(c)
	if err != nil {
		panic(err)
	}
	cronJob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))
	_, err = cronJob.AddFunc("@every 2s", func() {
		logx.Info("==========start generate block witness==========")
		start := time.Now()
		err := w.GenerateBlockWitness()
		if err != nil {
			logx.Errorf("failed to generate block witness, %v", err)
		} else {
			generateBlockWitnessTimeMetric.Set(float64(time.Since(start).Milliseconds()))
			start = time.Now()
		}
		w.RescheduleBlockWitness()
		scheduleNextBlockWitnessTimeMetric.Set(float64(time.Since(start).Milliseconds()))
	})
	if err != nil {
		panic(err)
	}
	cronJob.Start()

	exit := make(chan struct{})
	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown witness......")
		<-cronJob.Stop().Done()
		w.Shutdown()
		_ = logx.Close()
		exit <- struct{}{}
	})

	logx.Info("witness cronjob is starting......")

	<-exit
	return nil
}

func registerMetrics() error {
	if err := prometheus.Register(generateBlockWitnessTimeMetric); err != nil {
		return fmt.Errorf("prometheus.Register generateBlockWitnessTimeMetric error: %v", err)
	}
	if err := prometheus.Register(scheduleNextBlockWitnessTimeMetric); err != nil {
		return fmt.Errorf("prometheus.Register scheduleNextBlockWitnessTimeMetric error: %v", err)
	}
	return nil
}
