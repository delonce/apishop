minikube:
	minikube start

startDB: minikube
	kubectl apply -f deployments/.

test:
	go test github.com/delonce/apishop/internal/service/consumer
	go test github.com/delonce/apishop/internal/service/subject
	go test github.com/delonce/apishop/internal/delivery/handlers

run: test
	go run cmd/main.go