#!/bin/bash
# PostgreSQL Endpoints Test Script

API_BASE="http://localhost:8008"

echo "🐘 Testing PostgreSQL Endpoints"
echo "================================="

# Test pgcommand endpoint
echo ""
echo "1. Testing /pgcommand endpoint..."
curl -X POST "${API_BASE}/pgcommand" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT version()"
  }' | jq .

echo ""
echo "2. Testing /v1/pgcommand endpoint..."
curl -X POST "${API_BASE}/v1/pgcommand" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "CREATE TABLE IF NOT EXISTS test_users (id SERIAL PRIMARY KEY, name VARCHAR(100), created_at TIMESTAMP DEFAULT NOW())"
  }' | jq .

# Test pgselect endpoint
echo ""
echo "3. Testing /pgselect endpoint..."
curl -X POST "${API_BASE}/pgselect" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT current_database(), current_user, now() as current_time"
  }' | jq .

echo ""
echo "4. Testing /v1/pgselect endpoint..."
curl -X POST "${API_BASE}/v1/pgselect" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT table_name FROM information_schema.tables WHERE table_schema = '\''public'\'' LIMIT 5"
  }' | jq .

echo ""
echo "✅ PostgreSQL endpoints test completed!"
echo ""
echo "📊 Available endpoints:"
echo "  - POST /pgcommand       (Legacy)"
echo "  - POST /v1/pgcommand    (Recommended)"
echo "  - POST /pgselect        (Legacy)"
echo "  - POST /v1/pgselect     (Recommended)"
