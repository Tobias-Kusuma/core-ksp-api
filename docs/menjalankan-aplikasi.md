# Menjalankan aplikasi

## Konfigurasi environment

Kode membaca environment proses melalui `os.Getenv`. Pemanggilan `godotenv.Load()` masih dinonaktifkan, sehingga `go run .` tidak otomatis membaca `.env`. Docker Compose meneruskan `.env` melalui `env_file`.

Contoh `.env` dengan nilai dummy:

```dotenv
PORT=":8080"
JWTKEY="ganti-dengan-secret-jwt"
SIGNSECRET="ganti-dengan-secret-signature"
MAIN_DBSERVER="host.docker.internal"
MAIN_DBPORT="1433"
MAIN_DBUSER="api_user"
MAIN_DBPASSWORD="ganti-dengan-password-db"
MAIN_DBNAME="CORE_KSP"
```

Kelima `MAIN_DB...` wajib terisi. Prefix koneksi aktif adalah `MAIN`. `DOC_DB...` dan `LOG_DB...` tidak dibaca oleh inisialisasi yang aktif. `PORT` adalah alamat listen Go, misalnya `:8080`; jika kosong, default `:8080`.

`JWTKEY` dan `SIGNSECRET` dibaca saat startup. Setter saat ini tidak menolak nilai kosong. Jangan memakai secret contoh untuk deployment.

Redis masih menggunakan alamat tetap `redis:6379`, password kosong, database 0 di `cache/redis.go`; belum ada konfigurasi environment Redis. `INSTANCE_NAME` diteruskan Compose tetapi belum dibaca kode aplikasi.

## Docker Compose

```sh
docker compose up --build -d
docker compose logs -f api
docker compose down
```

| Service | Konfigurasi aktif |
| --- | --- |
| `api` | Image `core-ksp-api:1.0`, host port 5017 ke container port 8080 |
| `redis` | Image `redis:7-alpine`, host port 6379 ke container port 6379 |

Dockerfile membangun binary Linux dengan `CGO_ENABLED=0`, lalu menjalankannya di image Alpine. `.env` tidak disalin ke image karena dikecualikan di `.dockerignore`. SQL Server tidak dibuat oleh Compose. MinIO yang masih berupa komentar juga tidak dijalankan.

Jika mengubah port listen API di `.env`, sesuaikan sisi port container pada `compose.yaml`. Perubahan environment perlu recreate container; menjalankan ulang `docker compose up -d` menerapkan konfigurasi Compose terbaru.

Redis belum menggunakan volume persisten atau healthcheck. `depends_on` mengatur urutan startup service, tetapi tidak menjamin Redis siap menerima request. Dua endpoint public yang sekarang aktif belum menggunakan Redis.

## SQL Server di luar container

Untuk server remote, isi `MAIN_DBSERVER` dengan IP/hostname server dan `MAIN_DBPORT` dengan port TCP SQL Server. Container harus dapat menjangkau tujuan tersebut; SQL Server harus menerima TCP dan firewall harus mengizinkan koneksi. Mapping `5017:8080` hanya untuk akses masuk API, bukan untuk koneksi keluar ke SQL Server.

Untuk SQL Server di mesin Docker Desktop, gunakan `host.docker.internal`. `localhost` dari dalam container menunjuk ke container itu sendiri. Hostname khusus ini tidak otomatis tersedia pada semua instalasi Docker Engine Linux; gunakan alamat host yang dapat dijangkau atau konfigurasi host gateway sesuai lingkungan deployment.

Connection string saat ini memakai `encrypt=disable`. Kode tidak menjalankan migrasi otomatis. Database dan tabel `LOG_HTTP` perlu tersedia sebelum digunakan; pemetaan kolom dijelaskan di [logging HTTP](logging-http.md).

## Menjalankan langsung dari Go

Gunakan versi Go yang memenuhi `go.mod` (`go 1.25.7`) dan ekspor environment terlebih dahulu. Contoh PowerShell:

```powershell
$env:PORT = ':8080'
$env:MAIN_DBSERVER = '127.0.0.1'
$env:MAIN_DBPORT = '1433'
$env:MAIN_DBUSER = 'api_user'
$env:MAIN_DBPASSWORD = 'ganti-dengan-password-db'
$env:MAIN_DBNAME = 'CORE_KSP'
$env:JWTKEY = 'ganti-dengan-secret-jwt'
$env:SIGNSECRET = 'ganti-dengan-secret-signature'
go run .
```

Saat berjalan langsung, API tersedia di `http://localhost:8080`. Jika route baru memakai Redis, pastikan alamat `redis:6379` pada kode dapat diakses dari lingkungan proses Go tersebut; hostname service Compose biasanya hanya tersedia di jaringan Compose.

## Timeout dan shutdown

Server menggunakan timeout header 5 detik, baca request 15 detik, tulis response 60 detik, dan idle 60 detik. Upload besar tetap harus selesai dibaca dalam jendela read timeout. Write timeout tidak menghentikan eksekusi handler secara otomatis; operasi bisnis perlu context sendiri.

Pada interrupt/SIGTERM, server menunggu request aktif maksimal 20 detik, lalu flush antrean log maksimal 10 detik. Compose memberi `stop_grace_period: 35s`. Jika waktu habis atau proses dihentikan paksa, pekerjaan yang belum selesai bisa hilang.

Startup menjalankan `db.InitDB()` dan ping database. Kegagalan startup DB menghentikan aplikasi. Ping startup belum menggunakan timeout context eksplisit dari aplikasi.
