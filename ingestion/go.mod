module github.com/cloud-pricer/ingestion

go 1.22

require (
	github.com/cloud-pricer/shared v0.0.0
	github.com/lib/pq v1.10.9
)

replace github.com/cloud-pricer/shared => ../shared
