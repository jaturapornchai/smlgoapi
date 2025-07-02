# 🏥 `/health` API Documentation

## Overview

Health check endpoint to monitor API and database connection status.

## Endpoint Details

**URL:** `GET /v1/health`  
**Method:** `GET`  
**Content-Type:** `application/json`  
**Base URL:** `http://localhost:8008`

---

## 🚀 Usage Examples

### Basic Health Check

```bash
curl "http://localhost:8008/v1/health"
```

### PowerShell Example

```powershell
Invoke-RestMethod -Uri "http://localhost:8008/v1/health" -Method Get
```

### JavaScript/Node.js

```javascript
const checkHealth = async () => {
  const response = await fetch("http://localhost:8008/v1/health");
  return await response.json();
};

// Usage
const health = await checkHealth();
console.log(health.status);
```

### Python

```python
import requests

def check_health():
    url = "http://localhost:8008/v1/health"
    response = requests.get(url)
    return response.json()

# Usage
health = check_health()
print(f"Status: {health['status']}")
```

---

## 📊 Response Format

### Success Response (200 OK)

```json
{
  "status": "healthy",
  "timestamp": "2025-07-02T07:14:06.5541283+07:00",
  "version": "ClickHouse: 25.5.1.2782, PostgreSQL: PostgreSQL 16.9 (Debian 16.9-1.pgdg120+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit",
  "database": "connected"
}
```

### Error Response (500)

```json
{
  "success": false,
  "error": "Database connection failed: connection refused"
}
```

---

## 📋 Response Fields

| Field       | Type   | Description                                      |
| ----------- | ------ | ------------------------------------------------ |
| `status`    | string | Overall health status ("healthy" or "unhealthy") |
| `timestamp` | string | Current timestamp in ISO format                  |
| `version`   | string | Database versions (ClickHouse and PostgreSQL)    |
| `database`  | string | Database connection status                       |

---

## 🔧 Use Cases

### Monitoring & Alerting

```javascript
// Regular health monitoring
setInterval(async () => {
  try {
    const health = await checkHealth();
    if (health.status !== "healthy") {
      console.log("⚠️ API health issue:", health);
      // Send alert
    }
  } catch (error) {
    console.log("❌ API unreachable:", error);
    // Send critical alert
  }
}, 30000); // Check every 30 seconds
```

### Load Balancer Health Check

```bash
# Use in load balancer configuration
curl -f "http://localhost:8008/v1/health" || exit 1
```

---

## 📈 Performance

- **Response Time:** ~5-50ms
- **No Authentication Required**
- **Lightweight Operation**
- **Database Connection Test Included**
