# Cache Redis

Implementasi aktif ada di `cache/redis.go`. Cache memakai Redis bersama, bukan map in-memory per proses Go. Belum ada external token manager di repository saat ini.

## Koneksi

Client dibuat dengan konfigurasi tetap:

```go
redis.Options{
    Addr:     "redis:6379",
    Password: "",
    DB:       0,
}
```

Alamat ini cocok dengan service `redis` di Compose. Untuk menjalankan API langsung di host, pastikan hostname tersebut dapat di-resolve atau sesuaikan konfigurasi client. Client belum membaca environment Redis dan belum melakukan ping startup secara eksplisit.

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

Untuk key yang tidak ditemukan/expired, GET mengembalikan zero value, `false`, `nil`. Error koneksi atau JSON yang tidak cocok mengembalikan zero value, `false`, dan error. String juga disimpan melalui JSON, bukan sebagai raw string.

Jika menyimpan struct, gunakan field exported dan JSON tag sesuai kebutuhan. Tangani error cache sesuai kebutuhan bisnis; jangan menganggap `found=false` selalu berarti key memang tidak ada jika `err` terisi.

## Pemakaian yang tersedia

`auth.RequestIdAuth()` menggunakan key `request-id:<X-Request-Id>` dengan TTL 5 menit. Middleware ini tersedia tetapi belum dipasang pada route aktif. Pemeriksaan GET lalu SET belum atomik, dan kegagalan Redis belum membuat middleware menolak request. Detail ada di [autentikasi](internal-auth-response-helper.md).

Belum ada helper cache lama `Set/Get/Delete`, in-memory cache, atau manager refresh token eksternal. Untuk fitur baru, gunakan API Redis yang memang tersedia atau tambahkan implementasi beserta dokumentasinya.

## Lifecycle

Redis di Compose belum memakai volume persisten. Restart proses API saja tidak menghapus cache Redis, tetapi data tidak dijamin bertahan saat container Redis dihapus/dibuat ulang. TTL dikelola Redis. Beberapa instance API dapat membaca key yang sama jika menunjuk ke server/database Redis yang sama.
