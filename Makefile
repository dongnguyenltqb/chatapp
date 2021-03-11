default: build delivery deploy

build:
	@docker build -t gapp .
	@docker tag gapp gnodhn/gapp:latest
delivery:
	@docker push gnodhn/gapp:latest
deploy:
	@kubectl apply -f app.yaml
	@kubectl rollout restart deployments/gapp-deployment
