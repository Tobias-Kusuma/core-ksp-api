# Autentikasi dan response helper

Implementasi ada di `auth/auth_internal.go` dan `helper/helper.go`. Semua middleware di bawah tersedia untuk dipasang pada route baru; dua route aktif (`GET /` dan `GET /db-health`) masih public.

## Middleware yang tersedia

| Middleware | Pemeriksaan | Context yang diisi |
| --- | --- | --- |
| `JwtAuth()` | Header Bearer, validitas JWT HS256, tipe access, kelengkapan identitas | `token-string`, `detail-cust` |
| `TimestampAuth()` | Header `X-Timestamp`, format RFC3339, selisih absolut <= 5 menit | Tidak ada |
| `RequestIdAuth()` | Header `X-Request-Id`, duplikasi key Redis | `requestid` |
| `SignatureAuth()` | Header `X-Signature`, HMAC atas method/body/token/request ID | Tidak ada |

Untuk memakai semuanya:

```go
secureGroup.Use(auth.JwtAuth(), auth.TimestampAuth(), auth.RequestIdAuth(), auth.SignatureAuth())
```

Import package `auth` dan buat grup route terlebih dahulu. `JwtAuth()` serta `RequestIdAuth()` harus berjalan sebelum `SignatureAuth()` supaya nilai context tersedia. Endpoint signed tanpa JWT menggunakan token kosong dalam string signature; request ID tetap harus diperiksa sebelum signature.

## JWT

`main.go` memanggil `auth.SetJwtKey()` dan `auth.SetSignSecret()` untuk membaca `JWTKEY` serta `SIGNSECRET` dari environment proses. `.env` tidak otomatis dibaca saat menjalankan Go langsung.

```go
token, err := auth.GenerateToken(deviceid, phone, userid, kdcab, cno, username)
claims, err := auth.ValidateToken(token)
```

`GenerateToken()` membuat access token HS256 dengan masa berlaku 10 menit dan issuer `github.com/Tobias-Kusuma/core-ksp-api`. Claims khusus berisi `deviceid`, `nohp`, `id`, `kdcab`, `cno`, `nama`, dan `token_type`.

`ValidateToken()` memeriksa algoritma HS256 dan validitas parsing/claims JWT. Fungsi ini tidak mengonfigurasi pemeriksaan issuer/audience tertentu. `JwtAuth()` menambahkan pemeriksaan `token_type == "access"` dan mewajibkan keenam field identitas tidak kosong.

Header client:

```http
Authorization: Bearer <access-token>
```

Identitas hasil middleware dapat dibaca sebagai `*obj.DetailCust` dari key `detail-cust`. Belum ada endpoint penerbitan token/refresh token yang aktif.

## Timestamp dan request ID

Contoh format timestamp:

```http
X-Timestamp: 2026-10-20T10:00:00+07:00
X-Request-Id: <id-unik-per-request>
```

Timestamp contoh harus diganti waktu request aktual. Selisih absolut lebih dari 5 menit ditolak, baik timestamp terlalu lama maupun terlalu jauh di masa depan.

`RequestIdAuth()` memakai key Redis `request-id:<nilai-header>` dan TTL 5 menit. Implementasi melakukan GET lalu SET terpisah; dua request bersamaan belum dijamin ditolak secara atomik. Error GET tidak otomatis menolak request dan error SET saat ini tidak ditangani. Jadi middleware ini belum memberi jaminan anti replay saat Redis gagal atau terjadi race. Lihat [cache Redis](cache-redis.md).

## Signature

Algoritma mengikuti kode aktif:

```text
bodyHash     = hex_lowercase(SHA256(raw_body_bytes))
stringToSign = METHOD + ":" + bodyHash + ":" + accessToken + ":" + requestId
signature    = hex_lowercase(HMAC_SHA256(SIGNSECRET, stringToSign))
```

Kirim hasil di `X-Signature`. `METHOD` berasal dari method request aktual, token tidak menyertakan prefix `Bearer `, dan request ID memakai nilai header yang disimpan middleware. Jika JWT/request ID tidak dipasang, nilai yang tidak ada menjadi string kosong.

URL/path dan timestamp tidak termasuk string signature saat ini. Client harus menandatangani byte body yang persis dikirim; perubahan spasi atau urutan properti JSON dapat mengubah hash.

`SignatureAuth()` membaca seluruh body lalu mengembalikannya agar handler dapat membaca ulang. Batas cuplikan logger tidak memotong body signature/controller. Middleware signature sendiri belum membatasi alokasi body lengkap.

## Response helper

| Fungsi | HTTP status | Body |
| --- | --- | --- |
| `RespondMaskedOk(c, message, data)` | 200 | `{"message": ..., "data": ...}` |
| `RespondMaskedError(c, true, err)` | 400 | `{"message": err.Error()}` |
| `RespondMaskedError(c, false, err)` | Umumnya 400 setelah masking | `{"message": pesan_masking}` |
| `RespondErrorUnauthorized(c, err)` | 401 | `{"message": err.Error()}` |

Helper error memanggil `AbortWithStatusJSON`; handler tetap sebaiknya langsung `return`. Helper sukses memakai `c.JSON`, bukan abort.

`MaskSQLError()` mengenali duplikasi unique/PK, string terlalu panjang, konversi tipe/tanggal, overflow numerik, serta fallback general. Nilai `custom=false` tidak otomatis berarti HTTP 500: error non-nil yang tidak dikenali pun dipetakan menjadi `MaskedError` general dan biasanya dikirim sebagai 400. Jangan mengoper `nil` sebagai error ke helper response error.

Error pembacaan cuplikan body dari logger memakai `{"error": ...}` dengan status 400/408. Recovery panic memakai perilaku Gin, sedangkan 404/405 memakai teks. Jadi tidak semua error global memakai envelope `message`.
