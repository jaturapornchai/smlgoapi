# Enhanced Product Vector Search API

🚀 **Production-ready API server with advanced vector search capabilities, multi-view image processing, and comprehensive ClickHouse integration.**

## 🌟 Features

- **Advanced Vector Search**: TF-IDF based similarity search with Thai/English support
- **Multi-View Image Processing**: Enhanced accuracy through multiple viewing angles
- **Real-time SQL Execution**: Direct ClickHouse query execution via REST API
- **Comprehensive Logging**: Request tracing, performance metrics, and debug modes
- **Production Ready**: Timeout protection, connection pooling, and health monitoring

## 🏗️ Architecture

- **Framework**: Go + Gin
- **Database**: ClickHouse
- **Vector Engine**: Custom TF-IDF implementation
- **Cache**: In-memory with TTL
- **Concurrency**: Configurable worker pools

## 📚 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | API status and information |
| `GET` | `/health` | Health check for monitoring |
| `GET` | `/help` | Complete API documentation |
| `POST` | `/search` | Product vector search |
| `GET` | `/commandget` | SQL execution via GET |
| `POST` | `/commandpost` | SQL execution via POST |
| `POST` | `/imgupload` | Image upload & processing |
| `POST` | `/imgsearch` | Image similarity search |

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- ClickHouse database
- Git

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd smlgoapi
```

2. **Install dependencies**
```bash
go mod tidy
```

3. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your ClickHouse credentials
```

4. **Run the server**
```bash
go run .
```

The server will start on `http://localhost:8008`

## ⚙️ Configuration

Create a `.env` file with your settings:

```env
# ClickHouse Configuration
CLICKHOUSE_HOST=your-clickhouse-host
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=your-username
CLICKHOUSE_PASSWORD=your-password
CLICKHOUSE_DATABASE=your-database

# Server Configuration
PORT=8008
GIN_MODE=release

# Debug Configuration
DEBUG_MODE=true
DEBUG_LEVEL=4
LOG_STEP_BY_STEP=true
LOG_SQL_EXECUTION=true
LOG_REQUEST_RESPONSE=true
LOG_PERFORMANCE=true

# Performance Configuration
MAX_WORKERS=100
SIMILARITY_THRESHOLD=0.25
CACHE_ENABLED=true
CACHE_TTL_MINUTES=15
```

## 📖 API Documentation

### Product Search
```bash
curl -X POST "http://localhost:8008/search" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "brake pad ผ้าเบรค",
    "limit": 10,
    "offset": 0
  }'
```

### SQL Command Execution
```bash
# GET method with base64 encoded query
curl -X GET "http://localhost:8008/commandget?q=$(echo 'SELECT 1' | base64)"

# POST method with JSON payload
curl -X POST "http://localhost:8008/commandpost" \
  -H "Content-Type: application/json" \
  -d '{
    "query_base64": "'$(echo 'SELECT now()' | base64)'"
  }'
```

### Image Upload & Search
```bash
# Upload image
curl -X POST "http://localhost:8008/imgupload" \
  -H "Content-Type: application/json" \
  -d '{
    "barcode": "ABC123456789",
    "imagenumber": 1,
    "image_data": "base64-encoded-image-data",
    "use_multi_view": true
  }'

# Search similar images
curl -X POST "http://localhost:8008/imgsearch" \
  -H "Content-Type: application/json" \
  -d '{
    "image_data": "base64-encoded-query-image",
    "limit": 5,
    "similarity_threshold": 0.8
  }'
```

## 🔧 Development

### Code Structure
```
├── main.go                    # Main server and configuration
├── handler_root.go           # Root endpoint handler
├── handler_health.go         # Health check handler
├── handler_help.go           # API documentation handler
├── handler_search.go         # Product search handler
├── handler_command_get.go    # SQL GET command handler
├── handler_command_post.go   # SQL POST command handler
├── handler_image_upload.go   # Image upload handler
├── handler_image_search.go   # Image search handler
├── handler_utils.go          # Shared utility functions
├── go.mod                    # Go module dependencies
└── API_TEST_RESULTS.md       # Test results documentation
```

### Running Tests
```bash
# Test all endpoints
powershell -ExecutionPolicy Bypass -File test_all_apis.ps1

# Test individual endpoints
powershell -ExecutionPolicy Bypass -File test_search.ps1
powershell -ExecutionPolicy Bypass -File test_health.ps1
```

### Debug Mode
Enable comprehensive debugging:
```bash
DEBUG_MODE=true
DEBUG_LEVEL=5
LOG_STEP_BY_STEP=true
go run .
```

## 🌐 Production Deployment

### Docker (Optional)
```bash
# Build image
docker build -t smlgoapi .

# Run container
docker run -p 8008:8008 --env-file .env smlgoapi
```

### Systemd Service
```bash
# Copy binary to /usr/local/bin/
sudo cp smlgoapi /usr/local/bin/

# Create systemd service
sudo systemctl enable smlgoapi
sudo systemctl start smlgoapi
```

## 📊 Monitoring

- **Health Check**: `GET /health`
- **Metrics**: Performance metrics available in debug mode
- **Logs**: Structured logging with request tracing
- **Debug Traces**: `GET /debug/trace` (when debug mode enabled)

## 🛡️ Security

- Input validation and sanitization
- SQL injection prevention
- Request timeout protection
- Rate limiting ready (configurable)
- Secure base64 encoding for SQL commands

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 📞 Support

- **Documentation**: Visit `/help` endpoint for complete API documentation
- **Issues**: Use GitHub Issues for bug reports and feature requests
- **Debug**: Enable debug mode for detailed logging and troubleshooting

---

⭐ **Star this repository if you find it useful!**
