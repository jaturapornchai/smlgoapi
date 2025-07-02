# 📊 Database Endpoints API Documentation

## Overview

Database endpoints for executing SQL queries and commands on ClickHouse and PostgreSQL databases.

## Base URL

`http://localhost:8008/v1`

---

## 📋 Available Endpoints

### 1. GET `/tables`

Get all available database tables.

#### Usage Examples

```bash
curl "http://localhost:8008/v1/tables"
```

```javascript
const getTables = async () => {
  const response = await fetch("http://localhost:8008/v1/tables");
  return await response.json();
};
```

#### Response Format

```json
{
  "success": true,
  "data": [
    {
      "name": "ic_inventory",
      "engine": "PostgreSQL",
      "rows": 50000,
      "columns": 15
    }
  ]
}
```

---

### 2. POST `/command`

Execute SQL commands (INSERT, UPDATE, DELETE, CREATE) on ClickHouse.

#### Request Format

```json
{
  "command": "INSERT INTO table_name VALUES (...)"
}
```

#### Usage Examples

```bash
curl -X POST "http://localhost:8008/v1/command" \
  -H "Content-Type: application/json" \
  -d '{
    "command": "CREATE TABLE test (id Int32, name String) ENGINE = Memory"
  }'
```

```javascript
const executeCommand = async (command) => {
  const response = await fetch("http://localhost:8008/v1/command", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ command }),
  });
  return await response.json();
};
```

---

### 3. POST `/select`

Execute SELECT queries on ClickHouse.

#### Request Format

```json
{
  "query": "SELECT * FROM table_name WHERE condition"
}
```

#### Usage Examples

```bash
curl -X POST "http://localhost:8008/v1/select" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT * FROM ic_inventory LIMIT 10"
  }'
```

```python
def execute_select(query):
    url = "http://localhost:8008/v1/select"
    payload = {"query": query}
    response = requests.post(url, json=payload)
    return response.json()

# Usage
results = execute_select("SELECT COUNT(*) FROM ic_inventory")
```

---

### 4. POST `/pgcommand`

Execute SQL commands on PostgreSQL.

#### Request Format

```json
{
  "command": "CREATE INDEX idx_name ON table_name (column)"
}
```

#### Usage Examples

```bash
curl -X POST "http://localhost:8008/v1/pgcommand" \
  -H "Content-Type: application/json" \
  -d '{
    "command": "CREATE INDEX idx_code ON ic_inventory (code)"
  }'
```

---

### 5. POST `/pgselect`

Execute SELECT queries on PostgreSQL.

#### Request Format

```json
{
  "query": "SELECT * FROM table_name WHERE condition"
}
```

#### Usage Examples

```bash
curl -X POST "http://localhost:8008/v1/pgselect" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "SELECT * FROM ic_inventory WHERE code LIKE '\''%AC%'\''"
  }'
```

```javascript
const pgSelect = async (query) => {
  const response = await fetch("http://localhost:8008/v1/pgselect", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query }),
  });
  return await response.json();
};

// Usage
const products = await pgSelect(
  "SELECT * FROM ic_inventory WHERE name LIKE '%toyota%'"
);
```

---

## 📊 Response Formats

### Success Response

```json
{
  "success": true,
  "data": [
    {
      "column1": "value1",
      "column2": "value2"
    }
  ],
  "message": "Query executed successfully",
  "rows_affected": 1,
  "execution_time": 125.5
}
```

### Error Response

```json
{
  "success": false,
  "error": "SQL syntax error",
  "message": "Invalid query format"
}
```

---

## 🔧 Use Cases

### Database Administration

```javascript
// Create new table
await executeCommand(`
  CREATE TABLE products_backup AS 
  SELECT * FROM ic_inventory WHERE created_at > '2024-01-01'
`);

// Create index for better performance
await executePgCommand(`
  CREATE INDEX CONCURRENTLY idx_inventory_name 
  ON ic_inventory USING gin(to_tsvector('english', name))
`);
```

### Data Analysis

```python
# Get sales statistics
sales_query = """
SELECT
  DATE_TRUNC('month', created_at) as month,
  COUNT(*) as total_products,
  AVG(price) as avg_price
FROM ic_inventory
WHERE price > 0
GROUP BY month
ORDER BY month DESC
LIMIT 12
"""

results = execute_pg_select(sales_query)
```

### Data Migration

```bash
# Export data from ClickHouse
curl -X POST "http://localhost:8008/v1/select" \
  -H "Content-Type: application/json" \
  -d '{"query": "SELECT * FROM clickhouse_table"}' > export.json

# Import to PostgreSQL (after processing)
curl -X POST "http://localhost:8008/v1/pgcommand" \
  -H "Content-Type: application/json" \
  -d '{"command": "INSERT INTO pg_table VALUES (...)"}'
```

---

## 🚨 Security Considerations

### SQL Injection Prevention

- Always use parameterized queries when possible
- Validate input data
- Avoid dynamic SQL construction

### Access Control

- Commands require proper database permissions
- Some operations may be restricted
- Monitor query execution logs

---

## 📈 Performance Tips

### Query Optimization

- Use appropriate indexes
- Limit result sets with LIMIT clause
- Use EXPLAIN to analyze query plans
- Avoid SELECT \* for large tables

### Best Practices

```sql
-- Good: Specific columns and conditions
SELECT code, name, price
FROM ic_inventory
WHERE code = 'AC3006'

-- Better: With index hints
SELECT code, name, price
FROM ic_inventory
WHERE code = 'AC3006'
AND created_at > '2024-01-01'
LIMIT 100

-- Best: With proper indexing
CREATE INDEX idx_code_date ON ic_inventory (code, created_at);
```

---

## 🔍 Example Queries

### Inventory Management

```sql
-- Check low stock items
SELECT code, name, qty_available
FROM ic_inventory
WHERE qty_available < 10
AND qty_available > 0
ORDER BY qty_available ASC;

-- Find products by supplier
SELECT code, name, supplier_code, price
FROM ic_inventory
WHERE supplier_code = 'AC'
ORDER BY price DESC;
```

### Price Analysis

```sql
-- Average price by category
SELECT
  LEFT(code, 2) as category,
  COUNT(*) as product_count,
  AVG(price) as avg_price,
  MIN(price) as min_price,
  MAX(price) as max_price
FROM ic_inventory
WHERE price > 0
GROUP BY LEFT(code, 2)
ORDER BY avg_price DESC;
```

### Search Optimization

```sql
-- Create full-text search index
CREATE INDEX idx_inventory_search
ON ic_inventory
USING gin(to_tsvector('english', name || ' ' || code));

-- Use full-text search
SELECT code, name,
       ts_rank(to_tsvector('english', name || ' ' || code),
               plainto_tsquery('english', 'toyota brake')) as rank
FROM ic_inventory
WHERE to_tsvector('english', name || ' ' || code) @@
      plainto_tsquery('english', 'toyota brake')
ORDER BY rank DESC;
```
