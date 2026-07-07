# AWS S3 Integration

REST API service untuk integrasi dengan **Amazon S3** (Simple Storage Service) — upload, download, dan list file dari S3 bucket. Dibangun dengan **Go**, framework **Gin Gonic**, **GORM** + **PostgreSQL**, dan **AWS SDK Go v1**.

---

## 🧱 Tech Stack

| Komponen      | Library / Teknologi                                |
|---------------|-----------------------------------------------------|
| HTTP Framework | [Gin Gonic](https://github.com/gin-gonic/gin)      |
| ORM           | [GORM](https://gorm.io) + PostgreSQL driver         |
| Database      | PostgreSQL 16                                       |
| Config        | [Viper](https://github.com/spf13/viper)             |
| Validation    | [go-playground/validator v10](https://github.com/go-playground/validator) |
| Logging       | [Logrus](https://github.com/sirupsen/logrus)       |
| AWS SDK       | [aws-sdk-go v1](https://github.com/aws/aws-sdk-go)  |
| Container     | Docker + Docker Compose                             |

---

## 📋 Prerequisites

- **Go** 1.23+
- **PostgreSQL** 16 (atau Docker untuk menjalankannya)
- **AWS Account** dengan S3 bucket dan IAM credentials (Access Key + Secret Key)

---

## 🔧 Environment Variables

Salin `.env.example` ke `.env` dan isi nilainya:

```bash
cp .env.example .env
```

| Variable                | Deskripsi                                   | Default            |
|-------------------------|---------------------------------------------|--------------------|
| `APP_STATUS`            | Environment status (`development`/`production`) | `development`  |
| `APP_PORT`              | Port aplikasi                               | `4001`             |
| `AUTO_MIGRATION_SWITCH` | Auto-migrate database (`1` = on, `0` = off) | `1`                |
| `DB_USERNAME`           | Username PostgreSQL                         | `postgres`         |
| `DB_PASSWORD`           | Password PostgreSQL                         | `root`             |
| `DB_NAME`               | Nama database                               | `aws`              |
| `DB_HOST`               | Host database                               | `localhost`        |
| `DB_PORT`               | Port database                               | `5432`             |
| `DB_SSLMODE`            | SSL mode PostgreSQL                         | `disable`          |
| `DB_CONN_MAX_IDLE_TIME` | Max idle time koneksi DB                    | `30m`              |
| `DB_CONN_MAX_LIFE_TIME` | Max lifetime koneksi DB                     | `30m`              |
| `DB_MAX_OPEN_CONN`      | Max open connections                        | `50`               |
| `DB_MAX_IDLE_CONN`      | Max idle connections                        | `10`               |
| `LOGGER_STDOUT`         | Log ke stdout                               | `true`             |
| `LOGGER_FILE_LOCATION`  | Path file log                               | `app.log`          |
| `LOGGER_LEVEL`          | Level log (`info`, `debug`, `warn`, `error`) | `info`            |
| `AWS_ACCESS_KEY`        | AWS IAM Access Key ID                       | —                  |
| `AWS_SECRET_KEY`        | AWS IAM Secret Access Key                   | —                  |
| `AWS_DEFAULT_REGION`    | AWS Region S3 bucket                        | `ap-southeast-1`   |
| `AWS_BUCKET`            | Nama S3 bucket                              | —                  |
| `AWS_URL_API`           | Base URL S3 bucket                          | —                  |
| `AWS_ACL`               | ACL default untuk upload                    | `public-read`      |

---

## 🐳 Menjalankan dengan Docker Compose

Cara paling cepat untuk menjalankan seluruh stack (aplikasi + PostgreSQL):

```bash
# 1. Clone repository
git clone <repo-url> && cd go-aws-s3-bucket

# 2. Salin & isi environment variables
cp .env.example .env
# Edit .env — isi AWS_ACCESS_KEY, AWS_SECRET_KEY, AWS_BUCKET, AWS_URL_API, dll.

# 3. Jalankan dengan Docker Compose
docker compose up -d

# 4. Cek status container
docker compose ps

# 5. Lihat log aplikasi
docker compose logs -f app

# 6. Hentikan container
docker compose down
```

Aplikasi akan berjalan di `http://localhost:4001`.

---

## 🚀 Menjalankan Secara Lokal (Tanpa Docker)

```bash
# 1. Clone repository
git clone <repo-url> && cd go-aws-s3-bucket

# 2. Setup environment
cp .env.example .env
# Edit .env sesuai konfigurasi lokal

# 3. Install dependencies
go mod tidy
go mod download

# 4. Jalankan aplikasi
go run main.go
```

Pastikan PostgreSQL sudah berjalan sebelum menjalankan aplikasi.

---

## 📡 API Endpoints

Base URL: `http://localhost:4001/api/v1`

Semua response mengikuti format standar:

```json
{
  "api_id": "API_CALL_1750000000_1234567",
  "status": "SUCCESS",
  "message": "Success get data",
  "data": { ... },
  "meta": null
}
```

---

### 1. List Semua File dari S3 Bucket

Mengambil daftar seluruh objek/file yang ada di dalam S3 bucket.

```
GET /api/v1/list
```

**Contoh Request:**

```bash
curl --location 'http://localhost:4001/api/v1/list'
```

**Contoh Response (200 OK):**

```json
{
  "api_id": "API_CALL_1750000000_1234567",
  "status": "SUCCESS",
  "message": "Success get data",
  "data": [
    {
      "url": "https://my-bucket.s3.ap-southeast-1.amazonaws.com/images/photo.jpg",
      "name": "images/photo.jpg",
      "size": "2.5 MB",
      "last_modified": "2025-06-15T10:30:00+07:00"
    },
    {
      "url": "https://my-bucket.s3.ap-southeast-1.amazonaws.com/docs/report.pdf",
      "name": "docs/report.pdf",
      "size": "150.3 KB",
      "last_modified": "2025-06-14T08:15:00+07:00"
    }
  ],
  "meta": null
}
```

---

### 2. Upload File ke S3 Bucket

Mengunggah file ke folder tertentu di dalam S3 bucket. File akan otomatis dibuatkan nama unik (timestamp + random number) dan di-upload dengan ACL `public-read`.

```
POST /api/v1/upload
```

**Body:** `multipart/form-data`

| Field    | Tipe  | Wajib | Deskripsi                       |
|----------|-------|-------|----------------------------------|
| `folder` | text  | ✅    | Nama folder tujuan di S3 bucket |
| `file`   | file  | ✅    | File yang akan diupload         |

**Allowed MIME Types:** Semua file kecuali yang diblokir (executable, shell script, dll.)

**Contoh Request (cURL):**

```bash
curl --location 'http://localhost:4001/api/v1/upload' \
  --form 'folder="images"' \
  --form 'file=@"/path/to/your/photo.jpg"'
```

**Contoh Response (200 OK):**

```json
{
  "api_id": "API_CALL_1750000000_7654321",
  "status": "SUCCESS",
  "message": "Success save data",
  "data": {
    "path": "images/1750000000_7654321_photo.jpg"
  },
  "meta": null
}
```

**Contoh Response Gagal (400 Bad Request):**

```json
{
  "api_id": "API_CALL_1750000000_7654321",
  "status": "FAILED",
  "message": "Invalid payload data",
  "data": [
    {
      "field": "folder",
      "message": "folder is required"
    }
  ],
  "meta": null
}
```

---

### 3. Download File dari S3 Bucket

Mendownload file dari S3 bucket berdasarkan path/object key.

```
GET /api/v1/download?path={object_key}
```

**Query Parameter:**

| Parameter | Tipe   | Wajib | Deskripsi                          |
|-----------|--------|-------|-------------------------------------|
| `path`    | string | ✅    | Object key (path) file di S3 bucket |

**Contoh Request (cURL):**

```bash
curl --location 'http://localhost:4001/api/v1/download?path=images%2F1750000000_7654321_photo.jpg'
```

> **Catatan:** Gunakan URL-encoded path. Misalnya `/` menjadi `%2F`.

```bash
# Atau gunakan --output untuk menyimpan file
curl --location 'http://localhost:4001/api/v1/download?path=images%2Fphoto.jpg' \
  --output downloaded-photo.jpg
```

**Contoh Response (200 OK):**
- Binary file stream dengan header `Content-Disposition: attachment; filename="photo.jpg"`

**Contoh Response Gagal (400 Bad Request):**

```json
{
  "api_id": "API_CALL_1750000000_9999999",
  "status": "FAILED",
  "message": "Invalid payload data",
  "data": [
    {
      "field": "path",
      "message": "path is required"
    }
  ],
  "meta": null
}
```

---

### 4. Generate Presigned URL

Membuat **presigned URL** sementara untuk mengakses file di S3 bucket tanpa memerlukan AWS credentials. URL yang dihasilkan memiliki masa berlaku terbatas.

```
GET /api/v1/presigned-url?path={object_key}&expires={minutes}
```

**Query Parameter:**

| Parameter | Tipe   | Wajib | Default | Deskripsi                                      |
|-----------|--------|-------|---------|-------------------------------------------------|
| `path`    | string | ✅    | —       | Object key (path) file di S3 bucket             |
| `expires` | int    | ❌    | `15`    | Masa berlaku URL dalam menit (min: 1, max: 10080 / 7 hari) |

**Contoh Request (cURL):**

```bash
# Dengan expires default (15 menit)
curl --location 'http://localhost:4001/api/v1/presigned-url?path=images%2Fphoto.jpg'

# Dengan expires custom (60 menit)
curl --location 'http://localhost:4001/api/v1/presigned-url?path=images%2Fphoto.jpg&expires=60'
```

**Contoh Response (200 OK):**

```json
{
  "api_id": "API_CALL_1750000000_5555555",
  "status": "SUCCESS",
  "message": "Success get data",
  "data": {
    "url": "https://my-bucket.s3.ap-southeast-1.amazonaws.com/images/photo.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...&X-Amz-Signature=...",
    "path": "images/photo.jpg",
    "expires_at": "2025-06-15T11:00:00Z"
  },
  "meta": null
}
```

**Contoh Response Gagal (400 Bad Request):**

```json
{
  "api_id": "API_CALL_1750000000_5555555",
  "status": "FAILED",
  "message": "Invalid payload data",
  "data": [
    {
      "field": "path",
      "message": "path is required"
    }
  ],
  "meta": null
}
```

> **Use case:** Gunakan endpoint ini untuk memberikan akses sementara ke file private di S3 tanpa harus membuatnya public. URL akan otomatis kadaluarsa setelah waktu yang ditentukan.

---

### 5. Move File (Pindahkan dari `tmp/` ke Folder Final)

Memindahkan file dari satu path ke path lain di dalam S3 bucket. Operasi ini melakukan **copy lalu delete** pada source (S3 tidak memiliki operasi "move" native).

```
POST /api/v1/move
```

**Body:** `application/json`

| Field        | Tipe   | Wajib | Deskripsi                              |
|--------------|--------|-------|----------------------------------------|
| `source_path` | string | ✅    | Path sumber file (contoh: `tmp/photo.jpg`) |
| `dest_path`   | string | ✅    | Path tujuan file (contoh: `images/photo.jpg`) |

**Contoh Request (cURL):**

```bash
curl --location 'http://localhost:4001/api/v1/move' \
  --header 'Content-Type: application/json' \
  --data '{
    "source_path": "tmp/photo.jpg",
    "dest_path": "images/photo.jpg"
  }'
```

**Contoh Response (200 OK):**

```json
{
  "api_id": "API_CALL_1750000000_8888888",
  "status": "SUCCESS",
  "message": "Success save data",
  "data": {
    "source_path": "tmp/photo.jpg",
    "dest_path": "images/photo.jpg"
  },
  "meta": null
}
```

**Contoh Response Gagal (400 Bad Request):**

```json
{
  "api_id": "API_CALL_1750000000_8888888",
  "status": "FAILED",
  "message": "Invalid payload data",
  "data": [
    {
      "field": "source_path",
      "message": "source_path is required"
    }
  ],
  "meta": null
}
```

---

## 🔄 Two-Step Upload Flow (Staging via `tmp/`)

Pola ini berguna ketika kamu ingin mengupload file dulu ke folder staging (`tmp/`), memverifikasinya (misalnya lewat presigned URL), lalu mengkonfirmasi dengan memindahkan ke folder permanen.

```
┌──────────┐      ┌──────────────┐      ┌─────────────────┐
│  Upload   │────▶ │  Presigned   │────▶ │     Move         │
│  to tmp/  │      │  URL (cek)   │      │  tmp/ → final    │
└──────────┘      └──────────────┘      └─────────────────┘
```

**Step-by-step:**

```bash
# 1. Upload file ke folder staging "tmp/"
curl --location 'http://localhost:4001/api/v1/upload' \
  --form 'folder="tmp"' \
  --form 'file=@"/path/to/photo.jpg"'

# Response: { "data": { "path": "tmp/1750000000_1234567_photo.jpg" } }

# 2. Generate presigned URL untuk verifikasi file di tmp/
curl --location 'http://localhost:4001/api/v1/presigned-url?path=tmp%2F1750000000_1234567_photo.jpg&expires=15'

# Response: { "data": { "url": "https://...presigned...", "expires_at": "..." } }
# Buka URL tersebut di browser untuk verifikasi file.

# 3. Jika file sudah benar, pindahkan ke folder final
curl --location 'http://localhost:4001/api/v1/move' \
  --header 'Content-Type: application/json' \
  --data '{
    "source_path": "tmp/1750000000_1234567_photo.jpg",
    "dest_path": "images/gallery/photo.jpg"
  }'
```

> **Mengapa tidak otomatis detect dari presigned URL?** Presigned URL mengarah langsung ke S3 (tidak melalui server kita), sehingga server tidak bisa tahu kapan URL itu diakses. Oleh sebab itu, flow yang benar adalah **client yang mengkonfirmasi** via endpoint `/move` setelah verifikasi selesai.

---

## 📁 Struktur Proyek

```
go-aws-s3-bucket/
├── .env                        # Environment config (tidak di-commit)
├── .env.example                # Template environment config
├── .gitignore
├── Dockerfile                  # Multi-stage Docker build
├── docker-compose.yml          # Docker Compose orchestration
├── go.mod
├── go.sum
├── LICENSE
├── main.go                     # Entry point
├── README.md
├── logs/
│   └── app.log
└── app/
    ├── config/                 # App config, DB, server, validator, viper
    ├── constant/               # HTTP response constants
    ├── handler/                # HTTP request handlers
    ├── helper/                 # Utilities (logging, response, validation, file)
    ├── integration/            # AWS S3 integration layer
    ├── middleware/              # Custom middleware (RequestID)
    ├── model/                  # Database models (GORM)
    ├── resource/
    │   ├── request/            # Request DTO + validation rules
    │   └── response/           # Response DTO + formatters
    ├── router/                 # Route definitions
    └── service/                # Business logic layer
```

---

## 📄 Lisensi

MIT © 2025 Choirul Anwar — lihat file [LICENSE](LICENSE).
