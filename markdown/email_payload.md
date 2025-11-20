# Email Report API Payload

This document describes the JSON payload structure for the `/v1/sendreportemail` endpoint, which triggers generating a report PDF and sending it via email.

The system fetches the report configuration (queries, layout, recipients) directly from MongoDB Atlas based on the provided `schedule_id`. You do not need to send the full report configuration in the request body.

## Endpoint

`POST /v1/sendreportemail`

## Request Body

The request body should be a JSON object with the following fields:

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `shopid` | string | **Yes** | The Shop ID to identify the database connection and context. |
| `schedule_id` | string | **Yes** | The unique ID of the report schedule stored in MongoDB Atlas (collection `email_schedules`). |

### Example: Triggering a Scheduled Report

Use this payload to send the report to the recipients configured in the schedule:

```json
{
  "shopid": "shop123",
  "schedule_id": "sched_monthly_sales_001"
}
```

## Behavior

1.  The API validates `shopid` and `schedule_id`.
2.  It fetches the schedule configuration from MongoDB Atlas (`email_schedules` collection) matching the given `shopid` and `schedule_id`.
3.  It executes the queries defined in the fetched configuration against the PostgreSQL database.
4.  It generates a PDF report based on the layout configuration from MongoDB.
5.  **Email Sending:**
    *   The email is sent to the `recipients` and `cc_recipients` defined in the MongoDB schedule document.
