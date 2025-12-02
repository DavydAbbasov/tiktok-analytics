# TikTok Analytics Service

Backend service that tracks TikTok video statistics and calculates earnings based on views.
The service exposes a REST API, stores video metadata and hourly stats in PostgreSQL,
and fetches live metrics from the EnsembleData TikTok API.


## Features

* Track TikTok videos by URL or ID
* Hourly statistics with history
* Earnings calculation (configurable formula)
* Error logs stored per video
* Tracking statuses: `active`, `error`, `stopped`
* Clean Architecture + transactions for critical operations
* Provider retry logic
* Full Swagger documentation


## Tech stack

* **Language:** Go 1.22+
* **HTTP server:** net/http + chi router
* **Database:** PostgreSQL
* **External provider:** EnsembleData TikTok API


## External TikTok provider

The service does **not** scrape TikTok directly.
All statistics are fetched from the **EnsembleData TikTok API**.

* Provider: [EnsembleData](https://ensembledata.com)
* Base URL: `https://ensembledata.com/apis`
* Endpoint: `GET /tt/post/info`
* Auth: API token passed as `token`

Example:

```bash
curl "https://ensembledata.com/apis/tt/post/info \
  ?url=https%3A%2F%2Fwww.tiktok.com%2F@user%2Fvideo%2F1234567890 \
  &token=YOUR_TOKEN \
  &new_version=false \
  &download_video=false"
```

## How to run locally

### 1. Clone the repository

```bash
git clone https://github.com/DavydAbbasov/tiktok-analytics.git
```

### 2. Start the service

```bash
make run
```

Swagger UI:

```
http://localhost:8080/swagger/index.html
```

## Environment variables

All configuration lives in `.env`.
Template is provided in `.env.example`.


