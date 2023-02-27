package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
)

type MetricsContext struct {
	SendTxTotalMetrics prometheus.Counter

	SendTxMetrics prometheus.Counter
}

var metricsContext MetricsContext

func InitMetricsContext() {
	sendTxMetrics := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "sent_tx_count",
		Help:      "sent tx count",
	})

	sendTxTotalMetrics := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "sent_tx_total_count",
		Help:      "sent tx total count",
	})

	if err := prometheus.Register(sendTxMetrics); err != nil {
		logx.Error("prometheus.Register sendTxHandlerMetrics error: %v", err)
		return
	}

	if err := prometheus.Register(sendTxTotalMetrics); err != nil {
		logx.Error("prometheus.Register sendTxTotalMetrics error: %v", err)
		return
	}

	metricsContext = MetricsContext{
		SendTxMetrics:      sendTxMetrics,
		SendTxTotalMetrics: sendTxTotalMetrics,
	}
}

func MetricsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.RequestURI == "/api/v1/sendTx" {
			metricsContext.SendTxTotalMetrics.Inc()
		}
		next(writer, request)
	}
}

func SendTxMetricsInc() {
	metricsContext.SendTxMetrics.Inc()
}
