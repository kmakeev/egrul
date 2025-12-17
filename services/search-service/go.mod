module github.com/egrul-system/services/search-service

go 1.23

require (
	github.com/egrul-system/services/shared v0.0.0
	github.com/elastic/go-elasticsearch/v8 v8.16.0
	github.com/gin-gonic/gin v1.10.0
	github.com/rs/zerolog v1.33.0
)

replace github.com/egrul-system/services/shared => ../shared

