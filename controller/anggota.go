package controller

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tobias-Kusuma/core-ksp-api/auth"
	"github.com/Tobias-Kusuma/core-ksp-api/cache"
	"github.com/Tobias-Kusuma/core-ksp-api/helper"
	"github.com/Tobias-Kusuma/core-ksp-api/obj"
	"github.com/gin-gonic/gin"
)

// login untuk akun anggota
func LoginAnggota() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqLoginAnggota
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := ReqBody.KDCAB
		Cno := ReqBody.CNO
		Nohp := helper.NormalizePhoneNumberString(ReqBody.NOHP)

		// ambil data terkait Anggota
		var Anggota obj.MSTCST
		if err := AppTx.Main.Table("MSTCST").Select("kdcab, cno, cnm, pin").Where("kdcab = ? and cno = ? and nohp = ?", Kdcab, Cno, Nohp).Take(&Anggota).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// verifikasi pin
		if ok := helper.VerifyPassword(ReqBody.PIN, *Anggota.PIN); ok == false {
			err := fmt.Errorf("no. anggota, no. hp, atau PIN yang dimasukkan tidak sesuai!")
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, true, err)
			return
		}

		// kalau bener, kasih response berupa token
		// 0. siapkan session id
		newSessionId := helper.GetUUIDString()

		// 1. token akses
		Acc_token, err := auth.GenerateAccessToken(Anggota.KDCAB, Anggota.CNO, *Anggota.CNM, newSessionId)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// 2. token refresh
		Ref_token, err := auth.GenerateRefreshToken(Anggota.KDCAB, Anggota.CNO, *Anggota.CNM, newSessionId)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// sesi hanya dibuat saat login, bukan saat refresh token
		keySessionId := auth.PrefixSessionKey + newSessionId
		if err := cache.RedisSet(keySessionId, keySessionId, 1*21*time.Minute); err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Gagal membuat session"))
			return
		}

		ResData := obj.ResLoginAnggota{
			ACCESS_TOKEN:  Acc_token,
			REFRESH_TOKEN: Ref_token,
		}

		helper.RespondMaskedOk(c, "Berhasil melakukan login", ResData)
	}
}

// refresh token untuk session anggota
func RefreshSessionAnggota() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqRefreshSessionAnggota
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		var Anggota obj.MSTCST
		if err := AppTx.Main.Table("MSTCST").Select("kdcab, cno, cnm").Where("kdcab = ? and cno = ?", ReqBody.KDCAB, ReqBody.CNO).Take(&Anggota).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// validasi keabsahan token eak
		claims, err := auth.ValidateToken(ReqBody.REFRESH_TOKEN)
		if err != nil {
			log.Printf("ERROR: Token tidak valid! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("token tidak valid!"))
			return
		}

		if claims.TOKENTYPE != "refresh" {
			log.Printf("ERROR: %s", "Tipe token tidak valid! Bukan token bertipe REFRESH")
			helper.RespondMaskedError(c, true, fmt.Errorf("tipe token tidak valid!"))
			return
		}

		// cocokkan data dari claims, dengan hasil search, dan request
		if claims.KDCAB != Anggota.KDCAB || claims.CNO != Anggota.CNO || claims.CNM != *Anggota.CNM {
			err := fmt.Errorf("Invalid, identitas tidak sesuai dengan token claims!")
			log.Printf("ERROR: Token tidak valid! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("token tidak valid!"))
			return
		}

		// token akses dan refresh harus berasal dari sesi dan anggota yang sama
		DetailCust, err := helper.GetDetailCust(c)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		if DetailCust.SESSIONID != claims.SESSIONID || DetailCust.KDCAB != claims.KDCAB || DetailCust.CNO != claims.CNO {
			log.Printf("ERROR: Session token akses dan refresh tidak sesuai")
			helper.RespondMaskedError(c, true, fmt.Errorf("session tidak valid!"))
			return
		}

		// cek apakah session masih ada di redis
		keySessionId := auth.PrefixSessionKey + claims.SESSIONID
		_, isFound, err := cache.RedisGet[string](keySessionId)

		if err != nil {
			log.Printf("ERROR: Session tidak valid! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("session tidak valid!"))
			return
		}

		if isFound == false {
			log.Printf("ERROR: Session tidak valid!")
			helper.RespondMaskedError(c, true, fmt.Errorf("session tidak valid!"))
			return
		}

		// buatkan access_token yang baru
		Acc_token, err := auth.GenerateAccessToken(Anggota.KDCAB, Anggota.CNO, *Anggota.CNM, claims.SESSIONID)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		ResData := obj.ResRefreshSessionAnggota{
			ACCESS_TOKEN: Acc_token,
		}

		helper.RespondMaskedOk(c, "Berhasil melakukan refresh session", ResData)
	}
}

// logout session untuk anggota
func LogoutAnggota() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqLogoutAnggota
		_, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		DetailCust, err := helper.GetDetailCust(c)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		if DetailCust.KDCAB != ReqBody.KDCAB || DetailCust.CNO != ReqBody.CNO {
			err := fmt.Errorf("Identitas tidak valid")
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, true, err)
			return
		}

		// ambil session id
		SessionId := DetailCust.SESSIONID

		// hapus session id dari cache
		keySessionId := auth.PrefixSessionKey + SessionId
		if err := cache.RedisDel(keySessionId); err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		helper.RespondMaskedOk(c, "Berhasil melakukan logout", nil)
	}
}

// binding atau buka akun mobile untuk anggota koperasi terdaftar
func CreateAkunMobileAnggota() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqAkunMobileBaru
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// samakan format nama, email, notelp, dan nik
		Cno := strings.TrimSpace(ReqBody.CNO)
		Nama := strings.ToUpper(strings.Join(strings.Fields(ReqBody.CNM), " "))
		Email := ReqBody.EMAIL
		NoTelp := helper.NormalizePhoneNumberString(ReqBody.NOHP)
		Nik := ReqBody.NOID

		// cek apakah ada nasabah terdaftar
		type ReturnNasabah struct {
			CNO   string `gorm:"column:cno"`
			CNM   string `gorm:"column:cnm"`
			EMAIL string `gorm:"column:email"`
			NOHP  string `gorm:"column:nohp"`
			NOID  string `gorm:"column:noid"`
		}
		var ReturnNas ReturnNasabah
		if err := AppTx.Main.Table("MSTCST").Select("cno, dbo.fn_NormalisasiNama(cnm) as cnm, email, nohp, noid").Where("cno = ?", Cno).Where("email = ?", Email).
			Where("dbo.fn_NormalisasiNama(cnm) = ?", Nama).Where("nohp = ?", NoTelp).Where("noid = ?", Nik).Take(&ReturnNas).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// siapkan PIN yang sudah dihash
		HashedPin, err := helper.GenerateHashedPass(ReqBody.PIN)
		if err != nil {
			log.Printf("ERROR: Gagal menyimpan PIN! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Gagal menyimpan PIN!"))
			return
		}

		// update PIN dan nama pasar ke table
		DataUpdate := map[string]any{
			"pasar": ReqBody.PASAR,
			"pin":   HashedPin,
		}
		if err := AppTx.Main.Table("MSTCST").Where("kdcab = ? and cno = ?", ReqBody.KDCAB, Cno).Updates(&DataUpdate).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		helper.RespondMaskedOk(c, "Berhasil membuat akun mobile", nil)
	}
}

// update foto ktp
func UploadFotoKtp() gin.HandlerFunc {
	return func(c *gin.Context) {
		AppTx, err := helper.Get_DB_ReqBody(c, 30, nil)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// ambil data dari token
		DataCust, err := helper.GetDetailCust(c)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// maksimal ukuran gambar 4 MiB (4 * 1024 * 1024 bytes)
		const maxImageSize int64 = 4_000_000

		// maksimal resolusi 20 megapixel
		const maxImagePixels int64 = 20_000_000

		// 1. Batasi seluruh ukuran HTTP request
		c.Request.Body = http.MaxBytesReader(
			c.Writer,
			c.Request.Body,
			maxImageSize+(1<<20), // 4 MiB + 1 MiB overhead multipart
		)

		// 2. Ambil file dari form-data
		fileHeader, err := c.FormFile("file")
		if err != nil {
			log.Printf("ERROR: File gambar tidak ditemukan atau request terlalu besar. %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("file gambar tidak ditemukan atau request terlalu besar"))
			return
		}

		// 3. Validasi ukuran file
		if fileHeader.Size == 0 || fileHeader.Size > maxImageSize {
			err := fmt.Errorf("ukuran gambar harus antara 1 byte dan 4 MB")
			helper.RespondMaskedError(c, true, err)
			return
		}

		// 4. Validasi ekstensi file
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

		if ext != ".jpg" {
			err := fmt.Errorf("file harus berekstensi .jpg")
			helper.RespondMaskedError(c, true, err)
			return
		}

		// 5. Buka file upload
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("ERROR: Gagal membuka file gambar. %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("gagal membuka file gambar"))
			return
		}
		defer file.Close()

		// 6. Baca file menjadi []byte
		fileBytes, err := io.ReadAll(io.LimitReader(file, maxImageSize+1))
		if err != nil {
			log.Printf("ERROR: Gagal membaca file gambar. %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("gagal membaca file gambar"))
			return
		}

		if len(fileBytes) == 0 || int64(len(fileBytes)) > maxImageSize {
			err := fmt.Errorf("ukuran gambar tidak valid")
			helper.RespondMaskedError(c, true, err)
			return
		}

		// 7. Validasi MIME berdasarkan isi file
		mimeType := http.DetectContentType(fileBytes)

		if mimeType != "image/jpeg" {
			err := fmt.Errorf("file harus berformat JPG")
			helper.RespondMaskedError(c, true, err)
			return
		}

		// 8. Validasi header gambar dan dimensi
		config, format, err := image.DecodeConfig(bytes.NewReader(fileBytes))
		if err != nil {
			log.Printf("ERROR: Gagal membaca konfigurasi gambar. %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("file gambar tidak valid"))
			return
		}

		if format != "jpeg" || config.Width <= 0 || config.Height <= 0 {
			err := fmt.Errorf("format atau dimensi gambar tidak valid")
			helper.RespondMaskedError(c, true, err)
			return
		}

		// Batasi total pixel untuk mencegah penggunaan memori berlebihan
		totalPixels := int64(config.Width) * int64(config.Height)

		if totalPixels > maxImagePixels {
			err := fmt.Errorf("resolusi gambar terlalu besar (maksimal 20 megapixel)")
			helper.RespondMaskedError(c, true, err)
			return
		}

		// 9. Decode seluruh gambar untuk validasi
		_, decodedFormat, err := image.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			log.Printf("ERROR: Gambar rusak atau tidak dapat di-decode. %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("file JPG rusak atau tidak valid"))
			return
		}

		if decodedFormat != "jpeg" {
			err := fmt.Errorf("format gambar tidak sesuai")
			helper.RespondMaskedError(c, true, err)
			return
		}

		// 10. Gambar valid dan siap disimpan
		log.Printf("Upload JPG valid: ukuran=%d bytes, resolusi=%dx%d", len(fileBytes), config.Width, config.Height)

		// fileBytes merupakan []byte asli dari gambar JPG

		// begin transaction
		trans := AppTx.Main.Begin()

		// simpan ke db
		DataFoto := map[string]any{
			"Nasabahpic": fileBytes,
		}
		if err := AppTx.Main.Table("MSTCST").Where("kdcab = ? and cno = ?", DataCust.KDCAB, DataCust.CNO).Updates(&DataFoto).Error; err != nil {
			trans.Rollback()
			log.Printf("ERROR: Gagal menyimpan foto. %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		trans.Commit()
		c.JSON(http.StatusOK, gin.H{"kode": "000", "message": "Berhasil simpan foto KTP"})
	}
}

// ambil detail data anggota
func GetDataAnggota() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqDataAnggota
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// ambil data nasabah dari token
		DataCust, err := helper.GetDetailCust(c)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		if DataCust.CNO != ReqBody.CNO && DataCust.KDCAB != ReqBody.KDCAB {
			err := fmt.Errorf("Identitas tidak dapat dipastikan")
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, true, err)
			return
		}

		var Anggota obj.MSTCST
		if err := AppTx.Main.Debug().Table("MSTCST").Select("*").Where("kdcab = ? and cno = ?", ReqBody.KDCAB, ReqBody.CNO).Scan(&Anggota).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		DaftarRekening := make([]obj.SHDBLC, 0)
		if err := AppTx.Main.Debug().Table("SHDBLC").Select("*").Where("cno = ?", ReqBody.CNO).Scan(&DaftarRekening).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// hapus data pin dari Anggota
		Anggota.PIN = nil

		ResData := obj.ResDataAnggota{
			ANGGOTA:         Anggota,
			DAFTAR_REKENING: DaftarRekening,
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan daftar anggota", ResData)
	}
}

/* NOT USED : API LANGSUNG DICONSUME OLEH END-USER */

// // ambil list daftar anggota koperasi
// func GetDaftarAnggota() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var ReqBody obj.ReqDaftarAnggota
// 		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
// 		if err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		ResData := make([]obj.ResDaftarAnggota, 0)
// 		if err := AppTx.Main.Debug().Table("MSTCST").Select("kdcab, cno, tipenas, jnsnas, cnm").Where("kdcab = ?", ReqBody.KDCAB).Scan(&ResData).Error; err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		helper.RespondMaskedOk(c, "Berhasil mendapatkan daftar anggota", ResData)
// 	}
// }

// // registrasi anggota baru
// func PendaftaranAnggotaBaru() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var ReqBody obj.ReqAnggotaBaru
// 		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
// 		if err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		// ambil data fixed
// 		Kdcab := strings.TrimSpace(ReqBody.ANGGOTA.KDCAB)
// 		Nama := ReqBody.ANGGOTA.CNM
// 		JenisPok := strings.TrimSpace(ReqBody.SIMPANAN_POKOK.JENIS)
// 		JenisWaj := strings.TrimSpace(ReqBody.SIMPANAN_WAJIB.JENIS)
// 		JenisSuk := strings.TrimSpace(ReqBody.SIMPANAN_SUKARELA.JENIS)

// 		// buat cno
// 		// ambil no urut terakhir
// 		var last_cno int64
// 		if err := AppTx.Main.Debug().Table("MSTCST").Select("isnull(max(try_cast(cno as bigint)), 0) as last_cno").Scan(&last_cno).Error; err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		Cno := fmt.Sprintf("%010d", last_cno+1) // no.urut ditambahkan padding 0 (memastikan panjang karakter = 10)

// 		// buat noac; kdcab (3) + jenis (2) + no.urut dipadding 0 minimal 5 karakter %05d
// 		// simpanan pokok
// 		var last_noac_pok int64
// 		if err := AppTx.Main.Debug().Table("SHDBLC").Select("isnull(max(try_cast(right(noac, 5) as bigint)), 0) as last_noac").
// 			Where("kdcab = ? and jenis = ?", Kdcab, JenisPok).Scan(&last_noac_pok).Error; err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		NoacPok := fmt.Sprintf("%s%s%05d", Kdcab, JenisPok, last_noac_pok+1)

// 		// simpanan wajib
// 		var last_noac_waj int64
// 		if err := AppTx.Main.Debug().Table("SHDBLC").Select("isnull(max(try_cast(right(noac, 5) as bigint)), 0) as last_noac").
// 			Where("kdcab = ? and jenis = ?", Kdcab, JenisWaj).Scan(&last_noac_waj).Error; err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		NoacWaj := fmt.Sprintf("%s%s%05d", Kdcab, JenisWaj, last_noac_waj+1)

// 		// simpanan sukarela
// 		var last_noac_suk int64
// 		if err := AppTx.Main.Debug().Table("SHDBLC").Select("isnull(max(try_cast(right(noac, 5) as bigint)), 0) as last_noac").
// 			Where("kdcab = ? and jenis = ?", Kdcab, JenisSuk).Scan(&last_noac_suk).Error; err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		NoacSuk := fmt.Sprintf("%s%s%05d", Kdcab, JenisSuk, last_noac_suk+1)

// 		// ambil data biaya penutupan tabungan sesuai jenis
// 		ByTutupPok := 0.00
// 		if err := AppTx.Main.Debug().Table("RET500").Select("biayatutup").Where("jenis = ?", JenisPok).Scan(&ByTutupPok).Error; err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		ByTutupWaj := 0.00
// 		if err := AppTx.Main.Debug().Table("RET500").Select("biayatutup").Where("jenis = ?", JenisWaj).Scan(&ByTutupWaj).Error; err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		ByTutupSuk := 0.00
// 		if err := AppTx.Main.Debug().Table("RET500").Select("biayatutup").Where("jenis = ?", JenisSuk).Scan(&ByTutupSuk).Error; err != nil {
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		// validasi beberapa data
// 		// tgl lahir
// 		TglLahir, err := time.Parse(helper.DateFormat, ReqBody.ANGGOTA.TGLLHR)
// 		if err != nil {
// 			log.Printf("ERROR: Tanggal Lahir tidak valid! %s", err.Error())
// 			helper.RespondMaskedError(c, true, fmt.Errorf("Tanggal Lahir tidak valid!"))
// 		}

// 		// umur
// 		UmurInt := helper.HitUmur(TglLahir, time.Now())
// 		Umur := float64(UmurInt)

// 		// alamat email
// 		validMail := helper.IsValidEmail(ReqBody.ANGGOTA.EMAIL)
// 		if !validMail {
// 			log.Printf("ERROR: Alamat Email tidak valid! %s", err.Error())
// 			helper.RespondMaskedError(c, true, fmt.Errorf("Alamat Email tidak valid!"))
// 		}

// 		// nohp
// 		NoHp := helper.NormalizePhoneNumberString(ReqBody.ANGGOTA.NOHP)

// 		// tanggal hari ini
// 		TglNow := time.Now()

// 		// user input: mobile
// 		UserInput := "Via API"

// 		// saldo Rp0
// 		NolRupiah := 0.00

// 		// buat pin dihashed
// 		HashedPin, err := helper.GenerateHashedPass(ReqBody.ANGGOTA.PIN)
// 		if err != nil {
// 			log.Printf("ERROR: PIN tidak dalam format yang valid! %s", err.Error())
// 			helper.RespondMaskedError(c, true, fmt.Errorf("PIN tidak dalam format yang valid!"))
// 		}

// 		// masukkan ke dalam table
// 		trx := AppTx.Main.Begin()

// 		// 1. data anggota
// 		DataAnggota := obj.MSTCST{
// 			KDCAB:     Kdcab,
// 			CNO:       Cno,
// 			TIPENAS:   &ReqBody.ANGGOTA.TIPENAS,
// 			JNSNAS:    &ReqBody.ANGGOTA.JNSNAS,
// 			CNM:       &ReqBody.ANGGOTA.CNM,
// 			TGLLHR:    &TglLahir,
// 			UMUR:      &Umur,
// 			TMPLHR:    &ReqBody.ANGGOTA.TMPLHR,
// 			JNSKEL:    &ReqBody.ANGGOTA.JNSKEL,
// 			JNSID:     &ReqBody.ANGGOTA.JNSID,
// 			MARSTAT:   &ReqBody.ANGGOTA.MARSTAT,
// 			WARGA:     helper.StringPtr("1"),
// 			NOID:      &ReqBody.ANGGOTA.NOID,
// 			NPWP:      &ReqBody.ANGGOTA.NPWP,
// 			ALAMATID:  &ReqBody.ANGGOTA.ALAMATID,
// 			KELID:     &ReqBody.ANGGOTA.KELID,
// 			KECID:     &ReqBody.ANGGOTA.KECID,
// 			KOTAID:    &ReqBody.ANGGOTA.KOTAID,
// 			KDPOSID:   &ReqBody.ANGGOTA.KDPOSID,
// 			ALAMATSKR: &ReqBody.ANGGOTA.ALAMATSKR,
// 			KELSKR:    &ReqBody.ANGGOTA.KELSKR,
// 			KECSKR:    &ReqBody.ANGGOTA.KECSKR,
// 			KOTAISKR:  &ReqBody.ANGGOTA.KOTAISKR,
// 			KDPOSSKR:  &ReqBody.ANGGOTA.KDPOSSKR,
// 			STSRUMAH:  &ReqBody.ANGGOTA.STSRUMAH,
// 			NOHP:      &NoHp,
// 			EMAIL:     &ReqBody.ANGGOTA.EMAIL,
// 			NMIBU:     &ReqBody.ANGGOTA.NMIBU,
// 			STSPOKOK:  helper.StringPtr("Y"),
// 			STSWAJIB:  helper.StringPtr("Y"),
// 			STATUS:    helper.StringPtr("1"),
// 			TGLINP:    &TglNow,
// 			USRID:     &UserInput,
// 			DUPD:      &TglNow,
// 			OTR:       &UserInput,
// 			DOTR:      &TglNow,
// 			PIN:       &HashedPin,
// 		}
// 		if err := trx.Debug().Table("MSTCST").Create(&DataAnggota).Error; err != nil {
// 			trx.Rollback()
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		// 2. data simpanan pokok
// 		DataSimpPokok := obj.SHDBLC{
// 			KDCAB:    Kdcab,
// 			NOAC:     NoacPok,
// 			CNO:      Cno,
// 			ALIASNM:  &Nama,
// 			JENIS:    &JenisPok,
// 			NLHOLD:   &NolRupiah,
// 			SLDAWAL:  &NolRupiah,
// 			SLDAKHIR: &NolRupiah,
// 			SLDPAKAI: &NolRupiah,
// 			MINSLD:   &NolRupiah,
// 			NLBIAYA:  &ByTutupPok,
// 			MUTDB:    &NolRupiah,
// 			MUTCR:    &NolRupiah,
// 			NLSETOR:  &NolRupiah,
// 			TGLAWAL:  &TglNow,
// 			STATUS:   helper.StringPtr("1"),
// 		}
// 		if err := trx.Debug().Table("SHDBLC").Create(&DataSimpPokok).Error; err != nil {
// 			trx.Rollback()
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		// 3. data simpanan wajib
// 		DataSimpWajib := obj.SHDBLC{
// 			KDCAB:    Kdcab,
// 			NOAC:     NoacWaj,
// 			CNO:      Cno,
// 			ALIASNM:  &Nama,
// 			JENIS:    &JenisWaj,
// 			NLHOLD:   &NolRupiah,
// 			SLDAWAL:  &NolRupiah,
// 			SLDAKHIR: &NolRupiah,
// 			SLDPAKAI: &NolRupiah,
// 			MINSLD:   &NolRupiah,
// 			NLBIAYA:  &ByTutupWaj,
// 			MUTDB:    &NolRupiah,
// 			MUTCR:    &NolRupiah,
// 			NLSETOR:  &ReqBody.SIMPANAN_WAJIB.NLSETOR,
// 			TGLAWAL:  &TglNow,
// 			STATUS:   helper.StringPtr("1"),
// 		}
// 		if err := trx.Debug().Table("SHDBLC").Create(DataSimpWajib).Error; err != nil {
// 			trx.Rollback()
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		// 3. data simpanan wajib
// 		DataSimpSuk := obj.SHDBLC{
// 			KDCAB:    Kdcab,
// 			NOAC:     NoacSuk,
// 			CNO:      Cno,
// 			ALIASNM:  &Nama,
// 			JENIS:    &JenisSuk,
// 			NLHOLD:   &NolRupiah,
// 			SLDAWAL:  &NolRupiah,
// 			SLDAKHIR: &NolRupiah,
// 			SLDPAKAI: &NolRupiah,
// 			MINSLD:   &NolRupiah,
// 			NLBIAYA:  &ByTutupSuk,
// 			MUTDB:    &NolRupiah,
// 			MUTCR:    &NolRupiah,
// 			NLSETOR:  &NolRupiah,
// 			TGLAWAL:  &TglNow,
// 			STATUS:   helper.StringPtr("1"),
// 		}
// 		if err := trx.Debug().Table("SHDBLC").Create(DataSimpSuk).Error; err != nil {
// 			trx.Rollback()
// 			log.Printf("ERROR: %s", err.Error())
// 			helper.RespondMaskedError(c, false, err)
// 		}

// 		trx.Commit()
// 		helper.RespondMaskedOk(c, "Berhasil membuat anggota baru", nil)
// 	}
// }
