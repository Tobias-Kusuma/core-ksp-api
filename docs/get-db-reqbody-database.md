# Get_DB_ReqBody dan database

## Pool utama

`database/database.go` menyediakan `db.InitDB()` dan `db.Main()`. Startup memanggil `InitDB()` untuk membuat pool SQL Server dari `MAIN_DBSERVER`, `MAIN_DBPORT`, `MAIN_DBUSER`, `MAIN_DBPASSWORD`, dan `MAIN_DBNAME`, lalu melakukan ping.

| Pengaturan pool | Nilai |
| --- | --- |
| Max open connections | 200 |
| Max idle connections | 75 |
| Connection max lifetime | 5 menit |

Connection lifetime bukan timeout query. Logging HTTP dan operasi bisnis saat ini berbagi pool utama. `db.Main()` mengembalikan pool GORM, bukan transaksi yang sudah di-`Begin`. Jika pool belum diinisialisasi, fungsi ini panic.

## Helper request

```go
func Get_DB_ReqBody(c *gin.Context, timeout int, ReqBody any) (*obj.AppTx, error)
```

| Argumen | Perilaku |
| --- | --- |
| `c` | Gin context request aktif |
| `timeout` | Durasi context dalam detik; nilai <= 0 menjadi 10 detik |
| `ReqBody` | Pointer tujuan parsing JSON; `nil` untuk melewati parsing |

Urutan kerja helper:

1. Memanggil `c.ShouldBindJSON(ReqBody)` jika argumen tidak nil.
2. Setelah parsing selesai, membuat context turunan dari `c.Request.Context()` dengan timeout.
3. Menyimpan cancel function sebagai `cancelCtx` di Gin context.
4. Mengambil `db.Main()` dan mengembalikan `AppTx.Mobile` dengan `WithContext(ctx)`.

Karena timeout context baru dibuat setelah parsing, batas ini tidak mencakup waktu pembacaan/parsing JSON. Baca HTTP tetap dibatasi konfigurasi server. Error parsing dikembalikan bersama lokasi caller; caller perlu memilih response yang sesuai.

## Isi AppTx saat ini

| Field | Nilai dari Get_DB_ReqBody |
| --- | --- |
| `C` | Gin context |
| `Ctx` | Context dengan timeout proses bisnis |
| `Mobile` | `db.Main().WithContext(ctx)` |
| `Doc` | `nil`; belum diisi helper |
| `Log` | `nil`; belum diisi helper |

Nama `Mobile` masih dipakai pada struct, meskipun pool sumbernya bernama `Main`. Jangan mengakses `AppTx.Doc`/`AppTx.Log` seolah sudah terhubung. Tidak ada accessor aktif `db.Doc()` atau `db.Log()`.

Contoh di dalam handler:

```go
var ReqBody struct {
    NAME string `json:"name" binding:"required"`
}
AppTx, err := helper.Get_DB_ReqBody(c, 10, &ReqBody)
if err != nil {
    helper.RespondMaskedError(c, true, fmt.Errorf("Request tidak valid"))
    return
}

// Jalankan query melalui AppTx.Mobile agar context bisnis ikut dipakai.
// Gunakan placeholder untuk nilai dari client:
// AppTx.Mobile.Table("TABEL").Where("name = ?", ReqBody.NAME).Find(&result)
```

Contoh parsing tanpa database tidak perlu memakai helper ini; gunakan `c.ShouldBindJSON()` langsung. Untuk query tanpa JSON, panggil `Get_DB_ReqBody(c, 10, nil)`.

## Context dan transaksi

`ContextCleaner()` sudah terpasang global. Middleware ini memanggil `cancelCtx` melalui `defer`, termasuk saat handler panic. Gunakan `Get_DB_ReqBody()` satu kali per request: pemanggilan berulang akan menimpa cancel function yang tersimpan.

Context bisnis dapat selesai karena deadline, pembatalan request client, atau cleanup. Gunakan `AppTx.Ctx` untuk operasi lanjutan yang mendukung context. Jangan membawa Gin context ke goroutine setelah request selesai.

`AppTx` tidak otomatis memulai transaksi SQL. Jika beberapa query harus atomik, gunakan `AppTx.Mobile.Transaction(func(tx *gorm.DB) error { ... })`; gunakan `tx` untuk seluruh query di dalam callback agar kegagalan dapat di-rollback.

## Logging

Log request/response sudah dikelola middleware [logging HTTP](logging-http.md); controller tidak perlu mengisi identitas log atau memanggil fungsi update log sendiri.

`helper.LogHisQuery()` masih tersedia sebagai audit query terpisah ke `LOGHIS`, tetapi belum dipanggil route aktif. Fungsi ini menjalankan goroutine dengan session GORM baru dari DB yang diberikan. `NewDB: true` tidak otomatis membuat context background atau menjamin terlepas dari transaksi SQL induk; hindari mengandalkannya sebagai jaminan audit yang pasti tersimpan sesudah rollback/cancel. Antrean dan flush logging HTTP tidak mencakup fungsi ini.

Helper `BuildInsert...`, `BuildUpdate...`, dan `BuildDelete...` membentuk teks SQL. Untuk query bisnis baru, gunakan binding parameter GORM/SQL; jangan memasukkan input client ke SQL melalui penggabungan string.
