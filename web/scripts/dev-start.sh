#!/bin/bash

echo "🔥 Starting HADES-V2 Development Environment with HAProxy"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Stop any existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.dev.yml down 2>/dev/null || true

# Build and start services
echo "🏗️  Building and starting services..."
docker-compose -f docker-compose.dev.yml up --build -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check service status
echo "📊 Service status:"
docker-compose -f docker-compose.dev.yml ps

echo ""
echo "🌐 Development Environment Ready!"
echo "   Frontend: http://localhost:3000"
echo "   HAProxy: http://localhost:3000"
echo "   Hot Reload WebSocket: ws://localhost:3000/ws"
echo ""
echo "🔥 Hot swap functionality is now active!"
echo "   - Component changes will trigger hot swaps"
echo "   - Style changes will be applied immediately"
echo "   - No full page reloads required for core changes"
echo ""
echo "📝 To view logs:"
echo "   docker-compose -f docker-compose.dev.yml logs -f"
echo ""
echo "🛑 To stop:"
echo "   docker-compose -f docker-compose.dev.yml down"
