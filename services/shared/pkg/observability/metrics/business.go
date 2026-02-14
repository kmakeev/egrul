package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Business метрики для API Gateway
var (
	CompanySearchesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "egrul_company_searches_total",
		Help: "Total number of company searches",
	})

	CompanyExportsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "egrul_company_exports_total",
		Help: "Total number of company data exports",
	})

	GraphQLQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "egrul_graphql_queries_total",
			Help: "Total number of GraphQL queries by operation",
		},
		[]string{"operation", "status"},
	)
)

// Business метрики для Search Service
var (
	ElasticsearchQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "egrul_elasticsearch_queries_total",
			Help: "Total number of Elasticsearch queries",
		},
		[]string{"status"},
	)

	SearchResultsCount = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "egrul_search_results_count",
		Help:    "Number of search results returned",
		Buckets: []float64{0, 1, 5, 10, 50, 100, 500, 1000, 5000},
	})
)

// Business метрики для Change Detection
var (
	ChangesDetectedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "egrul_changes_detected_total",
			Help: "Total number of detected changes",
		},
		[]string{"entity_type", "change_type"},
	)

	KafkaMessagesProducedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "egrul_kafka_messages_produced_total",
			Help: "Total number of Kafka messages produced",
		},
		[]string{"topic", "status"},
	)
)

// Business метрики для Notification Service
var (
	EmailsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "egrul_emails_sent_total",
			Help: "Total number of emails sent",
		},
		[]string{"status"},
	)

	SMTPErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "egrul_smtp_errors_total",
		Help: "Total number of SMTP errors",
	})

	KafkaConsumerLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "egrul_kafka_consumer_lag",
			Help: "Kafka consumer lag",
		},
		[]string{"topic", "partition"},
	)
)
