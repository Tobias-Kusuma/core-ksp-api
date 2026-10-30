# DOKUMENTASI SPESIFIKASI API CORE KSP

**Versi API:** 0.0.2-BETA  
**Tanggal Rilis:** 28 Oktober 2026  
**Format Payload:** JSON (`application/json`)  

---

## 1. Panduan Integrasi & Standar Komunikasi

### 1.1 Base URL
Seluruh endpoint pada dokumentasi ini diakses menggunakan Base URL sesuai lingkungan sistem:

| Lingkungan | Base URL | Keterangan |
|---|---|---|
| **Development / Testing** | `http://103.82.240.123:8081` | Server aktif untuk integrasi dan pengujian |
| **Production** | `-` | Server resmi lingkungan produksi |

### 1.2 Standar Request Header
Setiap permintaan HTTP wajib menyertakan header berikut:

```http
Content-Type: application/json
Accept: application/json
```

### 1.3 Format Standar Response
API Core KSP menggunakan struktur JSON envelope terstandarisasi untuk seluruh respon.

#### Format Respon Sukses (HTTP 200 OK)
```json
{
  "message": "Pesan deskripsi keberhasilan operasi",
  "data": { ... }
}
```
*Catatan: Pada beberapa endpoint aksi/perintah (seperti pendaftaran), objek `data` dapat bernilai `null` apabila hanya konfirmasi keberhasilan yang dikembalikan.*

#### Format Respon Gagal (HTTP 400 Bad Request / 500 Internal Server Error)
```json
{
  "message": "Deskripsi penyebab kegagalan atau validasi yang tidak terpenuhi"
}
```

### 1.4 Kode Status HTTP
- **`200 OK`**: Permintaan berhasil diproses.
- **`400 Bad Request`**: Kesalahan input data, parameter wajib tidak terisi, atau pelanggaran aturan bisnis (misal: saldo tidak mencukupi, cicilan tertua belum dibayar, dll).
- **`404 Not Found`**: Endpoint atau data identitas yang dicari tidak ditemukan.
- **`500 Internal Server Error`**: Terjadi kendala teknis pada layanan backend.

---

## 2. Ringkasan Daftar Endpoint

| Modul | Method | Endpoint Path | Deskripsi Singkat |
|---|:---:|---|---|
| **System** | `GET` | `/` | Cek status server API Core KSP |
| **System** | `GET` | `/db-health` | Cek status konektivitas database |
| **Anggota** | `POST` | `/anggota/pendaftaran` | Pendaftaran anggota baru & pembukaan simpanan awal |
| **Anggota** | `POST` | `/anggota/getdaftaranggota` | Mengambil daftar anggota berdasarkan cabang |
| **Anggota** | `POST` | `/anggota/getdataanggota` | Mengambil data profil lengkap anggota & rekening |
| **Simpanan** | `POST` | `/simpanan/getproduksimpanan` | Katalog produk simpanan/tabungan aktif |
| **Simpanan** | `POST` | `/simpanan/simpanbaru` | Pembukaan rekening simpanan/tabungan baru |
| **Simpanan** | `POST` | `/simpanan/transfersimpanan` | Transfer saldo antar rekening simpanan |
| **Simpanan** | `POST` | `/simpanan/getmutasi` | Riwayat mutasi transaksi rekening tabungan |
| **Simpanan** | `POST` | `/simpanan/getdaftarrekening` | Daftar seluruh rekening simpanan milik anggota |
| **Simpanan** | `POST` | `/simpanan/getdatarekening` | Data detail satu rekening simpanan |
| **Simpanan** | `POST` | `/simpanan/tutuprekening` | Penutupan rekening simpanan nasabah |
| **Pinjaman** | `POST` | `/pinjaman/getprodukpinjaman` | Katalog produk pinjaman/pembiayaan aktif |
| **Pinjaman** | `POST` | `/pinjaman/simulasi` | Kalkulator simulasi kredit & tabel estimasi cicilan |
| **Pinjaman** | `POST` | `/pinjaman/pinjamanbaru` | Pembukaan fasilitas pinjaman baru & generate jadwal |
| **Pinjaman** | `POST` | `/pinjaman/getjadwalangsuran` | Rincian tabel rencana jadwal angsuran pinjaman |
| **Pinjaman** | `POST` | `/pinjaman/getdaftarpinjaman` | Daftar seluruh pinjaman milik nasabah |
| **Pinjaman** | `POST` | `/pinjaman/getdatapinjaman` | Data detail satu fasilitas pinjaman nasabah |
| **Pinjaman** | `POST` | `/pinjaman/gettagihan` | Inquiry tagihan angsuran berjalan & denda keterlambatan |
| **Pinjaman** | `POST` | `/pinjaman/bayarangsuran` | Pembayaran angsuran kredit bulanan |
| **Pinjaman** | `POST` | `/pinjaman/pelunasankredit` | Pelunasan pinjaman dipercepat (*early settlement*) |
| **Pinjaman** | `POST` | `/pinjaman/getriwayatbayar` | Riwayat transaksi pembayaran & pelunasan pinjaman |

> [!NOTE]
> Endpoint `/simpanan/setorsimpanan` dan `/simpanan/tariksimpanan` saat ini dinonaktifkan sementara dan belum dipublikasikan untuk integrasi publik.

---

## 3. Modul System & Health Check

### 3.1 Cek Status Layanan API
Memeriksa apakah server backend API berjalan normal.

- **Method:** `GET`
- **Path:** `/`
- **Request Body:** Tidak ada

#### Contoh Request (cURL)
```bash
curl -X GET "http://103.82.240.123:8081/" \
  -H "Accept: application/json"
```

#### Contoh Response (HTTP 200)
```json
{
  "message": "Selamat datang di API Mobile - EasyBank",
  "data": null
}
```

---

### 3.2 Cek Koneksi Database
Memverifikasi status koneksi aktif antara server API dengan basis data.

- **Method:** `GET`
- **Path:** `/db-health`
- **Request Body:** Tidak ada

#### Contoh Request (cURL)
```bash
curl -X GET "http://103.82.240.123:8081/db-health" \
  -H "Accept: application/json"
```

#### Contoh Response (HTTP 200)
```json
{
  "message": "Berhasil terhubung dengan database",
  "data": null
}
```

---

## 4. Modul Anggota (Membership)

### 4.1 Pendaftaran Anggota Baru
Mendaftarkan anggota koperasi baru ke dalam sistem dan secara otomatis membuka 3 rekening simpanan wajib awal (Simpanan Pokok, Simpanan Wajib, dan Simpanan Sukarela).

- **Method:** `POST`
- **Path:** `/anggota/pendaftaran`

#### Parameter Request (JSON Body)
- `anggota` (Object, Wajib): Data identitas diri nasabah.
  - `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
  - `Tipenas` (String, Wajib): Tipe nasabah (contoh: `"1"`).
  - `jnsnas` (String, Wajib): Jenis nasabah (contoh: `"1"`).
  - `cnm` (String, Wajib): Nama lengkap nasabah sesuai KTP.
  - `tgllhr` (String, Wajib): Tanggal lahir dengan format `YYYY-MM-DD`.
  - `umur` (Integer, Wajib): Usia nasabah.
  - `tmplhr` (String, Wajib): Kota/tempat lahir.
  - `jnskel` (String, Wajib): Jenis kelamin (`"L"` = Laki-laki, `"P"` = Perempuan).
  - `JnsID` (String, Wajib): Jenis tanda pengenal (contoh: `"KTP"`).
  - `marstat` (String, Wajib): Status pernikahan (`"1"` = Menikah, `"2"` = Belum Menikah, `"3"` = Duda/Janda).
  - `agama` (String, Wajib): Agama nasabah (contoh: `"ISLAM"`).
  - `warga` (String, Wajib): Kewarganegaraan (`"1"` = WNI, `"2"` = WNA).
  - `NoID` (String, Wajib): Nomor identitas KTP (16 digit).
  - `npwp` (String, Wajib): Nomor Pokok Wajib Pajak (NPWP).
  - `AlamatID` (String, Wajib): Alamat sesuai KTP.
  - `KelID` (String, Wajib): Kelurahan sesuai KTP.
  - `KecID` (String, Wajib): Kecamatan sesuai KTP.
  - `KotaID` (String, Wajib): Kota sesuai KTP.
  - `KdposID` (String, Wajib): Kode pos sesuai KTP.
  - `AlamatSkr` (String, Wajib): Alamat domisili sekarang.
  - `KelSkr` (String, Wajib): Kelurahan domisili sekarang.
  - `KecSkr` (String, Wajib): Kecamatan domisili sekarang.
  - `KotaISkr` (String, Wajib): Kota domisili sekarang.
  - `KdposSkr` (String, Wajib): Kode pos domisili sekarang.
  - `stsrumah` (String, Wajib): Status kepemilikan tempat tinggal (`"1"` = Milik Sendiri, `"2"` = Sewa/Kontrak, `"3"` = Ikut Keluarga).
  - `nohp` (String, Wajib): Nomor telepon / WhatsApp aktif.
  - `email` (String, Wajib): Alamat surat elektronik nasabah.
  - `nmibu` (String, Wajib): Nama gadis ibu kandung.
  - `stspokok` (String, Wajib): Status simpanan pokok (`"Y"` / `"N"`).
  - `stswajib` (String, Wajib): Status simpanan wajib (`"Y"` / `"N"`).
- `simpanan_pokok` (Object, Wajib):
  - `kdcab` (String, Wajib): Kode cabang.
  - `jenis` (String, Wajib): Kode jenis produk simpanan pokok (contoh: `"01"`).
  - `nlsetor` (Number, Wajib): Besaran setoran simpanan pokok.
- `simpanan_wajib` (Object, Wajib):
  - `kdcab` (String, Wajib): Kode cabang.
  - `jenis` (String, Wajib): Kode jenis produk simpanan wajib (contoh: `"02"`).
  - `nlsetor` (Number, Wajib): Besaran setoran simpanan wajib bulanan.
- `simpanan_sukarela` (Object, Wajib):
  - `kdcab` (String, Wajib): Kode cabang.
  - `jenis` (String, Wajib): Kode jenis produk simpanan sukarela (contoh: `"03"`).
  - `nlsetor` (Number, Wajib): Setoran awal simpanan sukarela (bisa `0`).

#### Contoh Request Body (JSON)
```json
{
  "anggota": {
    "kdcab": "001",
    "cno": "",
    "Tipenas": "1",
    "jnsnas": "1",
    "cnm": "Budi Santoso",
    "tgllhr": "1995-08-17",
    "umur": 31,
    "tmplhr": "Jakarta",
    "jnskel": "L",
    "JnsID": "KTP",
    "marstat": "1",
    "agama": "ISLAM",
    "warga": "1",
    "NoID": "3171011708950001",
    "npwp": "012345678901000",
    "AlamatID": "Jl. Merdeka No. 10",
    "KelID": "Gambir",
    "KecID": "Gambir",
    "KotaID": "Jakarta Pusat",
    "KdposID": "10110",
    "AlamatSkr": "Jl. Merdeka No. 10",
    "KelSkr": "Gambir",
    "KecSkr": "Gambir",
    "KotaISkr": "Jakarta Pusat",
    "KdposSkr": "10110",
    "stsrumah": "1",
    "nohp": "081234567890",
    "email": "budi.santoso@example.com",
    "nmibu": "Siti Aminah",
    "stspokok": "Y",
    "stswajib": "Y"
  },
  "simpanan_pokok": {
    "kdcab": "001",
    "cno": "",
    "jenis": "01",
    "nlsetor": 100000
  },
  "simpanan_wajib": {
    "kdcab": "001",
    "cno": "",
    "jenis": "02",
    "nlsetor": 50000
  },
  "simpanan_sukarela": {
    "kdcab": "001",
    "cno": "",
    "jenis": "03",
    "nlsetor": 0
  }
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil membuat anggota baru",
  "data": null
}
```

#### Contoh Response Gagal (HTTP 400)
```json
{
  "message": "Data pendaftaran anggota tidak lengkap"
}
```

---

### 4.2 Get Daftar Anggota
Mengambil daftar ringkas anggota yang terdaftar pada suatu kantor cabang.

- **Method:** `POST`
- **Path:** `/anggota/getdaftaranggota`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang koperasi (contoh: `"001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan daftar anggota",
  "data": [
    {
      "kdcab": "001",
      "cno": "0000000001",
      "tipenas": "01",
      "jnsnas": "A",
      "cnm": "OTTO"
    },
    {
      "kdcab": "001",
      "cno": "0000000002",
      "tipenas": "01",
      "jnsnas": "A",
      "cnm": "Budi Santoso"
    }
  ]
}
```

---

### 4.3 Get Data Detail Anggota
Mengambil informasi lengkap profil seorang anggota beserta seluruh daftar rekening simpanan yang dimilikinya.

- **Method:** `POST`
- **Path:** `/anggota/getdataanggota`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `cno` (String, Wajib): Nomor anggota (contoh: `"0000000001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "cno": "0000000001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan detail anggota",
  "data": {
    "anggota": {
      "kdcab": "001",
      "cno": "0000000001",
      "Tipenas": "01",
      "jnsnas": "A",
      "cnm": "OTTO",
      "tgllhr": "2000-01-01T00:00:00Z",
      "umur": 26,
      "tmplhr": "Jakarta",
      "jnskel": "L",
      "JnsID": "1",
      "marstat": "1",
      "agama": "ISLAM",
      "warga": "1",
      "NoID": "3171010101000001",
      "npwp": "012345678901000",
      "AlamatID": "Jl. Tebet Raya no. 89",
      "KelID": "Tebet Timur",
      "KecID": "Tebet",
      "KotaID": "Jakarta Selatan",
      "KdposID": "12820",
      "AlamatSkr": "Jl. Tebet Raya no. 89",
      "KelSkr": "Tebet Timur",
      "KecSkr": "Tebet",
      "KotaISkr": "Jakarta Selatan",
      "KdposSkr": "12820",
      "stsrumah": "1",
      "nohp": "081234567890",
      "email": "otto@example.com",
      "nmibu": "Siti Aminah",
      "stspokok": "Y",
      "stswajib": "Y",
      "status": "1"
    },
    "daftar_rekening": [
      {
        "kdcab": "001",
        "noac": "0010100001",
        "cno": "0000000001",
        "aliasnm": "OTTO",
        "jenis": "01",
        "sldakhir": 100000,
        "sldpakai": 100000,
        "Status": "1"
      },
      {
        "kdcab": "001",
        "noac": "0010200001",
        "cno": "0000000001",
        "aliasnm": "OTTO",
        "jenis": "02",
        "sldakhir": 50000,
        "sldpakai": 50000,
        "Status": "1"
      },
      {
        "kdcab": "001",
        "noac": "0011000001",
        "cno": "0000000001",
        "aliasnm": "OTTO",
        "jenis": "10",
        "sldakhir": 3950000,
        "sldpakai": 3930000,
        "Status": "1"
      }
    ]
  }
}
```

---

## 5. Modul Simpanan (Savings)

### 5.1 Get Katalog Produk Simpanan
Mengambil seluruh katalog produk simpanan/tabungan yang aktif pada sistem beserta ketentuan batas saldo, setoran minimal, biaya administrasi, dan denda.

- **Method:** `POST`
- **Path:** `/simpanan/getproduksimpanan`
- **Request Body:** `{}` (tanpa parameter wajib)

#### Contoh Request Body (JSON)
```json
{}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan daftar produk simpanan",
  "data": [
    {
      "jenis": "01",
      "nmproduk": "SIMPANAN POKOK",
      "jnsprd": "Tabungan",
      "jnstipe": "Pokok",
      "nlbiaya": 0,
      "minsetor": 100000,
      "wjbsetor": 100000,
      "minsld": 0,
      "maxtarik": 0,
      "biayatutup": 0,
      "penalty": 0,
      "pajak": 0,
      "status": "1"
    },
    {
      "jenis": "02",
      "nmproduk": "SIMPANAN WAJIB",
      "jnsprd": "Tabungan",
      "jnstipe": "Wajib",
      "nlbiaya": 0,
      "minsetor": 50000,
      "wjbsetor": 50000,
      "minsld": 0,
      "maxtarik": 0,
      "biayatutup": 0,
      "penalty": 0,
      "pajak": 0,
      "status": "1"
    },
    {
      "jenis": "10",
      "nmproduk": "TABUNGAN SUKARELA",
      "jnsprd": "Tabungan",
      "jnstipe": "Harian",
      "nlbiaya": 0,
      "minsetor": 10000,
      "wjbsetor": 0,
      "minsld": 20000,
      "maxtarik": 10000000,
      "biayatutup": 10000,
      "penalty": 0,
      "pajak": 0,
      "status": "1"
    }
  ]
}
```

---

### 5.2 Pembukaan Rekening Simpanan Baru
Membuka rekening tabungan/simpanan baru bagi anggota yang telah terdaftar. Sistem akan membuatkan nomor rekening (`noac`) baru secara otomatis.

- **Method:** `POST`
- **Path:** `/simpanan/simpanbaru`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `cno` (String, Wajib): Nomor anggota (contoh: `"0000000001"`).
- `jenis` (String, Wajib): Kode produk simpanan (contoh: `"10"`).
- `nlsetor` (Number, Opsional): Nominal setoran awal pembukaan rekening. Jika tidak diisi atau 0, akan menggunakan setoran wajib produk terkait.

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "cno": "0000000001",
  "jenis": "10",
  "nlsetor": 50000
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil membuka rekening simpanan baru",
  "data": {
    "nama": "OTTO",
    "noac": "0011000002"
  }
}
```

#### Contoh Response Gagal (HTTP 400)
```json
{
  "message": "Produk simpanan tidak ditemukan atau tidak aktif"
}
```

---

### 5.3 Transfer Antar Rekening Simpanan
Memindahkan saldo dari satu rekening tabungan ke rekening tabungan lain secara langsung dan atomik (debet pengirim dan kredit penerima).

- **Method:** `POST`
- **Path:** `/simpanan/transfersimpanan`

#### Parameter Request (JSON Body)
- `noac_asal` (String, Wajib): Nomor rekening sumber pengirim saldo.
- `noac_tujuan` (String, Wajib): Nomor rekening tujuan penerima saldo.
- `nominal` (Number, Wajib): Nilai saldo yang ditransfer (harus > 0).
- `keterangan` (String, Opsional): Berita atau catatan transfer.

#### Contoh Request Body (JSON)
```json
{
  "noac_asal": "0011000001",
  "noac_tujuan": "0010200001",
  "nominal": 25000,
  "keterangan": "Transfer simpanan"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil melakukan transfer simpanan",
  "data": {
    "noac_asal": "0011000001",
    "nama_pengirim": "OTTO",
    "noac_tujuan": "0010200001",
    "nama_penerima": "OTTO",
    "nominal": 25000,
    "sldakhir_asal": 3925000,
    "sldpakai_asal": 3905000,
    "keterangan": "Transfer simpanan",
    "waktu_transfer": "2026-10-28 17:00:00"
  }
}
```

#### Contoh Response Gagal (HTTP 400)
```json
{
  "message": "Saldo tidak mencukupi untuk melakukan transfer"
}
```

---

### 5.4 Get Mutasi Rekening Simpanan
Mengambil riwayat transaksi mutasi debet dan kredit pada suatu nomor rekening dalam rentang tanggal tertentu.

- **Method:** `POST`
- **Path:** `/simpanan/getmutasi`

#### Parameter Request (JSON Body)
- `noac` (String, Wajib): Nomor rekening simpanan.
- `tgl_awal` (String, Wajib): Tanggal awal pencarian mutasi format `YYYY-MM-DD`.
- `tgl_akhir` (String, Wajib): Tanggal akhir pencarian mutasi format `YYYY-MM-DD`.

#### Contoh Request Body (JSON)
```json
{
  "noac": "0011000001",
  "tgl_awal": "2026-01-01",
  "tgl_akhir": "2026-12-31"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan mutasi rekening",
  "data": [
    {
      "kdcab": "001",
      "noref": "TRF00120261028000001",
      "noac": "0011000001",
      "tipe_mutasi": "Debet",
      "nilai": 25000,
      "ket": "Transfer simpanan",
      "tgltrn": "2026-10-28T17:00:00Z"
    },
    {
      "kdcab": "001",
      "noref": "SET00120261001000001",
      "noac": "0011000001",
      "tipe_mutasi": "Kredit",
      "nilai": 4000000,
      "ket": "Setoran Awal",
      "tgltrn": "2026-10-01T08:30:00Z"
    }
  ]
}
```

---

### 5.5 Get Daftar Rekening Nasabah
Menampilkan daftar seluruh rekening simpanan yang dimiliki oleh seorang anggota beserta saldo saat ini.

- **Method:** `POST`
- **Path:** `/simpanan/getdaftarrekening`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `cno` (String, Wajib): Nomor anggota (contoh: `"0000000001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "cno": "0000000001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan daftar rekening",
  "data": [
    {
      "kdcab": "001",
      "noac": "0010100001",
      "cno": "0000000001",
      "aliasnm": "OTTO",
      "jenis": "01",
      "sldakhir": 100000,
      "sldpakai": 100000,
      "minsld": 0,
      "Status": "1"
    },
    {
      "kdcab": "001",
      "noac": "0011000001",
      "cno": "0000000001",
      "aliasnm": "OTTO",
      "jenis": "10",
      "sldakhir": 3925000,
      "sldpakai": 3905000,
      "minsld": 20000,
      "Status": "1"
    }
  ]
}
```

---

### 5.6 Get Data Detail Rekening Simpanan
Menampilkan data terperinci dari satu nomor rekening tertentu.

- **Method:** `POST`
- **Path:** `/simpanan/getdatarekening`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `noac` (String, Wajib): Nomor rekening simpanan (contoh: `"0011000001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "noac": "0011000001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan detail rekening",
  "data": {
    "kdcab": "001",
    "noac": "0011000001",
    "cno": "0000000001",
    "aliasnm": "OTTO",
    "jenis": "10",
    "blokir": "N",
    "sldawal": 0,
    "sldakhir": 3925000,
    "sldpakai": 3905000,
    "minsld": 20000,
    "maxtarik": 10000000,
    "Status": "1",
    "nmproduk": "TABUNGAN SUKARELA",
    "jnsprd": "Tabungan",
    "jnstipe": "Harian"
  }
}
```

---

### 5.7 Penutupan Rekening Simpanan
Menutup rekening simpanan nasabah, memotong biaya administrasi tutup rekening, serta memindahkan sisa dana ke rekening tujuan yang ditunjuk.

- **Method:** `POST`
- **Path:** `/simpanan/tutuprekening`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `noac` (String, Wajib): Nomor rekening yang hendak ditutup (contoh: `"0011000002"`).
- `noac_tujuan` (String, Opsional): Nomor rekening penampung sisa saldo yang dicairkan.
- `alasan` (String, Opsional): Keterangan / alasan penutupan rekening.

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "noac": "0011000002",
  "noac_tujuan": "0011000001",
  "alasan": "Tutup rekening atas permohonan nasabah"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil menutup rekening simpanan",
  "data": {
    "noac": "0011000002",
    "aliasnm": "OTTO",
    "saldo_terakhir": 50000,
    "biaya_tutup": 10000,
    "dana_dicairkan": 40000,
    "noac_tujuan": "0011000001",
    "tgl_tutup": "2026-10-28 17:15:00",
    "status": "0"
  }
}
```

---

## 6. Modul Pinjaman (Loans)

### 6.1 Get Katalog Produk Pinjaman
Mengambil seluruh produk pinjaman/pembiayaan yang tersedia di koperasi beserta rumus perhitungan bunga, jenis angsuran, suku bunga acuan (*baserate*), dan keharusan jaminan.

- **Method:** `POST`
- **Path:** `/pinjaman/getprodukpinjaman`
- **Request Body:** `{}` (tanpa parameter wajib)

#### Contoh Request Body (JSON)
```json
{}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan daftar produk pinjaman",
  "data": [
    {
      "kd_produk": "50",
      "nm_produk": "KREDIT MULTIGUNA",
      "flgjmn": "Y",
      "jnsang": "Pokok + Bunga",
      "rmsbunga": "FLAT",
      "baserate": 12,
      "bulat": 100
    },
    {
      "kd_produk": "51",
      "nm_produk": "KREDIT REKENING KORAN",
      "flgjmn": "Y",
      "jnsang": "Bunga",
      "rmsbunga": "EFEKTIF",
      "baserate": 14,
      "bulat": 100
    },
    {
      "kd_produk": "52",
      "nm_produk": "KREDIT ANUITAS",
      "flgjmn": "N",
      "jnsang": "Pokok + Bunga",
      "rmsbunga": "ANUITAS",
      "baserate": 11.5,
      "bulat": 100
    }
  ]
}
```

---

### 6.2 Simulasi Pinjaman / Kalkulator Kredit
Menghitung simulasi kredit secara langsung tanpa menyimpan data ke server. Menghasilkan besaran angsuran per bulan, rincian porsi pokok dan bunga, serta tabel rencana angsuran lengkap dari awal hingga akhir tenor.

- **Method:** `POST`
- **Path:** `/pinjaman/simulasi`

#### Parameter Request (JSON Body)
- `plf_debitur` (Number, Wajib): Nilai pokok / plafon pinjaman yang disimulasikan (harus > 0).
- `jk_wktawal` (Integer, Wajib): Jangka waktu / tenor kredit dalam hitungan bulan.
- `bunga` (Number, Wajib): Suku bunga pinjaman per tahun (persen, misal `12` untuk 12%).
- `rmsbunga` (String, Wajib): Kode metode rumus bunga:
  - `"2"` = ANUITAS
  - `"4"` = FLAT
  - `"5"` = EFEKTIF
- `jnsangsur` (String, Wajib): Kode pola angsuran:
  - `"1"` = Bunga Saja (Pokok dilunasi di akhir tenor)
  - `"2"` = Pokok + Bunga (Angsuran reguler bulanan)
- `tgl_awal` (String, Opsional): Tanggal mulai pembiayaan format `YYYY-MM-DD` (default: hari ini).
- `tglangsur` (Integer, Opsional): Tanggal jatuh tempo cicilan tiap bulan (1 - 31).

#### Contoh Request Body (JSON)
```json
{
  "plf_debitur": 12000000,
  "jk_wktawal": 12,
  "bunga": 12.0,
  "rmsbunga": "4",
  "jnsangsur": "2",
  "tgl_awal": "2026-11-01",
  "tglangsur": 1
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil membuat simulasi pinjaman",
  "data": {
    "plf_debitur": 12000000,
    "jk_wktawal": 12,
    "bunga": 12,
    "rmsbunga": "4",
    "rmsbunga_ket": "FLAT",
    "jnsangsur": "2",
    "jnsangsur_ket": "Pokok + Bunga",
    "tgl_awal": "2026-11-01",
    "jt_tempo": "2027-11-01",
    "angsur_deb": 1120000,
    "angspokok": 1000000,
    "angsbunga": 120000,
    "totbng": 1440000,
    "total_bayar": 13440000,
    "jadwal": [
      {
        "angsurke": 1,
        "tglval": "2026-12-01",
        "pokok_jdw": 1000000,
        "bunga_jdw": 120000,
        "total_angsuran": 1120000,
        "osdeb_jdw": 11000000
      },
      {
        "angsurke": 2,
        "tglval": "2027-01-01",
        "pokok_jdw": 1000000,
        "bunga_jdw": 120000,
        "total_angsuran": 1120000,
        "osdeb_jdw": 10000000
      },
      {
        "angsurke": 12,
        "tglval": "2027-11-01",
        "pokok_jdw": 1000000,
        "bunga_jdw": 120000,
        "total_angsuran": 1120000,
        "osdeb_jdw": 0
      }
    ]
  }
}
```

---

### 6.3 Pinjaman Baru & Generate Jadwal Angsuran
Mendaftarkan fasilitas pinjaman kredit baru untuk nasabah terdaftar sekaligus membuat seluruh jadwal angsuran secara atomik.

- **Method:** `POST`
- **Path:** `/pinjaman/pinjamanbaru`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `cno` (String, Wajib): Nomor anggota peminjam (contoh: `"0000000001"`).
- `jenis` (String, Wajib): Kode produk pinjaman (contoh: `"50"`).
- `plf_debitur` (Number, Wajib): Nilai plafon pinjaman yang disetujui (harus > 0).
- `jk_wktawal` (Integer, Wajib): Tenor dalam hitungan bulan (harus > 0).
- `bunga` (Number, Opsional): Persentase suku bunga tahunan. Jika 0, akan mengambil nilai *baserate* produk terkait.
- `rmsbunga` (String, Opsional): Kode rumus bunga (`"2"`/`"4"`/`"5"`). Jika kosong, mengambil konfigurasi produk.
- `jnsangsur` (String, Opsional): Kode jenis angsuran (`"1"`/`"2"`). Jika kosong, mengambil konfigurasi produk.
- `jnskredit` (String, Opsional): Jenis fasilitas kredit (default: `"1"` = Baru).
- `tgl_awal` (String, Wajib): Tanggal akad/pencairan format `YYYY-MM-DD`.
- `tglangsur` (Integer, Opsional): Tanggal jatuh tempo cicilan per bulan (1 - 31).
- `provisi` (Number, Opsional): Biaya provisi pinjaman.
- `adm` (Number, Opsional): Biaya administrasi pinjaman.
- `noac` (String, Opsional): Nomor rekening tabungan peminjam untuk keperluan autodebet cicilan.
- `kodeao` (String, Opsional): Kode petugas Account Officer (AO).
- `keterangan` (String, Opsional): Catatan atau tujuan penggunaan kredit.
- `nljaminan` (Number, Opsional): Nilai taksasi agunan/jaminan kredit.

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "cno": "0000000001",
  "jenis": "50",
  "plf_debitur": 6000000,
  "jk_wktawal": 6,
  "bunga": 12.0,
  "rmsbunga": "4",
  "jnsangsur": "2",
  "tgl_awal": "2026-11-01",
  "tglangsur": 1,
  "provisi": 60000,
  "adm": 25000,
  "noac": "0011000001",
  "kodeao": "0000",
  "keterangan": "Pembiayaan modal usaha",
  "nljaminan": 10000000
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil membuat fasilitas pinjaman baru dan jadwal angsuran",
  "data": {
    "kdcab": "001",
    "cno": "0000000001",
    "nofas": "0015000001",
    "nama": "OTTO",
    "jenis": "50",
    "nm_produk": "KREDIT MULTIGUNA",
    "plf_debitur": 6000000,
    "jk_wktawal": 6,
    "bunga": 12,
    "rmsbunga": "4",
    "jnsangsur": "2",
    "tgl_awal": "2026-11-01",
    "jt_tempo": "2027-05-01",
    "tglangs": "1",
    "angsur_deb": 1060000,
    "angspokok": 1000000,
    "angsbunga": 60000,
    "totbng": 360000,
    "provisi": 60000,
    "adm": 25000,
    "total_biaya": 85000,
    "jadwal": [
      {
        "angsurke": 1,
        "tglval": "2026-12-01",
        "pokok_jdw": 1000000,
        "bunga_jdw": 60000,
        "total_angsuran": 1060000,
        "osdeb_jdw": 5000000
      },
      {
        "angsurke": 2,
        "tglval": "2027-01-01",
        "pokok_jdw": 1000000,
        "bunga_jdw": 60000,
        "total_angsuran": 1060000,
        "osdeb_jdw": 4000000
      },
      {
        "angsurke": 6,
        "tglval": "2027-05-01",
        "pokok_jdw": 1000000,
        "bunga_jdw": 60000,
        "total_angsuran": 1060000,
        "osdeb_jdw": 0
      }
    ]
  }
}
```

---

### 6.4 Get Rincian Jadwal Angsuran
Mengambil seluruh baris rencana jadwal cicilan pinjaman beserta status pembayarannya (apakah sudah lunas atau belum).

- **Method:** `POST`
- **Path:** `/pinjaman/getjadwalangsuran`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `nofas` (String, Wajib): Nomor fasilitas pinjaman (contoh: `"0015000001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "nofas": "0015000001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan jadwal angsuran",
  "data": [
    {
      "angsurke": 1,
      "tglval": "2026-12-01",
      "pokok_jdw": 1000000,
      "bunga_jdw": 60000,
      "total_angsuran": 1060000,
      "osdeb_jdw": 5000000,
      "flagbayar": "0",
      "status": "BELUM BAYAR"
    },
    {
      "angsurke": 2,
      "tglval": "2027-01-01",
      "pokok_jdw": 1000000,
      "bunga_jdw": 60000,
      "total_angsuran": 1060000,
      "osdeb_jdw": 4000000,
      "flagbayar": "0",
      "status": "BELUM BAYAR"
    }
  ]
}
```

---

### 6.5 Get Daftar Pinjaman Nasabah
Menampilkan daftar seluruh fasilitas pinjaman yang dimiliki oleh anggota tertentu (plafon, sisa pokok pinjaman, kolektibilitas, dan status pinjaman).

- **Method:** `POST`
- **Path:** `/pinjaman/getdaftarpinjaman`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `cno` (String, Wajib): Nomor anggota (contoh: `"0000000001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "cno": "0000000001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan daftar pinjaman nasabah",
  "data": [
    {
      "kdcab": "001",
      "cno": "0000000001",
      "nofas": "0015000001",
      "jenis": "50",
      "nm_produk": "KREDIT MULTIGUNA",
      "plf_debitur": 6000000,
      "sldakhir": 6000000,
      "tgl_awal": "2026-11-01T00:00:00Z",
      "jt_tempo": "2027-05-01T00:00:00Z",
      "angsur_deb": 1060000,
      "kolek": "1",
      "status": "1",
      "status_ket": "AKTIF"
    }
  ]
}
```

---

### 6.6 Get Data Detail Pinjaman
Menampilkan informasi menyeluruh satu fasilitas pinjaman nasabah.

- **Method:** `POST`
- **Path:** `/pinjaman/getdatapinjaman`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `nofas` (String, Wajib): Nomor fasilitas pinjaman (contoh: `"0015000001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "nofas": "0015000001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan detail data pinjaman",
  "data": {
    "kdcab": "001",
    "cno": "0000000001",
    "nama": "OTTO",
    "nofas": "0015000001",
    "jenis": "50",
    "nm_produk": "KREDIT MULTIGUNA",
    "plf_debitur": 6000000,
    "sldakhir": 6000000,
    "jk_wktawal": 6,
    "bunga": 12,
    "rmsbunga": "4",
    "rmsbunga_ket": "FLAT",
    "jnsangsur": "2",
    "jnsangsur_ket": "Pokok + Bunga",
    "tgl_awal": "2026-11-01T00:00:00Z",
    "jt_tempo": "2027-05-01T00:00:00Z",
    "tglangsur": 1,
    "angsur_deb": 1060000,
    "angspokok": 1000000,
    "angsbunga": 60000,
    "totbng": 360000,
    "provisi": 60000,
    "adm": 25000,
    "total_biaya": 85000,
    "kolek": "1",
    "status": "AKTIF",
    "noac": "0011000001",
    "kodeao": "0000"
  }
}
```

---

### 6.7 Get Tagihan Pinjaman & Denda
Melakukan pengecekan (*inquiry*) tagihan angsuran pinjaman yang belum terbayar beserta estimasi denda keterlambatan jika tanggal transaksi telah melampaui tanggal jatuh tempo.

- **Method:** `POST`
- **Path:** `/pinjaman/gettagihan`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `nofas` (String, Wajib): Nomor fasilitas pinjaman (contoh: `"0015000001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "nofas": "0015000001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan tagihan pinjaman",
  "data": {
    "kdcab": "001",
    "nofas": "0015000001",
    "nama": "OTTO",
    "jenis": "50",
    "nm_produk": "KREDIT MULTIGUNA",
    "total_pokok": 6000000,
    "total_bunga": 360000,
    "total_denda": 0,
    "grand_total": 6360000,
    "jumlah_angsuran_tgk": 6,
    "tagihan_detail": [
      {
        "angsurke": 1,
        "tglval": "2026-12-01",
        "pokok": 1000000,
        "bunga": 60000,
        "denda": 0,
        "hari_terlambat": 0,
        "total": 1060000,
        "status": "BELUM LUNAS"
      }
    ]
  }
}
```

---

### 6.8 Pembayaran Angsuran Kredit
Memproses pembayaran cicilan kredit pinjaman. Mendukung pemotongan otomatis dari rekening tabungan (`"R"`), pembayaran tunai di kasir (`"T"`), atau via transfer bank (`"B"`).

- **Method:** `POST`
- **Path:** `/pinjaman/bayarangsuran`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `nofas` (String, Wajib): Nomor fasilitas pinjaman (contoh: `"0015000001"`).
- `angsurke` (Integer, Opsional): Nomor urut angsuran yang dibayar. Jika diisi 0 atau kosong, sistem otomatis memproses cicilan tertua yang belum lunas.
- `crbayar` (String, Opsional): Cara pembayaran:
  - `"R"` = Tabungan / Autodebet (Default)
  - `"T"` = Tunai di teller
  - `"B"` = Bank transfer / VA
- `noac_debet` (String, Opsional): Nomor rekening tabungan peminjam (wajib diisi apabila `crbayar` = `"R"`).
- `nominal` (Number, Opsional): Validasi nominal pembayaran.
- `keterangan` (String, Opsional): Catatan pembayaran.

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "nofas": "0015000001",
  "angsurke": 1,
  "crbayar": "R",
  "noac_debet": "0011000001",
  "keterangan": "Pembayaran Angsuran ke 1"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil memproses pembayaran angsuran",
  "data": {
    "noref": "KRD00120261028000001",
    "kdcab": "001",
    "nofas": "0015000001",
    "angsurke": 1,
    "tglbyr": "2026-10-28 17:00:00",
    "crbayar": "R",
    "noac_debet": "0011000001",
    "pokok": 1000000,
    "bunga": 60000,
    "denda": 0,
    "total_bayar": 1060000,
    "sisa_outstanding": 5000000,
    "status_pinjaman": "AKTIF",
    "keterangan": "Pembayaran Angsuran ke 1"
  }
}
```

#### Contoh Response Gagal (HTTP 400)
```json
{
  "message": "Saldo rekening tidak mencukupi untuk pembayaran angsuran"
}
```

---

### 6.9 Pelunasan Kredit Dipercepat
Melunasi sisa seluruh pokok pinjaman sebelum masa akhir tenor (*early payoff*). Menghitung otomatis akumulasi bunga berjalan, denda, penalti (bila ada), dan menutup pinjaman dengan status `"LUNAS"`.

- **Method:** `POST`
- **Path:** `/pinjaman/pelunasankredit`

#### Parameter Request (JSON Body)
- `kdcab` (String, Wajib): Kode cabang (contoh: `"001"`).
- `nofas` (String, Wajib): Nomor fasilitas pinjaman (contoh: `"0015000001"`).
- `crbayar` (String, Opsional): Cara pembayaran (`"R"` = Tabungan, `"T"` = Tunai, `"B"` = Bank). Default: `"R"`.
- `noac_debet` (String, Opsional): Nomor rekening tabungan peminjam (wajib jika `crbayar` = `"R"`).
- `keterangan` (String, Opsional): Catatan pelunasan.

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "nofas": "0015000001",
  "crbayar": "R",
  "noac_debet": "0011000001",
  "keterangan": "Pelunasan pinjaman dipercepat"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil memproses pelunasan kredit",
  "data": {
    "noref": "LNP00120261028000001",
    "kdcab": "001",
    "nofas": "0015000001",
    "tgl_lunas": "2026-10-28 17:05:00",
    "crbayar": "R",
    "noac_debet": "0011000001",
    "sisa_pokok": 5000000,
    "bunga": 60000,
    "denda": 0,
    "penalty": 0,
    "total_pelunasan": 5060000,
    "status": "LUNAS",
    "keterangan": "Pelunasan pinjaman dipercepat"
  }
}
```

---

### 6.10 Get Riwayat Pembayaran Pinjaman
Mengambil seluruh mutasi pembayaran angsuran berkala maupun pelunasan yang pernah dilakukan pada nomor fasilitas kredit tertentu.

- **Method:** `POST`
- **Path:** `/pinjaman/getriwayatbayar`

#### Parameter Request (JSON Body)
- `kdcab` (String, Opsional): Kode cabang.
- `nofas` (String, Wajib): Nomor fasilitas pinjaman (contoh: `"0015000001"`).

#### Contoh Request Body (JSON)
```json
{
  "kdcab": "001",
  "nofas": "0015000001"
}
```

#### Contoh Response Sukses (HTTP 200)
```json
{
  "message": "Berhasil mendapatkan riwayat pembayaran pinjaman",
  "data": [
    {
      "noref": "LNP00120261028000001",
      "tgltrn": "2026-10-28 17:05:00",
      "ststrn": "02",
      "ststrn_desc": "Pelunasan Kredit",
      "ket": "Pelunasan pinjaman dipercepat",
      "crbayar": "R",
      "crbayar_desc": "Tabungan",
      "noacdb": "0011000001",
      "angsurke": 2,
      "nlpokok": 5000000,
      "nlbunga": 60000,
      "nldenda": 0,
      "nlpinalty": 0,
      "total": 5060000
    },
    {
      "noref": "KRD00120261028000001",
      "tgltrn": "2026-10-28 17:00:00",
      "ststrn": "01",
      "ststrn_desc": "Pembayaran Angsuran",
      "ket": "Pembayaran Angsuran ke 1",
      "crbayar": "R",
      "crbayar_desc": "Tabungan",
      "noacdb": "0011000001",
      "angsurke": 1,
      "nlpokok": 1000000,
      "nlbunga": 60000,
      "nldenda": 0,
      "nlpinalty": 0,
      "total": 1060000
    }
  ]
}
```

---

## 7. Lampiran Referensi Kode & Parameter

### 7.1 Kode Rumus Bunga (`rmsbunga`)
| Kode | Label | Deskripsi |
|:---:|---|---|
| **`2`** | **ANUITAS** | Jumlah total angsuran bulanan tetap. Komposisi angsuran bunga mengecil dan pokok membesar setiap bulannya. |
| **`4`** | **FLAT** | Porsi angsuran pokok dan bunga nilainya sama tetap setiap bulan dihitung langsung dari plafon awal pinjaman. |
| **`5`** | **EFEKTIF** | Porsi pokok bernilai tetap setiap bulan, sementara porsi bunga dihitung berdasarkan sisa pokok pinjaman berjalan (*outstanding balance*). |

### 7.2 Kode Jenis Angsuran (`jnsangsur`)
| Kode | Label | Deskripsi |
|:---:|---|---|
| **`1`** | **Bunga Saja** | Nasabah hanya membayar bunga setiap bulan, pokok pinjaman dilunasi sekaligus di akhir periode jatuh tempo. |
| **`2`** | **Pokok + Bunga** | Nasabah membayar cicilan pokok dan cicilan bunga secara berkala setiap bulan. |

### 7.3 Kode Cara Bayar (`crbayar`)
| Kode | Label | Deskripsi |
|:---:|---|---|
| **`R`** | **Tabungan** | Pemotongan autodebet langsung dari rekening tabungan nasabah di koperasi. |
| **`T`** | **Tunai** | Penyetoran tunai fisik melalui loket teller cabang. |
| **`B`** | **Bank** | Penyetoran melalui transfer bank perantara atau Virtual Account. |

### 7.4 Kode Status Transaksi Pinjaman (`ststrn`)
| Kode | Label | Deskripsi |
|:---:|---|---|
| **`01`** | **Pembayaran Angsuran** | Transaksi pembayaran cicilan kredit berkala. |
| **`02`** | **Pelunasan Kredit** | Transaksi pelunasan sisa pokok pinjaman sebelum atau pada jatuh tempo. |
