# core-ksp-api

Backend API berbasis Go, Gin, GORM, dan SQL Server. Repository saat ini menyediakan route default, helper database/response, middleware autentikasi, cache Redis, dan logging HTTP background.

## Dokumentasi

- [Menjalankan aplikasi dan Docker](docs/menjalankan-aplikasi.md)
- [Membuat controller baru](docs/membuat-controller-baru.md)
- [Get_DB_ReqBody dan database](docs/get-db-reqbody-database.md)
- [Autentikasi dan response helper](docs/internal-auth-response-helper.md)
- [Logging HTTP](docs/logging-http.md)
- [Cache Redis](docs/cache-redis.md)

## Menjalankan dengan Docker

Siapkan `.env` menggunakan [panduan konfigurasi](docs/menjalankan-aplikasi.md), lalu jalankan:

```sh
docker compose up --build -d
docker compose logs -f api
```

API tersedia di `http://localhost:5017`. SQL Server berada di luar Compose; Redis disediakan sebagai service `redis`.

## Route aktif

Route yang didaftarkan oleh `routes/route.go`:

| Method | Endpoint | Akses | Perilaku |
| --- | --- | --- | --- |
| GET | `/` | Public | Pesan sambutan API |
| GET | `/db-health` | Public | Memeriksa keberadaan pool database di aplikasi |

`/db-health` belum melakukan ping/query database baru, sehingga response sukses tidak memastikan SQL Server masih dapat diakses saat itu. `db.Main()` panic jika pool belum diinisialisasi; startup normal menjalankan `db.InitDB()` terlebih dahulu.

Path yang tidak ditemukan menghasilkan 404. Handler `NoMethod` sudah terdaftar, tetapi `HandleMethodNotAllowed` belum diaktifkan; method yang tidak cocok saat ini tetap mengikuti 404.

## Struktur kode

| Lokasi | Isi |
| --- | --- |
| `main.go` | Startup, middleware global, HTTP server, shutdown |
| `routes/route.go` | Pendaftaran endpoint aktif |
| `database/database.go` | Pool SQL Server utama |
| `helper/helper.go` | Context bisnis, response, SQL error masking, utilitas |
| `helper/http_logger.go` | Antrean log request/response ke `LOG_HTTP` |
| `auth/auth_internal.go` | JWT, timestamp, request ID, signature |
| `cache/redis.go` | Helper Redis berbasis JSON |
| `obj/` | Struct database, request, response, dan context |

Package `controller` belum tersedia. Panduan controller menjelaskan pola untuk penambahan berikutnya. Middleware autentikasi tersedia, tetapi belum dipasang pada dua route aktif.

## Verifikasi

```sh
go test ./...
go vet ./...
```

Tes logger mencakup body besar yang tetap utuh untuk handler, pemotongan cuplikan log, 404, panic, timeout baca, antrean penuh, dan worker log yang tertahan. Tes tersebut tidak membutuhkan koneksi SQL Server langsung.
