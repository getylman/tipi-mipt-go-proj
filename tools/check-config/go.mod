module github.com/cloud-pricer/tools/check-config

go 1.22

require (
	github.com/cloud-pricer/ingestion v0.0.0
	github.com/cloud-pricer/pricing v0.0.0
)

require github.com/cloud-pricer/shared v0.0.0 // indirect

replace (
	github.com/cloud-pricer/ingestion => ../../ingestion
	github.com/cloud-pricer/pricing => ../../pricing
	github.com/cloud-pricer/shared => ../../shared
)
