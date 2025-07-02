# 📋 SMLGOAPI - Complete Documentation Index

## Overview

This directory contains comprehensive documentation for all SMLGOAPI endpoints, organized by functionality and use case.

## 📁 Documentation Structure

### Core Search API

- **[search-by-vector.md](search-by-vector.md)** - Advanced product search with vector database and PostgreSQL integration

### System Monitoring

- **[health.md](health.md)** - Health check endpoint for API and database status monitoring

### Database Operations

- **[database-endpoints.md](database-endpoints.md)** - SQL query execution on ClickHouse and PostgreSQL databases

### Geographic Data

- **[thai-admin-data.md](thai-admin-data.md)** - Thai administrative data (provinces, districts, sub-districts, postal codes)

### API Information

- **[documentation-endpoints.md](documentation-endpoints.md)** - Documentation, guides, and API information endpoints

---

## 🚀 Quick Start Guide

### 1. Check API Health

```bash
curl "http://localhost:8008/v1/health"
```

### 2. Search for Products

```bash
curl -X POST "http://localhost:8008/v1/search-by-vector" \
  -H "Content-Type: application/json" \
  -d '{"query": "toyota brake", "limit": 10}'
```

### 3. Get Thai Provinces

```bash
curl -X POST "http://localhost:8008/v1/provinces" \
  -H "Content-Type: application/json" \
  -d '{}'
```

---

## 🔗 API Base Information

**Base URL:** `http://localhost:8008`  
**API Version:** `/v1`  
**Content-Type:** `application/json`  
**Authentication:** Not required

---

## 📊 Endpoint Summary

| Endpoint               | Method | Purpose                       | Documentation                                            |
| ---------------------- | ------ | ----------------------------- | -------------------------------------------------------- |
| `/v1/search-by-vector` | POST   | Product search with AI/vector | [search-by-vector.md](search-by-vector.md)               |
| `/v1/health`           | GET    | API health status             | [health.md](health.md)                                   |
| `/v1/tables`           | GET    | Database tables list          | [database-endpoints.md](database-endpoints.md)           |
| `/v1/command`          | POST   | ClickHouse SQL commands       | [database-endpoints.md](database-endpoints.md)           |
| `/v1/select`           | POST   | ClickHouse SELECT queries     | [database-endpoints.md](database-endpoints.md)           |
| `/v1/pgcommand`        | POST   | PostgreSQL SQL commands       | [database-endpoints.md](database-endpoints.md)           |
| `/v1/pgselect`         | POST   | PostgreSQL SELECT queries     | [database-endpoints.md](database-endpoints.md)           |
| `/v1/provinces`        | POST   | Thai provinces data           | [thai-admin-data.md](thai-admin-data.md)                 |
| `/v1/amphures`         | POST   | Thai districts data           | [thai-admin-data.md](thai-admin-data.md)                 |
| `/v1/tambons`          | POST   | Thai sub-districts data       | [thai-admin-data.md](thai-admin-data.md)                 |
| `/v1/findbyzipcode`    | POST   | Location by postal code       | [thai-admin-data.md](thai-admin-data.md)                 |
| `/`                    | GET    | API overview                  | [documentation-endpoints.md](documentation-endpoints.md) |
| `/v1/docs`             | GET    | API documentation             | [documentation-endpoints.md](documentation-endpoints.md) |
| `/v1/guide`            | GET    | Developer guide               | [documentation-endpoints.md](documentation-endpoints.md) |

---

## 🎯 Use Case Guides

### E-commerce Integration

1. **Product Search**: Use `/v1/search-by-vector` for intelligent product discovery
2. **Address Forms**: Use Thai administrative endpoints for shipping addresses
3. **Health Monitoring**: Monitor API availability with `/v1/health`

### Data Analysis

1. **Database Queries**: Use database endpoints for custom analytics
2. **Geographic Analysis**: Use Thai administrative data for regional insights
3. **Performance Monitoring**: Track API response times and health status

### System Integration

1. **API Discovery**: Start with documentation endpoints
2. **Health Checks**: Implement monitoring with health endpoint
3. **Data Management**: Use database endpoints for data operations

---

## 🔧 Development Tools

### Testing Scripts

Located in the project root directory:

- `test_health.ps1` - Health endpoint testing
- `test_search_coil.ps1` - Product search testing
- Various other test scripts for validation

### Example Integrations

Each documentation file includes:

- cURL examples
- JavaScript/Node.js code
- Python implementations
- PHP snippets
- PowerShell scripts

---

## 📈 Performance Guidelines

### Recommended Limits

- **Search Limit**: 20-50 results per request (max 500)
- **Concurrent Requests**: Max 10 per client
- **Request Timeout**: 30 seconds
- **Payload Size**: Max 1MB

### Response Times

- **Health Check**: ~10-50ms
- **Product Search**: ~800-1200ms
- **Database Queries**: ~50-500ms
- **Thai Admin Data**: ~50-200ms

---

## 🚨 Error Handling

### Standard Error Format

```json
{
  "success": false,
  "message": "Error description",
  "error_code": "ERROR_CODE"
}
```

### Common HTTP Status Codes

- **200**: Success
- **400**: Bad Request (invalid parameters)
- **404**: Not Found (invalid endpoint)
- **405**: Method Not Allowed (wrong HTTP method)
- **500**: Internal Server Error (database issues)

---

## 🔍 Search Tips

### Product Search Best Practices

- Use specific brand names: "toyota", "honda", "nissan"
- Include product categories: "brake", "filter", "coil"
- Support both Thai and English: "โตโยต้า" or "toyota"
- Use product codes for exact matches: "AC3006"
- Include barcodes for precise lookup: "1234567890123"

### Database Query Tips

- Use LIMIT clauses for large datasets
- Create appropriate indexes for better performance
- Use parameterized queries to prevent SQL injection
- Monitor query execution times

---

## 📝 Contributing

### Documentation Updates

When updating endpoints or adding new features:

1. Update the relevant markdown file
2. Add examples in multiple programming languages
3. Include error handling scenarios
4. Update this index file with new endpoints

### Code Examples

Follow these guidelines for code examples:

- Include complete, runnable examples
- Show both success and error handling
- Use realistic data in examples
- Provide explanatory comments

---

## 📞 Support

For technical support or questions:

- Review the appropriate documentation file
- Check the developer guide at `/v1/guide`
- Test endpoints using the provided examples
- Monitor API health status

---

**Last Updated:** July 2, 2025  
**API Version:** 1.0.0  
**Server Status:** ✅ Active on `http://localhost:8008`
