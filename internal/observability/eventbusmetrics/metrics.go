package eventbusmetrics

import "github.com/prometheus/client_golang/prometheus"

var (
	publishTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "studio",
			Subsystem: "eventbus",
			Name:      "publish_total",
			Help:      "Total number of published business events.",
		},
		[]string{"backend", "topic", "event_type", "status"},
	)

	consumeTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "studio",
			Subsystem: "eventbus",
			Name:      "consume_total",
			Help:      "Total number of consumed business events.",
		},
		[]string{"backend", "group", "topic", "event_type", "status"},
	)

	outboxDispatchTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "studio",
			Subsystem: "outbox",
			Name:      "dispatch_total",
			Help:      "Total number of outbox dispatch attempts.",
		},
		[]string{"backend", "topic", "event_type", "status"},
	)
)

func init() {
	prometheus.MustRegister(publishTotal, consumeTotal, outboxDispatchTotal)
}

func RecordPublish(backend, topic, eventType, status string) {
	publishTotal.WithLabelValues(backend, topic, eventType, status).Inc()
}

func RecordConsume(backend, group, topic, eventType, status string) {
	consumeTotal.WithLabelValues(backend, group, topic, eventType, status).Inc()
}

func RecordOutboxDispatch(backend, topic, eventType, status string) {
	outboxDispatchTotal.WithLabelValues(backend, topic, eventType, status).Inc()
}
