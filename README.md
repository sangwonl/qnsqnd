# qnsqnd

go run test/loadtest/subscriber.go -concurrency 10 -duration 10 -host localhost:8080 -topic basic-topic

HOST=localhost:8080 TOPIC=basic-topic k6 run --vus 10 --duration 7s test/loadtest/publisher_basic.js
