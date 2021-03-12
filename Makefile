default: build delivery deploy

build:
	@docker build -t chatapp .
	@docker tag chatapp gnodhn/chatapp:latest
delivery:
	@docker push gnodhn/chatapp:latest
deploy:
	@kubectl apply -f app.yaml
	@kubectl rollout restart deployments/chatapp-deployment

dev:
	@go build
	@./chatapp
