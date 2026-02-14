module github.com/egrul-system/services/search-service

go 1.23

require (
	github.com/egrul-system/services/shared v0.0.0
	github.com/elastic/go-elasticsearch/v8 v8.16.0
	github.com/gin-gonic/gin v1.10.0
	github.com/prometheus/client_golang v1.19.0
	go.uber.org/zap v1.27.0
)

replace github.com/egrul-system/services/shared => ../shared

