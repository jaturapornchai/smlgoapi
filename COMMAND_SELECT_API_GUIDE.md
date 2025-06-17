# Command and Select API Endpoints Guide

## Overview
The SMLGOAPI now provides two powerful universal SQL endpoints that allow you to execute any ClickHouse SQL commands and queries via JSON POST requests from your frontend.

## Endpoints

### 1. `/command` - Execute SQL Commands
**Method:** POST  
**Content-Type:** application/json  
**Purpose:** Execute any SQL command (INSERT, UPDATE, DELETE, CREATE, ALTER, etc.)

#### Request Format
```json
{
  "query": "INSERT INTO products (name, price) VALUES ('New Product', 29.99)"
}
```

#### Response Format
```json
{
  "success": true,
  "message": "Command executed successfully",
  "result": {
    "rows_affected": 1
  },
  "command": "INSERT INTO products (name, price) VALUES ('New Product', 29.99)",
  "duration": 125.5
}
```

#### Error Response
```json
{
  "success": false,
  "error": "Command execution failed: syntax error near 'INVALID'",
  "command": "INVALID SQL COMMAND",
  "duration": 12.3
}
```

### 2. `/select` - Execute SELECT Queries
**Method:** POST  
**Content-Type:** application/json  
**Purpose:** Execute SELECT queries and retrieve data

#### Request Format
```json
{
  "query": "SELECT name, price, category FROM products WHERE price > 50 LIMIT 10"
}
```

#### Response Format
```json
{
  "success": true,
  "message": "Query executed successfully, 5 rows returned",
  "data": [
    {
      "name": "Premium Product",
      "price": 99.99,
      "category": "Electronics"
    },
    {
      "name": "Luxury Item",
      "price": 159.99,
      "category": "Fashion"
    }
  ],
  "query": "SELECT name, price, category FROM products WHERE price > 50 LIMIT 10",
  "row_count": 5,
  "duration": 87.2
}
```

#### Error Response
```json
{
  "success": false,
  "error": "Query execution failed: table 'nonexistent' doesn't exist",
  "query": "SELECT * FROM nonexistent",
  "duration": 15.1
}
```

## Example Use Cases

### 1. Database Maintenance
```bash
# Create a new table
curl -X POST http://localhost:8008/command \
  -H "Content-Type: application/json" \
  -d '{
    "query": "CREATE TABLE test_table (id UInt32, name String, created_at DateTime) ENGINE = MergeTree() ORDER BY id"
  }'
```

### 2. Data Insertion
```bash
# Insert data
curl -X POST http://localhost:8008/command \
  -H "Content-Type: application/json" \
  -d '{
    "query": "INSERT INTO products (name, price, category) VALUES ('"'"'Gaming Laptop'"'"', 1299.99, '"'"'Electronics'"'"')"
  }'
```

### 3. Data Retrieval
```bash
# Get data with filtering
curl -X POST http://localhost:8008/select \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT * FROM products WHERE category = '"'"'Electronics'"'"' ORDER BY price DESC LIMIT 5"
  }'
```

### 4. Analytics Queries
```bash
# Get aggregated data
curl -X POST http://localhost:8008/select \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT category, COUNT(*) as product_count, AVG(price) as avg_price FROM products GROUP BY category"
  }'
```

## Frontend Integration Examples

### JavaScript/Fetch
```javascript
// Execute a command
async function executeCommand(sqlCommand) {
  const response = await fetch('/command', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ query: sqlCommand })
  });
  return await response.json();
}

// Execute a select query
async function executeSelect(sqlQuery) {
  const response = await fetch('/select', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ query: sqlQuery })
  });
  return await response.json();
}

// Usage examples
const insertResult = await executeCommand(
  "INSERT INTO products (name, price) VALUES ('New Product', 99.99)"
);

const selectResult = await executeSelect(
  "SELECT * FROM products WHERE price > 50"
);
```

### React Example
```jsx
import React, { useState } from 'react';

function SQLExecutor() {
  const [query, setQuery] = useState('');
  const [result, setResult] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  const executeQuery = async () => {
    setIsLoading(true);
    try {
      const endpoint = query.trim().toUpperCase().startsWith('SELECT') 
        ? '/select' 
        : '/command';
      
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query })
      });
      
      const data = await response.json();
      setResult(data);
    } catch (error) {
      setResult({ success: false, error: error.message });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div>
      <textarea 
        value={query} 
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Enter your SQL query..."
      />
      <button onClick={executeQuery} disabled={isLoading}>
        {isLoading ? 'Executing...' : 'Execute Query'}
      </button>
      {result && (
        <pre>{JSON.stringify(result, null, 2)}</pre>
      )}
    </div>
  );
}
```

## Security Considerations

### Production Recommendations
1. **Authentication**: Add JWT or API key authentication
2. **Query Validation**: Implement query whitelisting or sanitization
3. **Rate Limiting**: Add rate limiting to prevent abuse
4. **CORS**: Configure CORS properly for your frontend domain
5. **Input Validation**: Validate query length and complexity
6. **Logging**: Monitor and log all SQL executions for audit trails

### Example Security Enhancements
```go
// Add to your handler
func validateQuery(query string) error {
    // Implement your validation logic
    if len(query) > 10000 {
        return errors.New("query too long")
    }
    
    // Block dangerous commands in production
    dangerous := []string{"DROP", "TRUNCATE", "DELETE FROM"}
    for _, cmd := range dangerous {
        if strings.Contains(strings.ToUpper(query), cmd) {
            return errors.New("dangerous command not allowed")
        }
    }
    
    return nil
}
```

## Error Handling

### Common Error Types
1. **Syntax Errors**: Invalid SQL syntax
2. **Permission Errors**: Insufficient database permissions
3. **Connection Errors**: Database connection issues
4. **Timeout Errors**: Query execution timeout
5. **Resource Errors**: Memory or disk space issues

### Error Response Structure
All endpoints return consistent error responses:
```json
{
  "success": false,
  "error": "Detailed error message",
  "query": "The SQL query that failed",
  "duration": 123.45
}
```

## Testing

### Health Check
First, verify the API is running:
```bash
curl http://localhost:8008/health
```

### Basic Functionality Test
```bash
# Test command endpoint
curl -X POST http://localhost:8008/command \
  -H "Content-Type: application/json" \
  -d '{"query": "SHOW TABLES"}'

# Test select endpoint  
curl -X POST http://localhost:8008/select \
  -H "Content-Type: application/json" \
  -d '{"query": "SELECT 1 as test"}'
```

## Performance Tips

1. **Use LIMIT**: Always use LIMIT for large result sets
2. **Optimize Queries**: Use proper indexes and WHERE clauses
3. **Batch Operations**: Combine multiple operations when possible
4. **Connection Pooling**: The API handles connection pooling automatically
5. **Async Processing**: For long-running queries, consider background processing

## Support

For issues or questions:
- Check the API logs for detailed error information
- Verify your ClickHouse connection settings
- Test queries directly in ClickHouse client first
- Ensure proper JSON formatting in requests
