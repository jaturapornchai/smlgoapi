# Variables
IMAGE_NAME = smlgoapi
TAG = latest
DOCKER_REGISTRY = ghcr.io/smlsoft

# Default target
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build       - Build Go application"
	@echo "  make run         - Run application locally"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run  - Run Docker container"
	@echo "  make deploy      - Deploy using Git"
	@echo "  make deploy-prod - Deploy to production server"
	@echo "  make deploy-full - Full deployment (Git + Production)"

# Git deployment
.PHONY: deploy
deploy:
	@echo "Deploying with Git..."
	git add .
	git commit -m "Deploying the latest changes"
	git push
	@echo "Deployment completed!"

# Production deployment
.PHONY: deploy-prod
deploy-prod:
	@echo "Deploying to production server..."
	@echo "Connecting to production server and pulling latest image..."
	ssh root@143.198.192.64 "cd /data/vectorapi-dev/ && docker pull ghcr.io/smlsoft/vectordbapi:main && docker compose up -d"
	@echo "Production deployment completed!"

# Full deployment (Git + Production)
.PHONY: deploy-full
deploy-full: deploy
	@echo "Waiting 10 seconds for CI/CD to build image..."
	sleep 10
	@make deploy-prod

# Docker commands for production
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(IMAGE_NAME):$(TAG) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -d \
		--name $(IMAGE_NAME) \
		-p 8080:8080 \
		-e SERVER_HOST=0.0.0.0 \
		-e SERVER_PORT=8080 \
		-e CLICKHOUSE_HOST=161.35.98.110 \
		-e CLICKHOUSE_PORT=9000 \
		-e CLICKHOUSE_USER=sml2 \
		-e CLICKHOUSE_PASSWORD=Md5WyoEwHfR1q6 \
		-e CLICKHOUSE_DATABASE=sml2 \
		-v ./image_cache:/root/image_cache \
		$(IMAGE_NAME):$(TAG)

.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker container..."
	-docker stop $(IMAGE_NAME)
	-docker rm $(IMAGE_NAME)

# Production Docker Compose
.PHONY: compose-up
compose-up:
	@echo "Starting production environment..."
	docker-compose up -d

.PHONY: compose-down
compose-down:
	@echo "Stopping containers..."
	docker-compose down

.PHONY: logs
logs:
	@echo "Showing container logs..."
	docker-compose logs -f

.PHONY: clean
clean:
	@echo "Cleaning up..."
	-docker-compose down -v
	-docker rmi $(IMAGE_NAME):$(TAG)
	@echo "Cleanup completed!"

# Testing
.PHONY: test
test:
	@echo "Testing API endpoints..."
	@echo "Health check:"
	curl -f http://localhost:8080/health || echo "Health check failed"
	@echo "\nAPI status check:"
	curl -f http://localhost:8080/api/v1/status || echo "Status check failed"

.PHONY: clean
clean:
	@echo "Cleaning up containers and images..."
	docker-compose down -v
	docker-compose -f docker-compose.local.yml down -v
	-docker rmi $(IMAGE_NAME):$(TAG)
	@echo "Cleanup completed!"

# Build Docker image
.PHONY: build
build:
	@echo "🔨 Building Docker image..."
	docker build -t $(IMAGE_NAME):$(TAG) .
	@echo "✅ Build completed!"

# Push to registry
.PHONY: push
push:
	@echo "📤 Pushing image to registry..."
	docker tag $(IMAGE_NAME):$(TAG) $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(TAG)
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(TAG)
	@echo "✅ Image pushed successfully!"

# Deploy to production
.PHONY: deploy
deploy:
	@echo "🚀 Deploying to production..."
	docker-compose -f docker-compose.prod.yml pull
	docker-compose -f docker-compose.prod.yml up -d
	@echo "✅ Production deployment completed!"

# Test API endpoints
.PHONY: test
test:
	@echo "🧪 Testing API endpoints..."
	@echo "Testing health check..."
	curl -f http://localhost:8008/health || echo "❌ Health check failed"
	@echo "\nTesting API documentation..."
	curl -f http://localhost:8008/ || echo "❌ API documentation failed"
	@echo "✅ Basic tests completed!"

# Check status
.PHONY: status
status:
	@echo "📊 Container status:"
	docker-compose ps
	@echo "\n📋 Container logs (last 20 lines):"
	docker-compose logs --tail=20

# Restart services
.PHONY: restart
restart: stop run
	@echo "♻️ Services restarted!"
