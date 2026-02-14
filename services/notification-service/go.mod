module github.com/egrul/notification-service

go 1.22

require (
	github.com/egrul-system/services/shared v0.0.0
	github.com/go-chi/chi/v5 v5.0.12
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.19.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/spf13/viper v1.18.2
	go.uber.org/zap v1.27.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

replace github.com/egrul-system/services/shared => ../shared
