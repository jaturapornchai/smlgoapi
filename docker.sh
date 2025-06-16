#!/bin/bash

# SMLGOAPI Docker Management Script

set -e

PROJECT_NAME="smlgoapi"
IMAGE_NAME="smlgoapi:latest"

show_help() {
    echo "SMLGOAPI Docker Management Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build       Build Docker image"
    echo "  run         Run container with ClickHouse"
    echo "  stop        Stop all services"
    echo "  logs        Show application logs"
    echo "  clean       Remove containers and images"
    echo "  status      Show container status"
    echo "  shell       Open shell in running container"
    echo "  help        Show this help message"
    echo ""
}

build() {
    echo "🏗️  Building Docker image..."
    docker build -t $IMAGE_NAME .
    echo "✅ Build complete!"
}

run() {
    echo "🚀 Starting services..."
    docker-compose up -d
    echo "✅ Services started!"
    echo "📊 API: http://localhost:8080"
    echo "🗄️  ClickHouse: http://localhost:8123"
    echo "💚 Health: http://localhost:8080/health"
}

stop() {
    echo "⏹️  Stopping services..."
    docker-compose down
    echo "✅ Services stopped!"
}

logs() {
    echo "📋 Showing logs..."
    docker-compose logs -f smlgoapi
}

clean() {
    echo "🧹 Cleaning up..."
    docker-compose down --rmi all --volumes --remove-orphans
    docker image rm $IMAGE_NAME 2>/dev/null || true
    echo "✅ Cleanup complete!"
}

status() {
    echo "📊 Container status:"
    docker-compose ps
}

shell() {
    echo "🐚 Opening shell in container..."
    docker-compose exec smlgoapi sh
}

case ${1:-help} in
    build)
        build
        ;;
    run)
        run
        ;;
    stop)
        stop
        ;;
    logs)
        logs
        ;;
    clean)
        clean
        ;;
    status)
        status
        ;;
    shell)
        shell
        ;;
    help|*)
        show_help
        ;;
esac
