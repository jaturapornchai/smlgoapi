# 📚 Documentation Endpoints API

## Overview

API endpoints for accessing documentation, guides, and general information about the SMLGOAPI.

## Base URL

`http://localhost:8008`

---

## 📋 Available Endpoints

### 1. GET `/`

Root endpoint with API overview and welcome message.

#### Usage Examples

```bash
curl "http://localhost:8008/"
```

```javascript
const getApiInfo = async () => {
  const response = await fetch("http://localhost:8008/");
  return await response.json();
};
```

#### Response Format

```json
{
  "success": true,
  "message": "Welcome to SMLGOAPI",
  "version": "1.0.0",
  "description": "Advanced product search and data management API",
  "endpoints": {
    "health": "/v1/health",
    "search": "/v1/search-by-vector",
    "documentation": "/v1/docs",
    "guide": "/v1/guide"
  },
  "features": [
    "Vector-based product search",
    "Thai administrative data",
    "Multiple database support",
    "RESTful API design"
  ]
}
```

---

### 2. GET `/v1/docs`

Get comprehensive API documentation.

#### Usage Examples

```bash
curl "http://localhost:8008/v1/docs"
```

```javascript
const getDocs = async () => {
  const response = await fetch("http://localhost:8008/v1/docs");
  return await response.json();
};
```

#### Response Format

```json
{
  "success": true,
  "data": {
    "title": "SMLGOAPI Documentation",
    "version": "1.0.0",
    "base_url": "http://localhost:8008/v1",
    "endpoints": [
      {
        "path": "/search-by-vector",
        "method": "POST",
        "description": "Advanced product search using vector database",
        "parameters": [
          {
            "name": "query",
            "type": "string",
            "required": true,
            "description": "Search term"
          }
        ]
      }
    ],
    "examples": {
      "product_search": {
        "request": {
          "query": "toyota brake",
          "limit": 10
        },
        "response": {
          "success": true,
          "data": "..."
        }
      }
    }
  }
}
```

---

### 3. GET `/v1/guide`

Get developer guide with tutorials and best practices.

#### Usage Examples

```bash
curl "http://localhost:8008/v1/guide"
```

```python
def get_guide():
    url = "http://localhost:8008/v1/guide"
    response = requests.get(url)
    return response.json()

# Usage
guide = get_guide()
print(guide['data']['getting_started'])
```

#### Response Format

```json
{
  "success": true,
  "data": {
    "title": "SMLGOAPI Developer Guide",
    "sections": {
      "getting_started": {
        "description": "Quick start guide for new developers",
        "steps": [
          "1. Check API health: GET /v1/health",
          "2. Test product search: POST /v1/search-by-vector",
          "3. Review documentation: GET /v1/docs"
        ]
      },
      "authentication": {
        "description": "Authentication is not required for current version",
        "note": "All endpoints are publicly accessible"
      },
      "rate_limits": {
        "description": "Recommended usage limits",
        "limits": {
          "requests_per_minute": 60,
          "concurrent_requests": 10,
          "max_payload_size": "1MB"
        }
      },
      "best_practices": [
        "Use specific search terms for better results",
        "Implement client-side caching",
        "Handle errors gracefully",
        "Monitor API health endpoints"
      ],
      "examples": {
        "basic_search": {
          "curl": "curl -X POST 'http://localhost:8008/v1/search-by-vector' -d '{\"query\":\"toyota\"}'",
          "javascript": "fetch('/v1/search-by-vector', {method: 'POST', body: JSON.stringify({query: 'toyota'})})",
          "python": "requests.post('/v1/search-by-vector', json={'query': 'toyota'})"
        }
      }
    }
  }
}
```

---

## 🔧 Integration Examples

### API Explorer

```html
<!DOCTYPE html>
<html>
  <head>
    <title>SMLGOAPI Explorer</title>
  </head>
  <body>
    <div id="api-info"></div>
    <div id="endpoints"></div>

    <script>
      // Load API information
      async function loadApiInfo() {
        try {
          const response = await fetch("/");
          const apiInfo = await response.json();

          document.getElementById("api-info").innerHTML = `
                    <h1>${apiInfo.message}</h1>
                    <p>Version: ${apiInfo.version}</p>
                    <p>${apiInfo.description}</p>
                `;

          // Load available endpoints
          const endpointsList = Object.entries(apiInfo.endpoints)
            .map(
              ([name, path]) =>
                `<li><a href="${path}">${name}: ${path}</a></li>`
            )
            .join("");

          document.getElementById("endpoints").innerHTML = `
                    <h2>Available Endpoints:</h2>
                    <ul>${endpointsList}</ul>
                `;
        } catch (error) {
          console.error("Failed to load API info:", error);
        }
      }

      // Load on page load
      window.addEventListener("load", loadApiInfo);
    </script>
  </body>
</html>
```

### Documentation Generator

```javascript
class APIDocumentationGenerator {
  constructor(baseUrl = "http://localhost:8008") {
    this.baseUrl = baseUrl;
  }

  async generateDocs() {
    try {
      const [apiInfo, docs, guide] = await Promise.all([
        fetch(`${this.baseUrl}/`).then((r) => r.json()),
        fetch(`${this.baseUrl}/v1/docs`).then((r) => r.json()),
        fetch(`${this.baseUrl}/v1/guide`).then((r) => r.json()),
      ]);

      return {
        overview: apiInfo,
        documentation: docs.data,
        guide: guide.data,
      };
    } catch (error) {
      throw new Error(`Failed to generate docs: ${error.message}`);
    }
  }

  async generateMarkdown() {
    const docs = await this.generateDocs();

    let markdown = `# ${docs.overview.message}\n\n`;
    markdown += `${docs.overview.description}\n\n`;
    markdown += `**Version:** ${docs.overview.version}\n\n`;

    // Add endpoints
    markdown += `## Available Endpoints\n\n`;
    Object.entries(docs.overview.endpoints).forEach(([name, path]) => {
      markdown += `- **${name}**: \`${path}\`\n`;
    });

    // Add features
    markdown += `\n## Features\n\n`;
    docs.overview.features.forEach((feature) => {
      markdown += `- ${feature}\n`;
    });

    return markdown;
  }
}

// Usage
const docGen = new APIDocumentationGenerator();
const markdown = await docGen.generateMarkdown();
console.log(markdown);
```

### Python Documentation Client

```python
import requests
import json

class APIDocumentationClient:
    def __init__(self, base_url="http://localhost:8008"):
        self.base_url = base_url

    def get_api_overview(self):
        """Get API overview and basic information"""
        response = requests.get(f"{self.base_url}/")
        return response.json()

    def get_documentation(self):
        """Get detailed API documentation"""
        response = requests.get(f"{self.base_url}/v1/docs")
        return response.json()

    def get_guide(self):
        """Get developer guide"""
        response = requests.get(f"{self.base_url}/v1/guide")
        return response.json()

    def print_api_summary(self):
        """Print a summary of the API"""
        overview = self.get_api_overview()

        print(f"API: {overview['message']}")
        print(f"Version: {overview['version']}")
        print(f"Description: {overview['description']}")
        print("\nAvailable Endpoints:")

        for name, path in overview['endpoints'].items():
            print(f"  - {name}: {path}")

        print("\nFeatures:")
        for feature in overview['features']:
            print(f"  - {feature}")

    def save_documentation(self, filename="api_docs.json"):
        """Save complete documentation to file"""
        docs = {
            'overview': self.get_api_overview(),
            'documentation': self.get_documentation(),
            'guide': self.get_guide()
        }

        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(docs, f, indent=2, ensure_ascii=False)

        print(f"Documentation saved to {filename}")

# Usage
client = APIDocumentationClient()
client.print_api_summary()
client.save_documentation()
```

---

## 🎯 Use Cases

### API Discovery

- New developers exploring available endpoints
- Understanding API capabilities and features
- Getting started with integration

### Documentation Generation

- Automated documentation updates
- API reference generation
- Integration guides for different languages

### Development Tools

- API testing tools
- Postman collection generation
- SDK documentation

### Monitoring & Analytics

- API usage tracking
- Feature adoption analysis
- Developer onboarding metrics

---

## 📊 Response Formats

### Standard Success Response

```json
{
  "success": true,
  "data": {
    "title": "Content Title",
    "content": "Detailed content here",
    "metadata": {
      "version": "1.0.0",
      "last_updated": "2025-07-02T10:00:00Z"
    }
  }
}
```

### Error Response

```json
{
  "success": false,
  "message": "Documentation not available",
  "error_code": "DOC_NOT_FOUND"
}
```

---

## 🔧 Features

### Dynamic Content

- Real-time endpoint discovery
- Version-specific documentation
- Environment-specific examples

### Multiple Formats

- JSON responses for programmatic access
- Human-readable documentation
- Code examples in multiple languages

### Interactive Elements

- Live API testing capabilities
- Example requests and responses
- Error handling demonstrations

---

## 📈 Performance

- **Response Time:** ~10-50ms
- **Content Caching:** Static documentation cached
- **No Authentication Required**
- **Lightweight Responses**

---

## 🔍 Content Structure

### Overview Section

- API description and purpose
- Version information
- Available endpoints summary
- Key features list

### Documentation Section

- Detailed endpoint specifications
- Request/response formats
- Parameter descriptions
- Example usage

### Guide Section

- Getting started tutorials
- Best practices
- Integration examples
- Troubleshooting tips
