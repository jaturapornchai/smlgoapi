# 🤖 SMLGOAPI Guide Endpoint - AI Agent Documentation

## Overview
The `/guide` endpoint provides comprehensive API documentation specifically designed for AI agents and automated systems. This endpoint returns a complete JSON structure containing all necessary information for AI agents to understand and interact with the SMLGOAPI.

## Endpoint Details
- **URL**: `GET /guide`
- **Method**: GET
- **Content-Type**: application/json
- **Authentication**: None required
- **Purpose**: Complete API documentation for AI agents

## Response Structure

The `/guide` endpoint returns a comprehensive JSON object with the following main sections:

### 1. API Metadata
```json
{
  "api_name": "SMLGOAPI",
  "version": "1.0.0",
  "description": "ClickHouse REST API with universal SQL execution...",
  "base_url": "http://localhost:8008",
  "last_updated": "2025-06-17"
}
```

### 2. Core Concepts
- **Overview**: High-level API description
- **Core Features**: Key capabilities list
- **Data Flow**: Request/response flow diagram
- **Security**: Current security status

### 3. Complete Endpoint Documentation
For each endpoint, the guide provides:
- **Method and URL**
- **Purpose and use cases**
- **Request format with examples**
- **Response format with examples**
- **Error handling information**
- **Integration examples**

#### Available Endpoints Documented:
1. `GET /health` - Health check and connectivity
2. `POST /command` - Universal SQL command execution
3. `POST /select` - SELECT query execution with data return
4. `POST /search` - Multi-step product search
5. `GET /imgproxy` - Image proxy with caching
6. `GET /api/tables` - Database table listing

### 4. AI Agent Instructions
```json
{
  "ai_agent_instructions": {
    "overview": "This API is designed to be AI-friendly...",
    "best_practices": [
      "Always check /health before executing operations",
      "Use /command for data modification",
      "Use /select for data retrieval",
      "Handle both success and error responses appropriately"
    ],
    "error_handling": {
      "all_endpoints_return": "Consistent JSON structure with success boolean",
      "error_format": {...},
      "common_errors": [...]
    }
  }
}
```

### 5. Integration Examples
The guide includes ready-to-use code examples:
- **cURL commands** for each endpoint
- **JavaScript/Fetch** examples
- **Error handling patterns**
- **Best practices for AI agents**

### 6. Production Considerations
- Security recommendations
- Performance optimization tips
- Monitoring suggestions
- Troubleshooting guidelines

## Use Cases for AI Agents

### 1. **API Discovery**
AI agents can call `/guide` to automatically discover all available endpoints and their capabilities without manual configuration.

### 2. **Dynamic Integration**
The guide provides enough detail for AI agents to generate appropriate requests for any endpoint dynamically.

### 3. **Error Handling**
Complete error format documentation helps AI agents handle failures gracefully.

### 4. **Best Practices**
Built-in recommendations ensure AI agents use the API efficiently and safely.

## Example Usage

### Basic Request
```bash
curl http://localhost:8008/guide
```

### Python Example for AI Agents
```python
import requests
import json

def get_api_guide():
    """Fetch complete API documentation for automated integration"""
    response = requests.get('http://localhost:8008/guide')
    return response.json()

def discover_endpoints(guide_data):
    """Extract all available endpoints from guide"""
    endpoints = guide_data.get('endpoints', {})
    return {name: info for name, info in endpoints.items()}

def get_best_practices(guide_data):
    """Get AI agent best practices"""
    return guide_data.get('ai_agent_instructions', {}).get('best_practices', [])

# Usage
guide = get_api_guide()
endpoints = discover_endpoints(guide)
practices = get_best_practices(guide)

print(f"Available endpoints: {list(endpoints.keys())}")
print(f"Best practices: {practices}")
```

### JavaScript Example for Web Applications
```javascript
async function loadAPIGuide() {
    try {
        const response = await fetch('/guide');
        const guide = await response.json();
        
        // Extract endpoint information
        const endpoints = guide.endpoints;
        const aiInstructions = guide.ai_agent_instructions;
        
        // Use guide data to configure API client
        return {
            baseUrl: guide.base_url,
            endpoints: endpoints,
            bestPractices: aiInstructions.best_practices
        };
    } catch (error) {
        console.error('Failed to load API guide:', error);
    }
}
```

## Benefits for AI Agents

### 1. **Self-Documenting API**
AI agents don't need external documentation - everything is available via the API itself.

### 2. **Version Awareness**
The guide includes version information and last update timestamp for version tracking.

### 3. **Consistent Format**
All endpoint documentation follows the same JSON structure for easy parsing.

### 4. **Complete Examples**
Ready-to-use request/response examples for immediate implementation.

### 5. **Error Prevention**
Best practices and common error information help prevent integration issues.

## Integration Workflow for AI Agents

1. **Initial Discovery**: Call `/guide` to understand API capabilities
2. **Health Check**: Call `/health` to verify connectivity
3. **Operation Planning**: Use guide data to plan database operations
4. **Execution**: Use appropriate endpoints (`/command` or `/select`)
5. **Error Handling**: Follow error handling patterns from guide
6. **Monitoring**: Track performance using duration metrics

## Advanced Features

### Schema Discovery
```javascript
// Get database schema information
const tablesResponse = await fetch('/api/tables');
const tables = await tablesResponse.json();

// Use schema to generate appropriate queries
const query = `SELECT * FROM ${tables[0].name} LIMIT 10`;
const dataResponse = await fetch('/select', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query })
});
```

### Dynamic Query Generation
The guide provides enough information for AI agents to:
- Generate appropriate SQL queries based on available tables
- Choose correct endpoints for different operations
- Handle responses appropriately
- Implement error recovery strategies

## Security and Best Practices

The guide includes security recommendations:
- Authentication requirements for production
- Query validation suggestions
- Rate limiting recommendations
- CORS configuration guidance

## Testing and Validation

AI agents can use the guide to:
- Validate their integration approach
- Test all available endpoints
- Verify response format handling
- Implement comprehensive error testing

## Support Information

The guide includes troubleshooting information for common issues:
- Connection problems
- Query syntax errors
- CORS issues
- Performance optimization

This makes the `/guide` endpoint a complete resource for any AI agent or automated system that needs to integrate with the SMLGOAPI.
