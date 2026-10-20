# Logging HTTP

Panduan ini menjelaskan middleware aktif, mapping SQL Server, batas cuplikan body, serta perilaku antrean dan shutdown. Konfigurasi menjalankan service ada di [panduan aplikasi](menjalankan-aplikasi.md).

## Alur middleware

Implementasi aktif ada di `helper/http_logger.go`. Urutan middleware di `main.go`:

```go
router.Use(gin.Logger(), helper.ResponseLogger(), gin.Recovery(), helper.RequestLogger(), helper.ContextCleaner())
```

`RequestLogger()` menyiapkan data `LOG_HTTP`, kemudian memasukkan pekerjaan insert ke antrean background. Handler tidak menunggu SQL Server. `ResponseLogger()` memasukkan pekerjaan update untuk tiga kolom response setelah handler/recovery selesai. Satu worker memproses antrean sesuai urutan, sehingga update selalu diproses setelah insert request terkait. Identitas hasil insert hanya diakses worker. Jika insert gagal atau ID tidak valid, update dilewati.

Setiap operasi database mendapat context independen dengan timeout 5 detik saat pekerjaan mulai diproses. Waktu menunggu di antrean tidak termasuk timeout ini. Pembatalan context proses bisnis maupun koneksi client tidak membatalkan penyimpanan log.

## Tabel LOG_HTTP

Mapping berada di `obj/obj_db.go`. Worker menulis ke `db.Main()`, sehingga tabel berada di database `MAIN_DBNAME`.

| Kolom | Pengisian |
| --- | --- |
| `logid` | `int64`, primary key auto increment; field struct disembunyikan dari JSON dengan `json:"-"` |
| `request_time` | Waktu sebelum logger membaca cuplikan request |
| `request_endpoint` | `c.FullPath()` atau fallback `c.Request.URL.Path` |
| `request_text` | Dump request line/header dan cuplikan body |
| `response_time` | Waktu setelah handler/recovery selesai, sebelum pekerjaan update masuk antrean |
| `response_code` | Status HTTP dan teksnya, misalnya `200 - OK` |
| `response_text` | Cuplikan body yang ditulis melalui writer middleware |
| `client_ip` | Hasil `c.ClientIP()` |

Semua field selain `LOGID` bertipe pointer. Kolom response perlu menerima NULL saat insert request awal. Tabel yang sudah ada perlu menyediakan `logid BIGINT IDENTITY(1,1)` sebagai primary key dan kolom lain dengan tipe yang sesuai. Gunakan kapasitas teks yang cukup untuk header + cuplikan, misalnya `NVARCHAR(MAX)` bagi request/response text. Kode belum menjalankan migrasi otomatis; tag GORM saja tidak mengubah tabel yang sudah ada.

`GET /items/123` yang cocok route `/items/:id` menyimpan `/items/:id` sebagai endpoint; method dan URL aktual tetap terlihat di request text. Endpoint bukan nama fungsi handler. Untuk path tidak ditemukan seperti `/missing`, endpoint menyimpan `/missing`, bukan NULL. Status 404 yang sengaja dikembalikan route terdaftar tetap memakai pola route tersebut.

Gin context menyimpan pointer identitas internal pada `LogApiCtxKey`. Pointer ini hanya dibaca/diisi worker, sehingga controller tidak boleh memakai nilainya sebagai ID insert yang sudah pasti tersedia. Antrean HTTP tidak dipakai oleh helper audit `LogHisQuery()`.

## Batas dan body

Batas yang digunakan:

| Pengaturan | Nilai |
| --- | --- |
| Ukuran body request untuk controller | Tidak dibatasi oleh logger |
| Cuplikan body request/response dalam log | 64 KiB; sisanya diberi penanda pemotongan |
| Kapasitas antrean | 128 pekerjaan insert/update, satu worker |
| Timeout baca header HTTP | 5 detik |
| Timeout baca seluruh request HTTP | 15 detik |
| Timeout tulis response HTTP | 60 detik |
| Timeout koneksi idle | 60 detik |

Logger membaca maksimal 64 KiB + 1 byte dari request, lalu mengembalikan byte tersebut di depan stream body asli. Controller tetap membaca body lengkap, termasuk request di atas 2 MiB. Close pada body tetap diteruskan ke stream asli. Body response ke client tidak dipotong. Kesalahan pembacaan cuplikan body ditolak dengan 400, atau 408 jika terjadi read timeout. Kesalahan baca pada sisa body ditangani controller. Timeout baca seluruh request tetap 15 detik, termasuk saat controller membaca sisa body. HTTP write timeout membatasi penulisan ke koneksi; bukan batas waktu eksekusi handler. Operasi bisnis tetap perlu menggunakan context sendiri.

Recovery berada di dalam `ResponseLogger()`, sehingga status/body setelah recovery ikut dicatat. Jika response sudah terlanjur dikirim sebelum panic, status yang dicatat mengikuti status yang sudah terkirim. `ContextCleaner()` menggunakan `defer` agar context proses bisnis dibersihkan juga saat panic. Handler default 404/405 menulis body di dalam rantai middleware agar body tersebut dapat ditangkap. `NoMethod` baru dipakai jika `HandleMethodNotAllowed` diaktifkan; konfigurasi sekarang belum mengaktifkannya.

Logger menangkap body lewat `Write`/`WriteString`, bukan isi koneksi yang di-hijack seperti WebSocket. Logger belum melakukan redaksi header/body; data Authorization dan payload sensitif dapat ikut masuk cuplikan raw request. Atur akses tabel sesuai data yang benar-benar dikirim aplikasi.

## Antrean dan shutdown

Antrean penuh akan melewati pekerjaan baru dan mencetak error ke stderr. Jika pekerjaan update dilewati, row request bisa tertinggal tanpa response. Antrean ini berada di RAM, sehingga log belum tersimpan bisa hilang saat crash atau penghentian paksa. Kebijakan ini menjaga request tetap berjalan saat database log bermasalah.

Saat SIGTERM/interrupt, server menunggu request aktif maksimal 20 detik, kemudian menunggu antrean log maksimal 10 detik. Compose menyediakan `stop_grace_period: 35s`. Antrean yang belum selesai setelah batas tersebut tidak dijamin tersimpan. Logging ini belum menggunakan antrean persisten atau retry database.

## Verifikasi

`helper/http_logger_test.go` memeriksa request normal, body di atas 3 MiB tetap lengkap di handler, path 404, recovery panic, cuplikan log, worker tertahan, antrean penuh, ID insert tidak valid, dan timeout baca HTTP. Worker persistence diganti dalam tes middleware; tes tersebut tidak memverifikasi schema atau koneksi SQL Server aktual.

```sh
go test ./...
go vet ./...
```
