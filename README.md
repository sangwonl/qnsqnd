# qnsqnd

go run test/loadtest/subscriber.go -concurrency 500 -duration 20 -host localhost:8080 -topic basic-topic

HOST=localhost:8080 TOPIC=basic-topic k6 run --vus 20 --duration 5s test/loadtest/publisher_basic.js
