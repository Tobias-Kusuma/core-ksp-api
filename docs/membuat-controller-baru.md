# Membuat controller baru

Route aktif saat ini memakai handler inline di `routes/route.go`. Package `controller` belum ada. Contoh berikut adalah pola penambahan endpoint baru, bukan route yang sudah tersedia.

## Struktur request dan response

Definisikan request di `obj/obj_request.go`, response di `obj/obj_response.go`, dan mapping tabel bila diperlukan di `obj/obj_db.go`. Ikuti pola field exported huruf besar dan tag JSON/GORM yang sudah ada.

Contoh request baru:

```go
type ReqEcho struct {
    MESSAGE string `json:"message" binding:"required"`
}
```

Contoh response baru:

```go
type ResEcho struct {
    MESSAGE string `json:"message"`
}
```

Payload untuk contoh ini adalah JSON langsung:

```json
{"message":"Halo"}
```

Belum ada wrapper request generik pada kode aktif. Jangan mengasumsikan format `request`/`body` tersedia tanpa menambahkan struct dan parsing yang sesuai.

## Handler tanpa database

Untuk handler yang mulai terpisah dari route, buat file `controller/echo.go` dengan pola factory `gin.HandlerFunc`:

```go
package controller

import (
    "fmt"

    "github.com/Tobias-Kusuma/core-ksp-api/helper"
    "github.com/Tobias-Kusuma/core-ksp-api/obj"
    "github.com/gin-gonic/gin"
)

func Echo() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ReqBody obj.ReqEcho
        if err := c.ShouldBindJSON(&ReqBody); err != nil {
            helper.RespondMaskedError(c, true, fmt.Errorf("Message wajib diisi"))
            return
        }

        ResBody := obj.ResEcho{MESSAGE: ReqBody.MESSAGE}
        helper.RespondMaskedOk(c, "Berhasil", ResBody)
    }
}
```

Tambahkan import package controller ke file route, lalu daftarkan di grup yang sesuai:

```go
defaultGroup.POST("/echo", controller.Echo())
```

Response sukses:

```json
{"message":"Berhasil","data":{"message":"Halo"}}
```

Untuk endpoint yang memakai database, ganti parsing langsung dengan `helper.Get_DB_ReqBody(c, timeout, &ReqBody)` dan gunakan `AppTx.Mobile`. [Panduan database](get-db-reqbody-database.md) menjelaskan timeout, transaksi, dan field AppTx yang sudah diisi.

## Middleware endpoint

Dua endpoint aktif masih public. Untuk endpoint baru yang membutuhkan pengamanan, pasang middleware sesuai kebutuhan:

```go
secureGroup := routerEngine.Group("/secure")
secureGroup.Use(auth.JwtAuth(), auth.TimestampAuth(), auth.RequestIdAuth(), auth.SignatureAuth())
secureGroup.POST("/echo", controller.Echo())
```

Contoh di atas memerlukan import package `auth` dan handler/struct contoh tadi. Urutan ini membuat token serta request ID tersedia sebelum pemeriksaan signature. Algoritma dan header dijelaskan di [autentikasi](internal-auth-response-helper.md).

## Logging dan verifikasi

Middleware global mencatat endpoint secara otomatis menggunakan pola route seperti `/items/:id`; fallback memakai path aktual untuk route yang tidak ditemukan. Controller tidak mengubah ID atau data log HTTP secara langsung.

Logger hanya membatasi cuplikan body dalam log, bukan body yang diterima handler. Validasi payload bisnis dan batas khusus endpoint, bila diperlukan, tetap menjadi tanggung jawab handler/middleware endpoint.

Setelah menambahkan route, perbarui daftar route pada README dan jalankan:

```sh
go test ./...
go vet ./...
```

Tes handler baru sebaiknya memeriksa perilaku yang bermakna: validasi request, response yang diterima client, akses ditolak, atau transaksi yang harus rollback. Jangan menggunakan tabel atau endpoint contoh seolah sudah disediakan aplikasi.
