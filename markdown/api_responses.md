# API Payload & Response Documentation

This document describes the request payloads and response structures for the SMLGOAPI.

## Common Response Structures

### Standard Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Standard Error Response
```json
{
  "success": false,
  "error": "Error message description"
}
```

---

## Health Check

### `GET /health` or `GET /v1/health`
Checks the API server status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-10-27T10:00:00Z",
  "version": "1.0.0",
  "database": "connected"
}
```

---

## PostgreSQL Endpoints

### `POST /v1/pgcommand`
Execute a SQL command (INSERT, UPDATE, DELETE, etc.).

**Request:**
```json
{
  "database_name": "my_database",
  "query": "INSERT INTO users (name) VALUES ('John')"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Command executed successfully",
  "result": { ... },
  "command": "INSERT INTO users (name) VALUES ('John')",
  "database": "my_database",
  "duration_ms": 12.5
}
```

### `POST /v1/pgselect`
Execute a SELECT query.

**Request:**
```json
{
  "database_name": "my_database",
  "query": "SELECT * FROM users LIMIT 10"
}
```

**Response:**
```json
{
  "success": true,
  "data": [
    { "id": 1, "name": "John" },
    { "id": 2, "name": "Jane" }
  ],
  "database": "my_database",
  "row_count": 2,
  "columns": ["id", "name"],
  "duration_ms": 5.2
}
```

### `POST /v1/pgselectgroup`
Execute multiple SELECT queries in a single request.

**Request:**
```json
{
  "database_name": "my_database",
  "queries": [
    "SELECT count(*) FROM users",
    "SELECT * FROM products LIMIT 5"
  ]
}
```

**Response:**
```json
{
  "success": true,
  "database": "my_database",
  "total_queries": 2,
  "successful": 2,
  "failed": 0,
  "duration_ms": 20.5,
  "results": [
    {
      "query_index": 0,
      "query": "SELECT count(*) FROM users",
      "success": true,
      "row_count": 1,
      "data": [{ "count": 100 }]
    },
    {
      "query_index": 1,
      "query": "SELECT * FROM products LIMIT 5",
      "success": true,
      "row_count": 5,
      "data": [ ... ]
    }
  ]
}
```

### `POST /v1/pgcheckdatabase`
Check if a database exists, create if not.

**Request:**
```json
{
  "database_name": "new_database"
}
```

**Response:**
```json
{
  "success": true,
  "status": "exists", 
  "database": "new_database",
  "message": "Database exists",
  "duration_ms": 2.1
}
```
*Status can be "exists" or "created".*

### `POST /v1/pgchecktable`
Check if a table exists, create from SQL script if not.

**Request:**
```json
{
  "database_name": "my_database",
  "table_name": "users"
}
```

**Response:**
```json
{
  "success": true,
  "status": "exists",
  "table_name": "users",
  "database": "my_database",
  "message": "Table exists",
  "duration_ms": 3.5
}
```
*Status can be "exists", "created", or "script_not_found".*

---

## ClickHouse Endpoints

### `POST /v1/chcommand`
Execute a command on ClickHouse.

**Request:**
```json
{
  "database_name": "default",
  "query": "CREATE TABLE IF NOT EXISTS logs ..."
}
```

**Response:**
Same structure as `pgcommand`.

### `POST /v1/chselect`
Execute a SELECT query on ClickHouse.

**Request:**
```json
{
  "database_name": "default",
  "query": "SELECT * FROM logs LIMIT 10"
}
```

**Response:**
Same structure as `pgselect`.

### `POST /v1/chselectgroup`
Execute multiple SELECT queries on ClickHouse.

**Request:**
```json
{
  "database_name": "default",
  "queries": ["SELECT 1", "SELECT 2"]
}
```

**Response:**
Same structure as `pgselectgroup`.

---

## Result Handling Endpoints

### `POST /v1/resultfromquery`
Execute queries and store results in a temporary table for reporting/PDF generation.

**Request:**
```json
{
  "shopid": "shop123",
  "guid": "unique-request-id",
  "query_items": [
    {
      "alias": "header",
      "query": "SELECT * FROM doc_header WHERE doc_no = 'INV001'",
      "summary_config": { ... },
      "link_config": { ... }
    },
    {
      "alias": "detail",
      "query": "SELECT * FROM doc_detail WHERE doc_no = 'INV001'"
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Query processed successfully",
  "guid": "unique-request-id"
}
```

### `POST /v1/resultget`
Retrieve stored results (pagination supported).

**Request:**
```json
{
  "shopid": "shop123",
  "guid": "unique-request-id",
  "limit": 100,
  "offset": 0
}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "querynumber": 0,
      "linenumber": 1,
      "level": 0,
      "typejson": 0,
      "datajson": { ... }
    }
  ]
}
```

### `POST /v1/resulttopdf`
Generate a PDF from stored results.

**Request:**
```json
{
  "shopid": "shop123",
  "guid": "unique-request-id",
  "pdf_config": {
    "title": "Invoice Report",
    "orientation": "P",
    "page_size": "A4"
  },
  "layout_config": {
    "sections": [
      {
        "alias": "header",
        "title": "Header Info",
        "columns": [
          { "field": "doc_no", "label": "Document No", "flex": 1 }
        ]
      }
    ],
    "styles": { ... }
  }
}
```

**Response:**
(Returns PDF file stream)

### `POST /v1/sendreportemail`
Generate report and send via email based on schedule configuration.

**Request:**
```json
{
  "shopid": "shop123",
  "schedule_id": "schedule_001"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Report generated and email sent successfully",
  "guid": "generated-guid"
}
```

---

## MongoDB Atlas Endpoints

### `POST /v1/mongoatlasupdate`
Insert or Update (Upsert) a document.

**Request:**
```json
{
  "collection": "my_collection",
  "filter": { "code": "P001" },
  "data": { "name": "Product 1", "price": 100 },
  "upsert": true,
  "replaceone": false
}
```

**Response:**
```json
{
  "status": "success",
  "code": 200,
  "matched_count": 1,
  "modified_count": 1,
  "upserted_id": null
}
```

### `POST /v1/mongoatlasdelete`
Delete document(s).

**Request:**
```json
{
  "collection": "my_collection",
  "filter": { "code": "P001" },
  "delete_many": false
}
```

**Response:**
```json
{
  "status": "success",
  "code": 200,
  "deleted_count": 1
}
```

### `POST /v1/mongoatlasget`
Query documents.

**Request:**
```json
{
  "collection": "my_collection",
  "filter": { "price": { "$gt": 50 } },
  "projection": { "name": 1, "price": 1 },
  "sort": { "price": -1 },
  "limit": 10,
  "skip": 0
}
```

**Response:**
```json
{
  "status": "success",
  "code": 200,
  "count": 5,
  "data": [
    { "_id": "...", "name": "Product 2", "price": 200 },
    { "_id": "...", "name": "Product 1", "price": 100 }
  ]
}
```
