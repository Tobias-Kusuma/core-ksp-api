package controller

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Tobias-Kusuma/core-ksp-api/auth"
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
		}

		var Anggota obj.MSTCST
		if err := AppTx.Main.Table("MSTCST").Select("kdcab, cno, cnm, pin").Where("kdcab = ? and cno = ?", ReqBody.KDCAB, ReqBody.CNO).Take(&Anggota).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// verifikasi pin
		if ok := helper.VerifyPassword(ReqBody.PIN, *Anggota.PIN); ok == false {
			err := fmt.Errorf("No. Nasabah atau PIN yang dimasukkan tidak sesuai!")
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// kalau bener, kasih response berupa token
		// 1. token akses
		Acc_token, err := auth.GenerateAccessToken(Anggota.KDCAB, Anggota.CNO, *Anggota.CNM)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// 2. token refresh
		Ref_token, err := auth.GenerateRefreshToken(Anggota.KDCAB, Anggota.CNO, *Anggota.CNM)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
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
		}

		var Anggota obj.MSTCST
		if err := AppTx.Main.Table("MSTCST").Select("kdcab, cno, cnm").Where("kdcab = ? and cno = ?", ReqBody.KDCAB, ReqBody.CNO).Take(&Anggota).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// validasi keabsahan token eak
		claims, err := auth.ValidateToken(ReqBody.REFRESH_TOKEN)
		if err != nil {
			log.Printf("ERROR: Token tidak valid! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Token tidak valid!"))
			return
		}

		if claims.TOKENTYPE != "refresh" {
			log.Printf("ERROR: %s", "Tipe token tidak valid! Bukan token bertipe REFRESH")
			helper.RespondMaskedError(c, true, fmt.Errorf("Tipe token tidak valid!"))
			return
		}

		// cocokkan data dari claims, dengan hasil search, dan request
		if claims.KDCAB != Anggota.KDCAB || claims.CNO != Anggota.CNO || claims.CNM != *Anggota.CNM {
			err := fmt.Errorf("Invalid, identitas tidak sesuai dengan token claims!")
			log.Printf("ERROR: Token tidak valid! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Token tidak valid!"))
			return
		}

		// buatkan access_token yang baru
		Acc_token, err := auth.GenerateAccessToken(Anggota.KDCAB, Anggota.CNO, *Anggota.CNM)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		ResData := obj.ResRefreshSessionAnggota{
			ACCESS_TOKEN: Acc_token,
		}

		helper.RespondMaskedOk(c, "Berhasil melakukan refresh session", ResData)
	}
}

// registrasi anggota
func DaftarAnggotaBaru() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqAnggotaBaru
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// ambil data fixed
		Kdcab := strings.TrimSpace(ReqBody.ANGGOTA.KDCAB)
		Nama := ReqBody.ANGGOTA.CNM
		JenisPok := strings.TrimSpace(ReqBody.SIMPANAN_POKOK.JENIS)
		JenisWaj := strings.TrimSpace(ReqBody.SIMPANAN_WAJIB.JENIS)
		JenisSuk := strings.TrimSpace(ReqBody.SIMPANAN_SUKARELA.JENIS)

		// buat cno
		// ambil no urut terakhir
		var last_cno int64
		if err := AppTx.Main.Debug().Table("MSTCST").Select("isnull(max(try_cast(cno as bigint)), 0) as last_cno").Scan(&last_cno).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		Cno := fmt.Sprintf("%010d", last_cno+1) // no.urut ditambahkan padding 0 (memastikan panjang karakter = 10)

		// buat noac; kdcab (3) + jenis (2) + no.urut dipadding 0 minimal 5 karakter %05d
		// simpanan pokok
		var last_noac_pok int64
		if err := AppTx.Main.Debug().Table("SHDBLC").Select("isnull(max(try_cast(right(noac, 5) as bigint)), 0) as last_noac").
			Where("kdcab = ? and jenis = ?", Kdcab, JenisPok).Scan(&last_noac_pok).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		NoacPok := fmt.Sprintf("%s%s%05d", Kdcab, JenisPok, last_noac_pok+1)

		// simpanan wajib
		var last_noac_waj int64
		if err := AppTx.Main.Debug().Table("SHDBLC").Select("isnull(max(try_cast(right(noac, 5) as bigint)), 0) as last_noac").
			Where("kdcab = ? and jenis = ?", Kdcab, JenisWaj).Scan(&last_noac_waj).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		NoacWaj := fmt.Sprintf("%s%s%05d", Kdcab, JenisWaj, last_noac_waj+1)

		// simpanan sukarela
		var last_noac_suk int64
		if err := AppTx.Main.Debug().Table("SHDBLC").Select("isnull(max(try_cast(right(noac, 5) as bigint)), 0) as last_noac").
			Where("kdcab = ? and jenis = ?", Kdcab, JenisSuk).Scan(&last_noac_suk).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		NoacSuk := fmt.Sprintf("%s%s%05d", Kdcab, JenisSuk, last_noac_suk+1)

		// ambil data biaya penutupan tabungan sesuai jenis
		ByTutupPok := 0.00
		if err := AppTx.Main.Debug().Table("RET500").Select("biayatutup").Where("jenis = ?", JenisPok).Scan(&ByTutupPok).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		ByTutupWaj := 0.00
		if err := AppTx.Main.Debug().Table("RET500").Select("biayatutup").Where("jenis = ?", JenisWaj).Scan(&ByTutupWaj).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		ByTutupSuk := 0.00
		if err := AppTx.Main.Debug().Table("RET500").Select("biayatutup").Where("jenis = ?", JenisSuk).Scan(&ByTutupSuk).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// validasi beberapa data
		// tgl lahir
		TglLahir, err := time.Parse(helper.DateFormat, ReqBody.ANGGOTA.TGLLHR)
		if err != nil {
			log.Printf("ERROR: Tanggal Lahir tidak valid! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Tanggal Lahir tidak valid!"))
		}

		// umur
		UmurInt := helper.HitUmur(TglLahir, time.Now())
		Umur := float64(UmurInt)

		// alamat email
		validMail := helper.IsValidEmail(ReqBody.ANGGOTA.EMAIL)
		if !validMail {
			log.Printf("ERROR: Alamat Email tidak valid! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Alamat Email tidak valid!"))
		}

		// nohp
		NoHp := helper.NormalizePhoneNumberString(ReqBody.ANGGOTA.NOHP)

		// tanggal hari ini
		TglNow := time.Now()

		// user input: mobile
		UserInput := "Via API"

		// saldo Rp0
		NolRupiah := 0.00

		// buat pin dihashed
		HashedPin, err := helper.GenerateHashedPass(ReqBody.ANGGOTA.PIN)
		if err != nil {
			log.Printf("ERROR: PIN tidak dalam format yang valid! %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("PIN tidak dalam format yang valid!"))
		}

		// masukkan ke dalam table
		trx := AppTx.Main.Begin()

		// 1. data anggota
		DataAnggota := obj.MSTCST{
			KDCAB:     Kdcab,
			CNO:       Cno,
			TIPENAS:   &ReqBody.ANGGOTA.TIPENAS,
			JNSNAS:    &ReqBody.ANGGOTA.JNSNAS,
			CNM:       &ReqBody.ANGGOTA.CNM,
			TGLLHR:    &TglLahir,
			UMUR:      &Umur,
			TMPLHR:    &ReqBody.ANGGOTA.TMPLHR,
			JNSKEL:    &ReqBody.ANGGOTA.JNSKEL,
			JNSID:     &ReqBody.ANGGOTA.JNSID,
			MARSTAT:   &ReqBody.ANGGOTA.MARSTAT,
			WARGA:     helper.StringPtr("1"),
			NOID:      &ReqBody.ANGGOTA.NOID,
			NPWP:      &ReqBody.ANGGOTA.NPWP,
			ALAMATID:  &ReqBody.ANGGOTA.ALAMATID,
			KELID:     &ReqBody.ANGGOTA.KELID,
			KECID:     &ReqBody.ANGGOTA.KECID,
			KOTAID:    &ReqBody.ANGGOTA.KOTAID,
			KDPOSID:   &ReqBody.ANGGOTA.KDPOSID,
			ALAMATSKR: &ReqBody.ANGGOTA.ALAMATSKR,
			KELSKR:    &ReqBody.ANGGOTA.KELSKR,
			KECSKR:    &ReqBody.ANGGOTA.KECSKR,
			KOTAISKR:  &ReqBody.ANGGOTA.KOTAISKR,
			KDPOSSKR:  &ReqBody.ANGGOTA.KDPOSSKR,
			STSRUMAH:  &ReqBody.ANGGOTA.STSRUMAH,
			NOHP:      &NoHp,
			EMAIL:     &ReqBody.ANGGOTA.EMAIL,
			NMIBU:     &ReqBody.ANGGOTA.NMIBU,
			STSPOKOK:  helper.StringPtr("Y"),
			STSWAJIB:  helper.StringPtr("Y"),
			STATUS:    helper.StringPtr("1"),
			TGLINP:    &TglNow,
			USRID:     &UserInput,
			DUPD:      &TglNow,
			OTR:       &UserInput,
			DOTR:      &TglNow,
			PIN:       &HashedPin,
		}
		if err := trx.Debug().Table("MSTCST").Create(&DataAnggota).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// 2. data simpanan pokok
		DataSimpPokok := obj.SHDBLC{
			KDCAB:    Kdcab,
			NOAC:     NoacPok,
			CNO:      Cno,
			ALIASNM:  &Nama,
			JENIS:    &JenisPok,
			NLHOLD:   &NolRupiah,
			SLDAWAL:  &NolRupiah,
			SLDAKHIR: &NolRupiah,
			SLDPAKAI: &NolRupiah,
			MINSLD:   &NolRupiah,
			NLBIAYA:  &ByTutupPok,
			MUTDB:    &NolRupiah,
			MUTCR:    &NolRupiah,
			NLSETOR:  &NolRupiah,
			TGLAWAL:  &TglNow,
			STATUS:   helper.StringPtr("1"),
		}
		if err := trx.Debug().Table("SHDBLC").Create(&DataSimpPokok).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// 3. data simpanan wajib
		DataSimpWajib := obj.SHDBLC{
			KDCAB:    Kdcab,
			NOAC:     NoacWaj,
			CNO:      Cno,
			ALIASNM:  &Nama,
			JENIS:    &JenisWaj,
			NLHOLD:   &NolRupiah,
			SLDAWAL:  &NolRupiah,
			SLDAKHIR: &NolRupiah,
			SLDPAKAI: &NolRupiah,
			MINSLD:   &NolRupiah,
			NLBIAYA:  &ByTutupWaj,
			MUTDB:    &NolRupiah,
			MUTCR:    &NolRupiah,
			NLSETOR:  &ReqBody.SIMPANAN_WAJIB.NLSETOR,
			TGLAWAL:  &TglNow,
			STATUS:   helper.StringPtr("1"),
		}
		if err := trx.Debug().Table("SHDBLC").Create(DataSimpWajib).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		// 3. data simpanan wajib
		DataSimpSuk := obj.SHDBLC{
			KDCAB:    Kdcab,
			NOAC:     NoacSuk,
			CNO:      Cno,
			ALIASNM:  &Nama,
			JENIS:    &JenisSuk,
			NLHOLD:   &NolRupiah,
			SLDAWAL:  &NolRupiah,
			SLDAKHIR: &NolRupiah,
			SLDPAKAI: &NolRupiah,
			MINSLD:   &NolRupiah,
			NLBIAYA:  &ByTutupSuk,
			MUTDB:    &NolRupiah,
			MUTCR:    &NolRupiah,
			NLSETOR:  &NolRupiah,
			TGLAWAL:  &TglNow,
			STATUS:   helper.StringPtr("1"),
		}
		if err := trx.Debug().Table("SHDBLC").Create(DataSimpSuk).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		trx.Commit()
		helper.RespondMaskedOk(c, "Berhasil membuat anggota baru", nil)
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
		}

		var Anggota obj.MSTCST
		if err := AppTx.Main.Debug().Table("MSTCST").Select("*").Where("kdcab = ? and cno = ?", ReqBody.KDCAB, ReqBody.CNO).Scan(&Anggota).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
		}

		DaftarRekening := make([]obj.SHDBLC, 0)
		if err := AppTx.Main.Debug().Table("SHDBLC").Select("*").Where("cno = ?", ReqBody.CNO).Scan(&DaftarRekening).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
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
