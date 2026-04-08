# Website Crawler API

Aplikasi/API untuk crawling website dan menyimpan hasilnya dalam file HTML. Mendukung website bertipe **SPA**, **SSR**, dan **PWA**.

## Tech Stack

- **Bahasa**: Go 1.24
- **Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Crawler Engine**: [chromedp](https://github.com/chromedp/chromedp) (Headless Chrome) + HTTP fallback
- **Storage**: File system

## Struktur Project

```
crawler-website/
├── main.go                         # Entry point
├── config/
│   └── config.go                   # Konfigurasi aplikasi
├── model/
│   └── crawl_result.go             # Entity hasil crawl
├── request/
│   └── crawl_request.go            # Request DTO
├── response/
│   └── crawl_response.go           # Response DTO
├── repository/
│   └── crawl_repository.go         # Penyimpanan file HTML
├── service/
│   └── crawl_service.go            # Business logic crawling
├── controller/
│   └── crawl_controller.go         # HTTP handler
├── api/
│   └── router.go                   # Route setup
├── helper/
│   └── error.go                    # Utility functions
└── crawled_pages/                  # Output hasil crawl
    ├── cmlabs_co.html
    ├── sequence_day.html
    └── tokopedia_com.html
```

## Cara Menjalankan

### Prasyarat

- Go 1.24+
- Google Chrome / Chromium (opsional, untuk full SPA rendering)

### Instalasi

```bash
git clone https://github.com/ilhaamms/cmlabs-backend-crawler-freelance-test.git
cd cmlabs-backend-crawler-freelance-test
go mod tidy
```

### Jalankan Server

```bash
go run main.go
```

Saat dijalankan, aplikasi akan otomatis meng-crawl 3 website berikut lalu menyalakan HTTP server di port `8080`:

| # | Website | Tipe |
|---|---------|------|
| 1 | https://cmlabs.co | SSR |
| 2 | https://sequence.day | SPA |
| 3 | https://tokopedia.com | PWA |

## API Endpoints

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| `GET` | `/health` | Health check |
| `POST` | `/api/crawl` | Crawl satu URL |
| `POST` | `/api/crawl/batch` | Crawl banyak URL sekaligus |
| `GET` | `/api/crawl/files` | List semua file hasil crawl |
| `GET` | `/api/crawl/files/:filename` | Download file HTML |

### Contoh Request

**Crawl satu URL:**

```bash
curl -X POST http://localhost:8080/api/crawl \
  -H "Content-Type: application/json" \
  -d '{"url": "https://cmlabs.co", "wait_selector": "body", "timeout": 30}'
```

**Response:**

```json
{
  "status": "success",
  "message": "Website berhasil di-crawl",
  "data": {
    "url": "https://cmlabs.co",
    "title": "Professional and Reliable SEO Agency Indonesia - cmlabs",
    "file_path": "crawled_pages/cmlabs_co.html",
    "crawled_at": "2026-04-08T12:02:29+07:00",
    "status_code": 200,
    "content_length": 601839,
    "crawl_method": "http"
  }
}
```

**Crawl batch (banyak URL):**

```bash
curl -X POST http://localhost:8080/api/crawl/batch \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      {"url": "https://cmlabs.co"},
      {"url": "https://sequence.day"}
    ],
    "timeout": 30
  }'
```

**List file hasil crawl:**

```bash
curl http://localhost:8080/api/crawl/files
```

**Download file HTML:**

```bash
curl http://localhost:8080/api/crawl/files/cmlabs_co.html -o cmlabs_co.html
```

## Strategi Crawling

Aplikasi menggunakan dua metode crawling:

1. **chromedp (Headless Chrome)** — Digunakan sebagai metode utama. Dapat merender JavaScript sehingga mendukung website SPA, SSR, dan PWA secara penuh.
2. **HTTP Client** — Digunakan sebagai fallback jika Chrome tidak tersedia. Cocok untuk website SSR yang tidak membutuhkan JavaScript rendering.

```
Request → chromedp (Headless Chrome)
              ↓ (jika gagal/tidak tersedia)
         HTTP Client (fallback)
              ↓
         Simpan ke file HTML
```

## Environment Variables

| Variable | Default | Deskripsi |
|----------|---------|-----------|
| `PORT` | `8080` | Port HTTP server |
| `CRAWLED_PAGES_DIR` | `crawled_pages` | Direktori penyimpanan file HTML |
| `CHROME_TIMEOUT` | `60` | Timeout crawling dalam detik |

## Task

1. ✅ Buat Aplikasi / API untuk crawling website dan menyimpan hasilnya dalam file HTML dengan ketentuan crawler harus bisa meng-crawl website tipe SPA, SSR, ataupun PWA.
2. ✅ Crawl website berikut dan simpan hasilnya di file HTML:
   - https://cmlabs.co
   - https://sequence.day
   - https://tokopedia.com (website bebas)
