module github.com/m-lab/go

go 1.16

// These v1 versions were published incorrectly. Retracting to prevent go mod
// from automatically selecting them.
retract [v1.0.0, v1.4.1]

require (
	cloud.google.com/go/bigquery v1.50.0
	cloud.google.com/go/datastore v1.11.0
	cloud.google.com/go/storage v1.29.0
	github.com/araddon/dateparse v0.0.0-20200409225146-d820a6159ab1
	github.com/go-test/deep v1.0.6
	github.com/googleapis/google-cloud-go-testing v0.0.0-20200911160855-bcd43fbb19e8
	github.com/kabukky/httpscerts v0.0.0-20150320125433-617593d7dcb3
	github.com/kr/pretty v0.3.0
	github.com/m-lab/uuid-annotator v0.4.1
	github.com/prometheus/client_golang v1.7.1
	golang.org/x/net v0.9.0
	golang.org/x/oauth2 v0.7.0
	google.golang.org/api v0.114.0
	google.golang.org/grpc v1.56.3 // indirect
	gopkg.in/yaml.v2 v2.2.8
)
