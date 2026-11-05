# Cache Redis

Implementasi aktif ada di `cache/redis.go`. Cache memakai Redis bersama, bukan map in-memory per proses Go. Belum ada external token manager di repository saat ini.

## Koneksi

Client dibuat dengan konfigurasi dinamis berbasis environment variables:

```go
Addr:     fmt.Sprintf("%s:%s", host, port), // default "redis:6379" atau baca REDIS_ADDR / REDIS_HOST & REDIS_PORT
Password: password,                          // default "" atau baca REDIS_PASSWORD
DB:       dbNum,                             // default 0 atau baca REDIS_DB
```

Variabel yang didukung di `.env`:
* `REDIS_HOST`: Nama host atau IP server Redis (default: `"redis"`).
* `REDIS_PORT`: Port server Redis (default: `"6379"`).
* `REDIS_ADDR`: Alamat lengkap `host:port` (opsional, jika diset akan mengabaikan host & port terpisah).
* `REDIS_PASSWORD`: Password autentikasi Redis (default: kosong).
* `REDIS_DB`: Indeks database Redis (default: `0`).

Saat startup, `main.go` menjalankan `cache.InitRedisClient()` dan melakukan uji koneksi eksplisit via `cache.PingRedis(3 * time.Second)`.

Semua helper memakai `cache.Ctx_bg`, yaitu `context.Background()`. Pembatalan request HTTP tidak otomatis membatalkan operasi Redis. Timeout jaringan mengikuti konfigurasi/default library Redis; tidak ada timeout context per pemanggilan pada wrapper saat ini.

## Operasi

```go
err := cache.RedisSet("contoh:key", "nilai", 5*time.Minute)
value, found, err := cache.RedisGet[string]("contoh:key")
err = cache.RedisDel("contoh:key")
```

Contoh tersebut membutuhkan import package `cache` dan `time`.

| Fungsi | Perilaku |
| --- | --- |
| `RedisSet[T](key, value, ttl)` | Marshal JSON lalu SET Redis dengan TTL; mengembalikan error marshal/Redis |
| `RedisGet[T](key)` | Mengembalikan nilai, found, error; tipe `T` harus sesuai JSON tersimpan |
| `RedisDel(key)` | Menghapus key dan mengembalikan error Redis |
| `PingRedis(timeout)` | Memeriksa konektivitas ke Redis dengan batas waktu timeout |
| `CloseRedis()` | Menutup koneksi client Redis |

Untuk key yang tidak ditemukan/expired, GET mengembalikan zero value, `false`, `nil`. Error koneksi atau JSON yang tidak cocok mengembalikan zero value, `false`, dan error. String juga disimpan melalui JSON, bukan sebagai raw string.

Jika menyimpan struct, gunakan field exported dan JSON tag sesuai kebutuhan. Tangani error cache sesuai kebutuhan bisnis; jangan menganggap `found=false` selalu berarti key memang tidak ada jika `err` terisi.

## Pemakaian yang tersedia

* **Session Login Anggota** (`controller/anggota.go`): Menyimpan key session anggota saat login dengan TTL 21 menit (`keySessionId`). Logout menghapus key melalui `RedisDel`.
* **Idempotency Request** (`auth.RequestIdAuth()`): Menggunakan key `request-id:<X-Request-Id>` dengan TTL 5 menit untuk mencegah replaying request.

## Lifecycle & Persistensi

* **Volume Persisten**: Service Redis di `compose.yaml` menggunakan volume `redis_data:/data` dan mengaktifkan Append-Only File (`redis-server --appendonly yes`). Data sesi dan cache bertahan meskipun container di-recreate/restart.
* **Healthcheck**: Service Redis dilengkapi healthcheck berkala (`redis-cli ping`). Service `api` memiliki konfigurasi `depends_on: redis: condition: service_healthy` sehingga API hanya mulai menerima request setelah socket Redis siap.
* Beberapa instance API (`api-1`, `api-2`) dapat berbagi satu container Redis melalui network internal Compose.
