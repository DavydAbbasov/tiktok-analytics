# TikTok Analytics Service

Backend service that tracks TikTok video statistics and estimates earnings based on views.
The service exposes a REST API, stores video metadata and stats in PostgreSQL, and fetches
live metrics from the EnsembleData TikTok API.

## Tech stack
- **Language:** Go 1.22+
- **HTTP server:** net/http + chi router
- **Database:** PostgreSQL
- **External provider:** EnsembleData TikTok API

## External TikTok provider
The service does **not** scrape TikTok directly.
All video statistics are fetched via the **EnsembleData TikTok API**.

- Provider: [EnsembleData](https://ensembledata.com)
- Base URL: `https://ensembledata.com/apis`
- Endpoint used: `GET /tt/post/info`
- Auth: API token passed as `token` query parameter
- Required params:
  - `url` – full TikTok video URL
  - `token` – your EnsembleData API token

Example raw request:
```bash
curl "https://ensembledata.com/apis/tt/post/info \
  ?url=https%3A%2F%2Fwww.tiktok.com%2F@user%2Fvideo%2F1234567890 \
  &token=YOUR_TOKEN \
  &new_version=false \
  &download_video=false"
```
## How to run locally
1. **Clone the repository**

```bash
1) git clone https://github.com/DavydAbbasov/tiktok-analytics.git

```
