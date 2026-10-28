package controller

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Tobias-Kusuma/core-ksp-api/helper"
	"github.com/Tobias-Kusuma/core-ksp-api/obj"
	"github.com/gin-gonic/gin"
)

func GetProdukPinjaman() gin.HandlerFunc {
	return func(c *gin.Context) {
		AppTx, err := helper.Get_DB_ReqBody(c, 30, nil)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		querySelect := `
			rtrim(kd_produk) as kd_produk,
			rtrim(nm_produk) as nm_produk,
			rtrim(flgjmn) as flgjmn,
			case rtrim(jnsang)
				when '0' then 'PRK'
				when '1' then 'Bunga'
				when '2' then 'Pokok + Bunga'
				else rtrim(jnsang)
			end as jnsang,
			case rtrim(rmsbunga)
				when '2' then 'ANUITAS'
				when '4' then 'FLAT'
				when '5' then 'EFEKTIF'
				else rtrim(rmsbunga)
			end as rmsbunga,
			baserate,
			bulat			
		`

		ResData := make([]obj.ResProdukPinjaman, 0)
		if err := AppTx.Main.Debug().Table("PRODUK").
			Select(querySelect).
			Order("kd_produk asc").
			Scan(&ResData).Error; err != nil {
			log.Printf("ERROR: Gagal mengambil data produk pinjaman: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan daftar produk pinjaman", ResData)
	}
}

func PinjamanBaru() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqPinjamanBaru
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		Cno := strings.TrimSpace(ReqBody.CNO)
		Jenis := strings.TrimSpace(ReqBody.JENIS)

		if Kdcab == "" || Cno == "" || Jenis == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter kdcab, cno, dan jenis wajib diisi"))
			return
		}

		if ReqBody.PLF_DEBITUR <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Nilai plafon pinjaman (plf_debitur) harus lebih dari 0"))
			return
		}

		if ReqBody.JK_WKTAWAL <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Jangka waktu pinjaman (jk_wktawal) harus lebih dari 0 bulan"))
			return
		}

		// 1. Verifikasi keberadaan anggota di MSTCST
		var member struct {
			KDCAB string  `gorm:"column:kdcab"`
			CNO   string  `gorm:"column:cno"`
			CNM   *string `gorm:"column:cnm"`
		}
		if err := AppTx.Main.Debug().Table("MSTCST").
			Select("kdcab, cno, cnm").
			Where("kdcab = ? and cno = ?", Kdcab, Cno).
			First(&member).Error; err != nil {
			log.Printf("ERROR: Data anggota tidak ditemukan: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Data anggota tidak ditemukan"))
			return
		}

		NamaAnggota := ""
		if member.CNM != nil {
			NamaAnggota = strings.TrimSpace(*member.CNM)
		}

		// 2. Verifikasi produk pinjaman di PRODUK
		var produk obj.PRODUK
		if err := AppTx.Main.Debug().Table("PRODUK").
			Where("rtrim(kd_produk) = ?", Jenis).
			First(&produk).Error; err != nil {
			log.Printf("ERROR: Produk pinjaman tidak ditemukan: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Produk pinjaman tidak ditemukan"))
			return
		}

		NmProduk := ""
		if produk.NMPRODUK != nil {
			NmProduk = strings.TrimSpace(*produk.NMPRODUK)
		}

		// 3. Tentukan parameter suku bunga, rumus bunga, jenis angsuran, jenis kredit
		Bunga := ReqBody.BUNGA
		if Bunga <= 0 && produk.BASERATE != nil {
			Bunga = *produk.BASERATE
		}

		RmsBunga := helper.NormalizeRmsBunga(ReqBody.RMSBUNGA)
		if ReqBody.RMSBUNGA == "" && produk.RMSBUNGA != nil {
			RmsBunga = helper.NormalizeRmsBunga(*produk.RMSBUNGA)
		}

		JnsAngsur := helper.NormalizeJnsAngsur(ReqBody.JNSANGSUR)
		if ReqBody.JNSANGSUR == "" && produk.JNSANG != nil {
			JnsAngsur = helper.NormalizeJnsAngsur(*produk.JNSANG)
		}

		JnsKredit := strings.TrimSpace(ReqBody.JNSKREDIT)
		if JnsKredit == "" {
			JnsKredit = "1" // Default: 1 = Pinjaman Baru
		}

		// 4. Tanggal pencairan (tgl_awal) dan penentuan jatuh tempo (jt_tempo)
		tglAwal, err := helper.ParseDateFlexible(ReqBody.TGL_AWAL)
		if err != nil {
			helper.RespondMaskedError(c, true, err)
			return
		}

		tglAngsur := ReqBody.TGLANGSUR
		if tglAngsur <= 0 || tglAngsur > 31 {
			tglAngsur = tglAwal.Day()
		}

		tenor := int(ReqBody.JK_WKTAWAL)
		jtTempo := helper.CalculateInstallmentDate(tglAwal, tenor, tglAngsur)

		// 5. Generate jadwal angsuran berdasarkan rumus bunga dan jenis angsuran
		schedule, angsurDeb, angsPokok, angsBunga, totBng := helper.GenerateLoanSchedule(
			ReqBody.PLF_DEBITUR,
			tenor,
			Bunga,
			RmsBunga,
			JnsAngsur,
			tglAwal,
			tglAngsur,
		)

		tglAngs := schedule[0].TglVal

		// Biaya provisi dan administrasi
		Provisi := ReqBody.PROVISI
		Adm := ReqBody.ADM
		TotalBiaya := Provisi + Adm

		// 6. Jalankan transaksi database untuk insert KRD100 dan seluruh jadwal KRDANGS
		trx := AppTx.Main.Begin()
		defer func() {
			if r := recover(); r != nil {
				trx.Rollback()
			}
		}()

		// Generate nomor fasilitas pinjaman baru (nofas)
		// Format: kdcab (3) + jenis (2) + nomor urut dipadding 5 digit: %05d
		var lastUrut int64
		if err := trx.Table("KRD100").
			Select("isnull(max(try_cast(right(nofas, 5) as bigint)), 0) as last_urut").
			Where("kdcab = ? and Jenis = ?", Kdcab, Jenis).
			Scan(&lastUrut).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal membuat nomor fasilitas pinjaman: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Nofas := fmt.Sprintf("%s%s%05d", Kdcab, Jenis, lastUrut+1)

		tglNow := time.Now()
		nol := 0.0
		usrInput := "Via API"
		tglAngsurF := float64(tglAngsur)
		tenorF := float64(tenor)
		bcadHariAwal := helper.Round2(angsBunga / 30.0)

		kodeAO := strings.TrimSpace(ReqBody.KODEAO)
		if kodeAO == "" {
			kodeAO = "0000"
		}

		// 7. Simpan master pinjaman ke tabel KRD100
		newKrd100 := obj.KRD100{
			KDCAB:      Kdcab,
			CNO:        Cno,
			NOFAS:      Nofas,
			JENIS:      &Jenis,
			JNSANGSUR:  &JnsAngsur,
			JNSKREDIT:  &JnsKredit,
			PLFDEBITUR: &ReqBody.PLF_DEBITUR,
			JKWKTAWAL:  &tenorF,
			BUNGA:      &Bunga,
			TGLAWAL:    &tglAwal,
			JTTEMPO:    &jtTempo,
			TGLANGS:    &tglAngs,
			TGLACCRUE:  &tglAwal,
			FLAGBAYAR:  helper.StringPtr("0"),
			TGLANGSUR:  &tglAngsurF,
			NLDENDA:    &nol,
			RMSBUNGA:   &RmsBunga,
			SLDAKHIR:   &ReqBody.PLF_DEBITUR,
			ANGSURDEB:  &angsurDeb,
			TOTBNG:     &totBng,
			CADHARI:    &nol,
			BCADHARI:   &bcadHariAwal,
			BCADANG:    &nol,
			BJMLHARI:   &nol,
			PROVISI:    &Provisi,
			ADM:        &Adm,
			TOTALBIAYA: &TotalBiaya,
			ANGSPOKOK:  &angsPokok,
			ANGSBUNGA:  &angsBunga,
			NLPPAP:     &nol,
			NOAC:       helper.StringPtr(strings.TrimSpace(ReqBody.NOAC)),
			KODEAO:     &kodeAO,
			KOLEK:      helper.StringPtr("1"),
			NLJAMINAN:  &ReqBody.NLJAMINAN,
			NOREF:      helper.StringPtr(strings.TrimSpace(ReqBody.NOREF)),
			SALDOOLD:   &nol,
			BACKKET:    helper.StringPtr(strings.TrimSpace(ReqBody.KETERANGAN)),
			STSKRD:     helper.StringPtr("OK"),
			STSDEB:     helper.StringPtr("01"),
			STATUS:     helper.StringPtr("1"),
			USRID:      &usrInput,
			DUPD:       &tglNow,
			OTR:        &usrInput,
			DOTR:       &tglNow,
		}

		if err := trx.Table("KRD100").Create(&newKrd100).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal menyimpan data pinjaman (KRD100): %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// 8. Simpan jadwal angsuran ke tabel KRDANGS
		resJadwal := make([]obj.ResJadwalAngsuran, len(schedule))
		krdangsList := make([]obj.KRDANGS, len(schedule))

		for i, s := range schedule {
			angsurKeF := float64(s.AngsurKe)
			pokokVal := s.PokokJdw
			bungaVal := s.BungaJdw
			osdebVal := s.OsdebJdw
			bcadVal := s.BcadHari

			krdangsList[i] = obj.KRDANGS{
				KDCAB:     Kdcab,
				NOFAS:     Nofas,
				ANGSURKE:  angsurKeF,
				TGLVAL:    s.TglVal,
				POKOKJDW:  &pokokVal,
				BUNGAJDW:  &bungaVal,
				OSDEBJDW:  &osdebVal,
				NLPOKOK:   &pokokVal,
				NLBUNGA:   &bungaVal,
				NLDENDA:   &nol,
				NLSALDO:   &nol,
				BNGHARI:   &nol,
				BCADHARI:  &bcadVal,
				FLAGBAYAR: helper.StringPtr("0"),
				STATUS:    helper.StringPtr("1"),
				USRID:     &usrInput,
				DUPD:      &tglNow,
				OTR:       &usrInput,
				DOTR:      &tglNow,
			}

			resJadwal[i] = obj.ResJadwalAngsuran{
				ANGSURKE: angsurKeF,
				TGLVAL:   s.TglVal.Format(helper.DateFormat),
				POKOKJDW: s.PokokJdw,
				BUNGAJDW: s.BungaJdw,
				TOTAL:    s.Total,
				OSDEBJDW: s.OsdebJdw,
			}
		}

		if err := trx.Table("KRDANGS").Create(&krdangsList).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal membuat jadwal angsuran (KRDANGS): %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		if err := trx.Commit().Error; err != nil {
			log.Printf("ERROR: Gagal commit transaksi pinjaman: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// 9. Format response sukses
		resData := obj.ResPinjamanBaru{
			KDCAB:       Kdcab,
			CNO:         Cno,
			NOFAS:       Nofas,
			NAMA:        NamaAnggota,
			JENIS:       Jenis,
			NM_PRODUK:   NmProduk,
			PLF_DEBITUR: ReqBody.PLF_DEBITUR,
			JK_WKTAWAL:  ReqBody.JK_WKTAWAL,
			BUNGA:       Bunga,
			RMSBUNGA:    RmsBunga,
			JNSANGSUR:   JnsAngsur,
			TGL_AWAL:    tglAwal.Format(helper.DateFormat),
			JT_TEMPO:    jtTempo.Format(helper.DateFormat),
			TGLANGS:     tglAngs.Format(helper.DateFormat),
			ANGSUR_DEB:  angsurDeb,
			ANGSPOKOK:   angsPokok,
			ANGSBUNGA:   angsBunga,
			TOTBNG:      totBng,
			PROVISI:     Provisi,
			ADM:         Adm,
			TOTAL_BIAYA: TotalBiaya,
			JADWAL:      resJadwal,
		}

		helper.RespondMaskedOk(c, "Berhasil membuat pinjaman baru", resData)
	}
}

func GetJadwalAngsuran() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqGetJadwalAngsuran
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		Nofas := strings.TrimSpace(ReqBody.NOFAS)

		if Kdcab == "" || Nofas == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter kdcab dan nofas wajib diisi"))
			return
		}

		var jadwalList []obj.KRDANGS
		if err := AppTx.Main.Debug().Table("KRDANGS").
			Where("kdcab = ? and nofas = ?", Kdcab, Nofas).
			Order("angsurke asc").
			Find(&jadwalList).Error; err != nil {
			log.Printf("ERROR: Gagal mengambil data jadwal angsuran: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan jadwal angsuran", jadwalList)
	}
}

func GetDaftarPinjaman() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqDaftarPinjaman
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		Cno := strings.TrimSpace(ReqBody.CNO)

		if Cno == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter cno wajib diisi"))
			return
		}

		query := AppTx.Main.Debug().Table("KRD100").
			Select(`
				KRD100.kdcab,
				KRD100.cno,
				KRD100.nofas,
				rtrim(KRD100.Jenis) as jenis,
				isnull(rtrim(PRODUK.nm_produk), '') as nm_produk,
				isnull(KRD100.plf_debitur, 0) as plf_debitur,
				isnull(KRD100.sldakhir, 0) as sldakhir,
				KRD100.tgl_awal,
				KRD100.jt_tempo,
				isnull(KRD100.angsur_deb, 0) as angsur_deb,
				isnull(rtrim(KRD100.Kolek), '1') as kolek,
				isnull(rtrim(KRD100.Status), '1') as status,
				case 
					when isnull(KRD100.sldakhir, 0) <= 0 or rtrim(KRD100.Status) = '0' then 'LUNAS'
					else 'AKTIF'
				end as status_ket
			`).
			Joins("LEFT JOIN PRODUK ON rtrim(KRD100.Jenis) = rtrim(PRODUK.kd_produk)").
			Where("KRD100.cno = ?", Cno)

		if Kdcab != "" {
			query = query.Where("KRD100.kdcab = ?", Kdcab)
		}

		daftarPinjaman := make([]obj.ResDaftarPinjamanItem, 0)
		if err := query.Order("KRD100.tgl_awal desc, KRD100.nofas asc").Scan(&daftarPinjaman).Error; err != nil {
			log.Printf("ERROR: Gagal mengambil daftar pinjaman: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan daftar pinjaman", daftarPinjaman)
	}
}

func GetDataPinjaman() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqDataPinjaman
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		Nofas := strings.TrimSpace(ReqBody.NOFAS)

		if Nofas == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter nofas wajib diisi"))
			return
		}

		querySelect := `
			KRD100.kdcab,
			KRD100.cno,
			isnull(rtrim(MSTCST.cnm), '') as nama,
			KRD100.nofas,
			rtrim(KRD100.Jenis) as jenis,
			isnull(rtrim(PRODUK.nm_produk), '') as nm_produk,
			isnull(KRD100.plf_debitur, 0) as plf_debitur,
			isnull(KRD100.sldakhir, 0) as sldakhir,
			isnull(KRD100.jk_wktawal, 0) as jk_wktawal,
			isnull(KRD100.Bunga, 0) as bunga,
			isnull(rtrim(KRD100.rmsbunga), '') as rmsbunga,
			case rtrim(KRD100.rmsbunga)
				when '2' then 'ANUITAS'
				when '4' then 'FLAT'
				when '5' then 'EFEKTIF'
				else rtrim(KRD100.rmsbunga)
			end as rmsbunga_ket,
			isnull(rtrim(KRD100.JnsAngsur), '') as jnsangsur,
			case rtrim(KRD100.JnsAngsur)
				when '1' then 'Bunga'
				when '2' then 'Pokok + Bunga'
				else rtrim(KRD100.JnsAngsur)
			end as jnsangsur_ket,
			KRD100.tgl_awal,
			KRD100.jt_tempo,
			KRD100.tglangs,
			isnull(KRD100.tglangsur, 0) as tglangsur,
			isnull(KRD100.angsur_deb, 0) as angsur_deb,
			isnull(KRD100.angspokok, 0) as angspokok,
			isnull(KRD100.angsbunga, 0) as angsbunga,
			isnull(KRD100.totbng, 0) as totbng,
			isnull(KRD100.provisi, 0) as provisi,
			isnull(KRD100.Adm, 0) as adm,
			isnull(KRD100.TotalBiaya, 0) as total_biaya,
			isnull(rtrim(KRD100.Kolek), '1') as kolek,
			isnull(rtrim(KRD100.Stskrd), 'OK') as stskrd,
			case 
				when isnull(KRD100.sldakhir, 0) <= 0 or rtrim(KRD100.Status) = '0' then 'LUNAS'
				else 'AKTIF'
			end as status,
			isnull(rtrim(KRD100.Noac), '') as noac,
			isnull(rtrim(KRD100.KodeAO), '') as kodeao
		`

		query := AppTx.Main.Debug().Table("KRD100").
			Select(querySelect).
			Joins("LEFT JOIN MSTCST ON KRD100.cno = MSTCST.cno").
			Joins("LEFT JOIN PRODUK ON rtrim(KRD100.Jenis) = rtrim(PRODUK.kd_produk)").
			Where("KRD100.nofas = ?", Nofas)

		if Kdcab != "" {
			query = query.Where("KRD100.kdcab = ?", Kdcab)
		}

		var detailPinjaman obj.ResDataPinjaman
		if err := query.First(&detailPinjaman).Error; err != nil {
			log.Printf("ERROR: Data pinjaman tidak ditemukan: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Data pinjaman tidak ditemukan"))
			return
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan data pinjaman", detailPinjaman)
	}
}

func GetTagihan() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqTagihanPinjaman
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		Nofas := strings.TrimSpace(ReqBody.NOFAS)

		if Nofas == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter nofas wajib diisi"))
			return
		}

		// 1. Ambil data master pinjaman dan aturan denda dari PRODUK
		type loanHeaderInfo struct {
			Kdcab       string   `gorm:"column:kdcab"`
			Nofas       string   `gorm:"column:nofas"`
			Nama        string   `gorm:"column:nama"`
			Jenis       string   `gorm:"column:jenis"`
			NmProduk    string   `gorm:"column:nm_produk"`
			FlgDenda    *string  `gorm:"column:flgdenda"`
			HariDenda   *float64 `gorm:"column:hari_denda"`
			PersenDenda *float64 `gorm:"column:persen_denda"`
		}

		var header loanHeaderInfo
		queryHeader := AppTx.Main.Debug().Table("KRD100").
			Select(`
				KRD100.kdcab,
				KRD100.nofas,
				isnull(rtrim(MSTCST.cnm), '') as nama,
				rtrim(KRD100.Jenis) as jenis,
				isnull(rtrim(PRODUK.nm_produk), '') as nm_produk,
				PRODUK.flgdenda,
				PRODUK.hari_denda,
				PRODUK.persen_denda
			`).
			Joins("LEFT JOIN MSTCST ON KRD100.cno = MSTCST.cno").
			Joins("LEFT JOIN PRODUK ON rtrim(KRD100.Jenis) = rtrim(PRODUK.kd_produk)").
			Where("KRD100.nofas = ?", Nofas)

		if Kdcab != "" {
			queryHeader = queryHeader.Where("KRD100.kdcab = ?", Kdcab)
		}

		if err := queryHeader.First(&header).Error; err != nil {
			log.Printf("ERROR: Data pinjaman tidak ditemukan: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Data pinjaman tidak ditemukan"))
			return
		}

		flgDenda := "N"
		if header.FlgDenda != nil {
			flgDenda = strings.TrimSpace(*header.FlgDenda)
		}
		hariDenda := 0
		if header.HariDenda != nil {
			hariDenda = int(*header.HariDenda)
		}
		persenDenda := 0.0
		if header.PersenDenda != nil {
			persenDenda = *header.PersenDenda
		}

		// 2. Ambil seluruh angsuran yang belum dibayar (flagbayar = '0')
		var unpaidList []obj.KRDANGS
		queryAngs := AppTx.Main.Debug().Table("KRDANGS").
			Where("nofas = ? and flagbayar = ?", Nofas, "0")

		if Kdcab != "" {
			queryAngs = queryAngs.Where("kdcab = ?", Kdcab)
		}

		if err := queryAngs.Order("angsurke asc").Find(&unpaidList).Error; err != nil {
			log.Printf("ERROR: Gagal mengambil data jadwal angsuran: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		tagihanDetail := make([]obj.ResItemTagihan, 0, len(unpaidList))
		totalPokok := 0.0
		totalBunga := 0.0
		totalDenda := 0.0

		for _, item := range unpaidList {
			pokok := 0.0
			if item.NLPOKOK != nil {
				pokok = *item.NLPOKOK
			} else if item.POKOKJDW != nil {
				pokok = *item.POKOKJDW
			}

			bunga := 0.0
			if item.NLBUNGA != nil {
				bunga = *item.NLBUNGA
			} else if item.BUNGAJDW != nil {
				bunga = *item.BUNGAJDW
			}

			denda := 0.0
			if item.NLDENDA != nil {
				denda = *item.NLDENDA
			}

			// Hitung denda keterlambatan berjalan jika lewat tglval
			tglValDate := time.Date(item.TGLVAL.Year(), item.TGLVAL.Month(), item.TGLVAL.Day(), 0, 0, 0, 0, item.TGLVAL.Location())
			daysLate := 0
			if today.After(tglValDate) {
				daysLate = int(today.Sub(tglValDate).Hours() / 24)
			}

			if daysLate > hariDenda && flgDenda == "Y" && persenDenda > 0 {
				calcDenda := helper.Round2((pokok + bunga) * (persenDenda / 100.0) * float64(daysLate-hariDenda))
				if calcDenda > denda {
					denda = calcDenda
				}
			}

			subtotal := helper.Round2(pokok + bunga + denda)

			totalPokok += pokok
			totalBunga += bunga
			totalDenda += denda

			statusItem := "BELUM JATUH TEMPO"
			if daysLate > 0 {
				statusItem = fmt.Sprintf("TERLAMBAT %d HARI", daysLate)
			}

			tagihanDetail = append(tagihanDetail, obj.ResItemTagihan{
				ANGSURKE:       item.ANGSURKE,
				TGLVAL:         item.TGLVAL.Format(helper.DateFormat),
				POKOK:          pokok,
				BUNGA:          bunga,
				DENDA:          denda,
				HARI_TERLAMBAT: daysLate,
				TOTAL:          subtotal,
				STATUS:         statusItem,
			})
		}

		grandTotal := helper.Round2(totalPokok + totalBunga + totalDenda)

		res := obj.ResTagihanPinjaman{
			KDCAB:               header.Kdcab,
			NOFAS:               header.Nofas,
			NAMA:                header.Nama,
			JENIS:               header.Jenis,
			NM_PRODUK:           header.NmProduk,
			TOTAL_POKOK:         helper.Round2(totalPokok),
			TOTAL_BUNGA:         helper.Round2(totalBunga),
			TOTAL_DENDA:         helper.Round2(totalDenda),
			GRAND_TOTAL:         grandTotal,
			JUMLAH_ANGSURAN_TGK: len(unpaidList),
			TAGIHAN_DETAIL:      tagihanDetail,
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan tagihan pinjaman", res)
	}
}

func SimulasiPinjaman() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqSimulasiPinjaman
		if err := c.ShouldBindJSON(&ReqBody); err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Format request tidak valid: %s", err.Error()))
			return
		}

		if ReqBody.PLF_DEBITUR <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Nilai plafon pinjaman (plf_debitur) harus lebih dari 0"))
			return
		}

		if ReqBody.JK_WKTAWAL <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Jangka waktu pinjaman (jk_wktawal) harus lebih dari 0 bulan"))
			return
		}

		if ReqBody.BUNGA < 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Suku bunga pinjaman (bunga) tidak boleh negatif"))
			return
		}

		RmsBunga := helper.NormalizeRmsBunga(ReqBody.RMSBUNGA)
		RmsBungaKet := helper.RmsBungaLabel(RmsBunga)

		JnsAngsur := helper.NormalizeJnsAngsur(ReqBody.JNSANGSUR)
		JnsAngsurKet := helper.JnsAngsurLabel(JnsAngsur)

		tglAwal, err := helper.ParseDateFlexible(ReqBody.TGL_AWAL)
		if err != nil {
			helper.RespondMaskedError(c, true, err)
			return
		}

		tglAngsur := ReqBody.TGLANGSUR
		if tglAngsur <= 0 || tglAngsur > 31 {
			tglAngsur = tglAwal.Day()
		}

		tenor := int(ReqBody.JK_WKTAWAL)
		jtTempo := helper.CalculateInstallmentDate(tglAwal, tenor, tglAngsur)

		schedule, angsurDeb, angsPokok, angsBunga, totBng := helper.GenerateLoanSchedule(
			ReqBody.PLF_DEBITUR,
			tenor,
			ReqBody.BUNGA,
			RmsBunga,
			JnsAngsur,
			tglAwal,
			tglAngsur,
		)

		resJadwal := make([]obj.ResJadwalAngsuran, len(schedule))
		for i, s := range schedule {
			resJadwal[i] = obj.ResJadwalAngsuran{
				ANGSURKE: float64(s.AngsurKe),
				TGLVAL:   s.TglVal.Format(helper.DateFormat),
				POKOKJDW: s.PokokJdw,
				BUNGAJDW: s.BungaJdw,
				TOTAL:    s.Total,
				OSDEBJDW: s.OsdebJdw,
			}
		}

		totalBayar := helper.Round2(ReqBody.PLF_DEBITUR + totBng)

		res := obj.ResSimulasiPinjaman{
			PLF_DEBITUR:   ReqBody.PLF_DEBITUR,
			JK_WKTAWAL:    ReqBody.JK_WKTAWAL,
			BUNGA:         ReqBody.BUNGA,
			RMSBUNGA:      RmsBunga,
			RMSBUNGA_KET:  RmsBungaKet,
			JNSANGSUR:     JnsAngsur,
			JNSANGSUR_KET: JnsAngsurKet,
			TGL_AWAL:      tglAwal.Format(helper.DateFormat),
			JT_TEMPO:      jtTempo.Format(helper.DateFormat),
			ANGSUR_DEB:    angsurDeb,
			ANGSPOKOK:     angsPokok,
			ANGSBUNGA:     angsBunga,
			TOTBNG:        totBng,
			TOTAL_BAYAR:   totalBayar,
			JADWAL:        resJadwal,
		}

		helper.RespondMaskedOk(c, "Berhasil membuat simulasi pinjaman", res)
	}
}

func BayarAngsuran() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqBayarAngsuran
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Nofas := strings.TrimSpace(ReqBody.NOFAS)
		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		if Nofas == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter nofas wajib diisi"))
			return
		}

		// 1. Cek fasilitas pinjaman di KRD100
		var loan obj.KRD100
		queryLoan := AppTx.Main.Debug().Table("KRD100").Where("nofas = ?", Nofas)
		if Kdcab != "" {
			queryLoan = queryLoan.Where("kdcab = ?", Kdcab)
		}
		if err := queryLoan.First(&loan).Error; err != nil {
			log.Printf("ERROR: Data pinjaman tidak ditemukan: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Data fasilitas pinjaman tidak ditemukan"))
			return
		}

		if Kdcab == "" {
			Kdcab = loan.KDCAB
		}

		// Periksa apakah pinjaman masih aktif
		statusLoan := ""
		if loan.STATUS != nil {
			statusLoan = strings.TrimSpace(*loan.STATUS)
		}
		sldAkhirLoan := 0.0
		if loan.SLDAKHIR != nil {
			sldAkhirLoan = *loan.SLDAKHIR
		}
		if statusLoan == "0" || sldAkhirLoan <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Pinjaman %s sudah lunas atau tidak aktif", Nofas))
			return
		}

		// 2. Tentukan jadwal angsuran yang akan dibayar
		var targetAngs obj.KRDANGS
		if ReqBody.ANGSURKE > 0 {
			// Cek apakah angsuran ke tersebut ada dan belum lunas
			if err := AppTx.Main.Debug().Table("KRDANGS").
				Where("kdcab = ? and nofas = ? and angsurke = ?", Kdcab, Nofas, ReqBody.ANGSURKE).
				First(&targetAngs).Error; err != nil {
				helper.RespondMaskedError(c, true, fmt.Errorf("Jadwal angsuran ke-%d tidak ditemukan", ReqBody.ANGSURKE))
				return
			}
			flg := "0"
			if targetAngs.FLAGBAYAR != nil {
				flg = strings.TrimSpace(*targetAngs.FLAGBAYAR)
			}
			if flg == "1" {
				helper.RespondMaskedError(c, true, fmt.Errorf("Angsuran ke-%d sudah dibayar", ReqBody.ANGSURKE))
				return
			}

			// Validasi urutan: apakah ada angsuran sebelumnya yang belum lunas?
			var priorUnpaidCount int64
			AppTx.Main.Debug().Table("KRDANGS").
				Where("kdcab = ? and nofas = ? and angsurke < ? and flagbayar = ?", Kdcab, Nofas, ReqBody.ANGSURKE, "0").
				Count(&priorUnpaidCount)
			if priorUnpaidCount > 0 {
				helper.RespondMaskedError(c, true, fmt.Errorf("Angsuran sebelumnya belum lunas. Pembayaran harus dilakukan berurutan"))
				return
			}
		} else {
			// Ambil angsuran pertama yang belum dibayar
			if err := AppTx.Main.Debug().Table("KRDANGS").
				Where("kdcab = ? and nofas = ? and flagbayar = ?", Kdcab, Nofas, "0").
				Order("angsurke asc").
				First(&targetAngs).Error; err != nil {
				helper.RespondMaskedError(c, true, fmt.Errorf("Seluruh jadwal angsuran untuk fasilitas %s sudah lunas", Nofas))
				return
			}
		}

		// 3. Ambil konfigurasi denda dari PRODUK
		var produk obj.PRODUK
		jenisPinjaman := ""
		if loan.JENIS != nil {
			jenisPinjaman = strings.TrimSpace(*loan.JENIS)
		}
		_ = AppTx.Main.Debug().Table("PRODUK").Where("kd_produk = ?", jenisPinjaman).First(&produk).Error

		flgDenda := "N"
		if produk.FLGDENDA != nil {
			flgDenda = strings.TrimSpace(*produk.FLGDENDA)
		}
		hariDenda := 0
		if produk.HARIDENDA != nil {
			hariDenda = int(*produk.HARIDENDA)
		}
		persenDenda := 0.0
		if produk.PERSENDENDA != nil {
			persenDenda = *produk.PERSENDENDA
		}

		pokok := 0.0
		if targetAngs.POKOKJDW != nil {
			pokok = *targetAngs.POKOKJDW
		}
		bunga := 0.0
		if targetAngs.BUNGAJDW != nil {
			bunga = *targetAngs.BUNGAJDW
		}
		denda := 0.0
		if targetAngs.NLDENDA != nil {
			denda = *targetAngs.NLDENDA
		}

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		tglValDate := time.Date(targetAngs.TGLVAL.Year(), targetAngs.TGLVAL.Month(), targetAngs.TGLVAL.Day(), 0, 0, 0, 0, targetAngs.TGLVAL.Location())
		daysLate := 0
		if today.After(tglValDate) {
			daysLate = int(today.Sub(tglValDate).Hours() / 24)
		}

		if daysLate > hariDenda && flgDenda == "Y" && persenDenda > 0 {
			calcDenda := helper.Round2((pokok + bunga) * (persenDenda / 100.0) * float64(daysLate-hariDenda))
			if calcDenda > denda {
				denda = calcDenda
			}
		}

		totalBayar := helper.Round2(pokok + bunga + denda)

		// Validasi nominal jika disertakan
		if ReqBody.NOMINAL > 0 && ReqBody.NOMINAL < totalBayar {
			helper.RespondMaskedError(c, true, fmt.Errorf("Nominal pembayaran (Rp %.2f) kurang dari total tagihan (Rp %.2f)",
				ReqBody.NOMINAL, totalBayar))
			return
		}

		crBayar := strings.ToUpper(strings.TrimSpace(ReqBody.CRBAYAR))
		if crBayar == "" {
			crBayar = "R" // Default tabungan
		}
		if crBayar != "R" && crBayar != "T" && crBayar != "B" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Metode cara bayar (crbayar) tidak valid. Gunakan 'R' (Tabungan), 'T' (Tunai), atau 'B' (Bank)"))
			return
		}

		noacDebet := strings.TrimSpace(ReqBody.NOAC_DEBET)
		if crBayar == "R" {
			if noacDebet == "" && loan.NOAC != nil {
				noacDebet = strings.TrimSpace(*loan.NOAC)
			}
			if noacDebet == "" {
				helper.RespondMaskedError(c, true, fmt.Errorf("Nomor rekening tabungan (noac_debet) untuk pendebetan wajib diisi"))
				return
			}

			// Validasi ketersediaan saldo rekening di SHDBLC
			var rekTabungan obj.SHDBLC
			if err := AppTx.Main.Debug().Table("SHDBLC").
				Where("kdcab = ? and noac = ? and status = ?", Kdcab, noacDebet, "1").
				First(&rekTabungan).Error; err != nil {
				log.Printf("ERROR: Rekening tabungan %s tidak ditemukan atau tidak aktif: %s", noacDebet, err.Error())
				helper.RespondMaskedError(c, true, fmt.Errorf("Rekening tabungan pendebetan tidak ditemukan atau tidak aktif"))
				return
			}

			sldPakai := 0.0
			if rekTabungan.SLDPAKAI != nil {
				sldPakai = *rekTabungan.SLDPAKAI
			} else if rekTabungan.SLDAKHIR != nil {
				sldPakai = *rekTabungan.SLDAKHIR
			}
			if sldPakai < totalBayar {
				helper.RespondMaskedError(c, true, fmt.Errorf("Saldo rekening tabungan tidak mencukupi (Saldo: Rp %.2f, Tagihan: Rp %.2f)",
					sldPakai, totalBayar))
				return
			}
		}

		angsurKeInt := int(targetAngs.ANGSURKE)
		ketTrn := strings.TrimSpace(ReqBody.KETERANGAN)
		if ketTrn == "" {
			ketTrn = fmt.Sprintf("Pembayaran Angsuran ke %d", angsurKeInt)
		}

		// 4. Jalankan Database Transaction
		trx := AppTx.Main.Begin()
		defer func() {
			if r := recover(); r != nil {
				trx.Rollback()
			}
		}()

		// Generate nomor referensi transaksi
		// Format: KRD + kdcab (3) + YYYYMMDD (8) + urut (6) = 20 chars
		tglStr := now.Format("20060102")
		patternRef := fmt.Sprintf("KRD%s%s%%", Kdcab, tglStr)
		var lastUrutRef int64
		if err := trx.Table("KRDH400").
			Select("isnull(max(try_cast(right(noref, 6) as bigint)), 0)").
			Where("noref LIKE ?", patternRef).
			Scan(&lastUrutRef).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal generate nomor referensi: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}
		noRef := fmt.Sprintf("KRD%s%s%06d", Kdcab, tglStr, lastUrutRef+1)

		// Debet saldo tabungan jika crbayar == "R"
		if crBayar == "R" {
			var rek obj.SHDBLC
			if err := trx.Table("SHDBLC").
				Where("kdcab = ? and noac = ?", Kdcab, noacDebet).
				First(&rek).Error; err != nil {
				trx.Rollback()
				helper.RespondMaskedError(c, false, err)
				return
			}
			curSldAkhir := 0.0
			if rek.SLDAKHIR != nil {
				curSldAkhir = *rek.SLDAKHIR
			}
			curSldPakai := 0.0
			if rek.SLDPAKAI != nil {
				curSldPakai = *rek.SLDPAKAI
			}
			newSldAkhir := helper.Round2(curSldAkhir - totalBayar)
			newSldPakai := helper.Round2(curSldPakai - totalBayar)

			if err := trx.Table("SHDBLC").
				Where("kdcab = ? and noac = ?", Kdcab, noacDebet).
				Updates(map[string]interface{}{
					"sldakhir": newSldAkhir,
					"sldpakai": newSldPakai,
				}).Error; err != nil {
				trx.Rollback()
				log.Printf("ERROR: Gagal memotong saldo tabungan: %s", err.Error())
				helper.RespondMaskedError(c, false, err)
				return
			}
		}

		// Update KRDANGS
		newSisaLoan := helper.Round2(sldAkhirLoan - pokok)
		if newSisaLoan < 0 {
			newSisaLoan = 0
		}
		flagBayarLunas := "1"
		statusLunas := "LUNAS"
		usrId := "Via API"
		tglTrnNow := now

		angsUpdates := map[string]interface{}{
			"flagbayar": flagBayarLunas,
			"tglbyr":    tglTrnNow,
			"nlpokok":   pokok,
			"nlbunga":   bunga,
			"nldenda":   denda,
			"nlsaldo":   newSisaLoan,
			"status":    statusLunas,
			"usrid":     usrId,
			"dupd":      tglTrnNow,
		}
		if err := trx.Table("KRDANGS").
			Where("kdcab = ? and nofas = ? and angsurke = ?", Kdcab, Nofas, targetAngs.ANGSURKE).
			Updates(angsUpdates).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal mengupdate data angsuran KRDANGS: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// Cek sisa angsuran yang belum dibayar di KRDANGS
		var remainingUnpaid int64
		trx.Table("KRDANGS").
			Where("kdcab = ? and nofas = ? and flagbayar = ?", Kdcab, Nofas, "0").
			Count(&remainingUnpaid)

		loanUpdates := map[string]interface{}{
			"sldakhir": newSisaLoan,
			"usrid":    usrId,
			"dupd":     tglTrnNow,
		}
		statusPinjaman := "AKTIF"
		if remainingUnpaid == 0 || newSisaLoan <= 0 {
			loanUpdates["Status"] = "0"
			loanUpdates["Stskrd"] = "LN"
			loanUpdates["tgl_lunas"] = tglTrnNow
			loanUpdates["flagbayar"] = "1"
			statusPinjaman = "LUNAS"
		}

		if err := trx.Table("KRD100").
			Where("kdcab = ? and nofas = ?", Kdcab, Nofas).
			Updates(loanUpdates).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal mengupdate saldo KRD100: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// Insert KRDH400
		stsTrn := "01"
		stsHdr := "1"
		krdH400 := obj.KRDH400{
			KDCAB:   Kdcab,
			NOFAS:   Nofas,
			NOREF:   noRef,
			TGLTRN:  tglTrnNow,
			CRBAYAR: &crBayar,
			NOACDB:  &noacDebet,
			STSTRN:  &stsTrn,
			KET:     &ketTrn,
			STATUS:  &stsHdr,
			USRID:   &usrId,
			DUPD:    &tglTrnNow,
		}
		if err := trx.Table("KRDH400").Create(&krdH400).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal mencatat header pembayaran KRDH400: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// Insert KRD400
		nlPinalty := 0.0
		stsDet := "1"
		krd400 := obj.KRD400{
			KDCAB:     Kdcab,
			NOFAS:     Nofas,
			ANGSURKE:  targetAngs.ANGSURKE,
			NOREF:     noRef,
			TGLTRN:    tglTrnNow,
			TGLVAL:    &targetAngs.TGLVAL,
			NLPOKOK:   &pokok,
			NLBUNGA:   &bunga,
			NLDENDA:   &denda,
			NLPINALTY: &nlPinalty,
			STATUS:    &stsDet,
			USRID:     &usrId,
			DUPD:      &tglTrnNow,
		}
		if err := trx.Table("KRD400").Create(&krd400).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal mencatat detail pembayaran KRD400: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// Commit
		if err := trx.Commit().Error; err != nil {
			log.Printf("ERROR: Gagal commit transaksi pembayaran: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		res := obj.ResBayarAngsuran{
			NOREF:            noRef,
			KDCAB:            Kdcab,
			NOFAS:            Nofas,
			ANGSURKE:         targetAngs.ANGSURKE,
			TGLBYR:           tglTrnNow.Format(helper.DateTimeFormat),
			CRBAYAR:          crBayar,
			NOAC_DEBET:       noacDebet,
			POKOK:            pokok,
			BUNGA:            bunga,
			DENDA:            denda,
			TOTAL_BAYAR:      totalBayar,
			SISA_OUTSTANDING: newSisaLoan,
			STATUS_PINJAMAN:  statusPinjaman,
			KETERANGAN:       ketTrn,
		}

		helper.RespondMaskedOk(c, "Berhasil memproses pembayaran angsuran", res)
	}
}

func PelunasanKredit() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqPelunasanKredit
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Nofas := strings.TrimSpace(ReqBody.NOFAS)
		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		if Nofas == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter nofas wajib diisi"))
			return
		}

		// 1. Cek fasilitas pinjaman di KRD100
		var loan obj.KRD100
		queryLoan := AppTx.Main.Debug().Table("KRD100").Where("nofas = ?", Nofas)
		if Kdcab != "" {
			queryLoan = queryLoan.Where("kdcab = ?", Kdcab)
		}
		if err := queryLoan.First(&loan).Error; err != nil {
			log.Printf("ERROR: Data pinjaman tidak ditemukan: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Data fasilitas pinjaman tidak ditemukan"))
			return
		}

		if Kdcab == "" {
			Kdcab = loan.KDCAB
		}

		// Periksa apakah pinjaman masih aktif
		statusLoan := ""
		if loan.STATUS != nil {
			statusLoan = strings.TrimSpace(*loan.STATUS)
		}
		sldAkhirLoan := 0.0
		if loan.SLDAKHIR != nil {
			sldAkhirLoan = *loan.SLDAKHIR
		}
		if statusLoan == "0" || sldAkhirLoan <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Pinjaman %s sudah lunas atau tidak aktif", Nofas))
			return
		}

		// 2. Ambil seluruh angsuran yang belum dibayar (flagbayar = '0')
		var unpaidList []obj.KRDANGS
		if err := AppTx.Main.Debug().Table("KRDANGS").
			Where("kdcab = ? and nofas = ? and flagbayar = ?", Kdcab, Nofas, "0").
			Order("angsurke asc").
			Find(&unpaidList).Error; err != nil {
			log.Printf("ERROR: Gagal mengambil data jadwal angsuran: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		if len(unpaidList) == 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Tidak ada jadwal angsuran yang belum lunas"))
			return
		}

		// 3. Ambil konfigurasi denda dan penalty dari PRODUK
		var produk obj.PRODUK
		jenisPinjaman := ""
		if loan.JENIS != nil {
			jenisPinjaman = strings.TrimSpace(*loan.JENIS)
		}
		_ = AppTx.Main.Debug().Table("PRODUK").Where("kd_produk = ?", jenisPinjaman).First(&produk).Error

		flgDenda := "N"
		if produk.FLGDENDA != nil {
			flgDenda = strings.TrimSpace(*produk.FLGDENDA)
		}
		hariDenda := 0
		if produk.HARIDENDA != nil {
			hariDenda = int(*produk.HARIDENDA)
		}
		persenDenda := 0.0
		if produk.PERSENDENDA != nil {
			persenDenda = *produk.PERSENDENDA
		}
		persenPenalty := 0.0
		if produk.PENALTY != nil {
			persenPenalty = *produk.PENALTY
		}

		// 4. Hitung rincian pelunasan:
		// - Sisa Pokok = sldakhir
		// - Bunga berjalan = bunga jadwal angsuran pertama yang unpaid
		// - Denda = akumulasi denda dari angsuran yang menunggak
		// - Penalty = sldakhir * persenPenalty / 100
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		sisaPokok := helper.Round2(sldAkhirLoan)
		bungaBerjalan := 0.0
		if unpaidList[0].BUNGAJDW != nil {
			bungaBerjalan = helper.Round2(*unpaidList[0].BUNGAJDW)
		}
		totalDenda := 0.0
		for _, u := range unpaidList {
			itemPokok := 0.0
			if u.POKOKJDW != nil {
				itemPokok = *u.POKOKJDW
			}
			itemBunga := 0.0
			if u.BUNGAJDW != nil {
				itemBunga = *u.BUNGAJDW
			}
			itemDenda := 0.0
			if u.NLDENDA != nil {
				itemDenda = *u.NLDENDA
			}
			tglValDate := time.Date(u.TGLVAL.Year(), u.TGLVAL.Month(), u.TGLVAL.Day(), 0, 0, 0, 0, u.TGLVAL.Location())
			daysLate := 0
			if today.After(tglValDate) {
				daysLate = int(today.Sub(tglValDate).Hours() / 24)
			}
			if daysLate > hariDenda && flgDenda == "Y" && persenDenda > 0 {
				calcDenda := helper.Round2((itemPokok + itemBunga) * (persenDenda / 100.0) * float64(daysLate-hariDenda))
				if calcDenda > itemDenda {
					itemDenda = calcDenda
				}
			}
			totalDenda = helper.Round2(totalDenda + itemDenda)
		}

		penalty := helper.Round2(sisaPokok * (persenPenalty / 100.0))
		totalPelunasan := helper.Round2(sisaPokok + bungaBerjalan + totalDenda + penalty)

		// 5. Validasi cara bayar dan rekening tabungan debet
		crBayar := strings.ToUpper(strings.TrimSpace(ReqBody.CRBAYAR))
		if crBayar == "" {
			crBayar = "R"
		}
		if crBayar != "R" && crBayar != "T" && crBayar != "B" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Metode cara bayar (crbayar) tidak valid. Gunakan 'R' (Tabungan), 'T' (Tunai), atau 'B' (Bank)"))
			return
		}

		noacDebet := strings.TrimSpace(ReqBody.NOAC_DEBET)
		if crBayar == "R" {
			if noacDebet == "" && loan.NOAC != nil {
				noacDebet = strings.TrimSpace(*loan.NOAC)
			}
			if noacDebet == "" {
				helper.RespondMaskedError(c, true, fmt.Errorf("Nomor rekening tabungan (noac_debet) untuk pendebetan wajib diisi"))
				return
			}

			var rekTabungan obj.SHDBLC
			if err := AppTx.Main.Debug().Table("SHDBLC").
				Where("kdcab = ? and noac = ? and status = ?", Kdcab, noacDebet, "1").
				First(&rekTabungan).Error; err != nil {
				log.Printf("ERROR: Rekening tabungan %s tidak ditemukan atau tidak aktif: %s", noacDebet, err.Error())
				helper.RespondMaskedError(c, true, fmt.Errorf("Rekening tabungan pendebetan tidak ditemukan atau tidak aktif"))
				return
			}

			sldPakai := 0.0
			if rekTabungan.SLDPAKAI != nil {
				sldPakai = *rekTabungan.SLDPAKAI
			} else if rekTabungan.SLDAKHIR != nil {
				sldPakai = *rekTabungan.SLDAKHIR
			}
			if sldPakai < totalPelunasan {
				helper.RespondMaskedError(c, true, fmt.Errorf("Saldo rekening tabungan tidak mencukupi (Saldo: Rp %.2f, Total Pelunasan: Rp %.2f)",
					sldPakai, totalPelunasan))
				return
			}
		}

		ketTrn := strings.TrimSpace(ReqBody.KETERANGAN)
		if ketTrn == "" {
			ketTrn = "Pelunasan Kredit"
		}

		// 6. Jalankan Database Transaction
		trx := AppTx.Main.Begin()
		defer func() {
			if r := recover(); r != nil {
				trx.Rollback()
			}
		}()

		// Generate nomor referensi pelunasan: LNP + kdcab (3) + YYYYMMDD (8) + urut (6) = 20 chars
		tglStr := now.Format("20060102")
		patternRef := fmt.Sprintf("LNP%s%s%%", Kdcab, tglStr)
		var lastUrutRef int64
		if err := trx.Table("KRDH400").
			Select("isnull(max(try_cast(right(noref, 6) as bigint)), 0)").
			Where("noref LIKE ?", patternRef).
			Scan(&lastUrutRef).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal generate nomor referensi pelunasan: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}
		noRef := fmt.Sprintf("LNP%s%s%06d", Kdcab, tglStr, lastUrutRef+1)

		// Debet rekening tabungan jika crbayar == "R"
		if crBayar == "R" {
			var rek obj.SHDBLC
			if err := trx.Table("SHDBLC").
				Where("kdcab = ? and noac = ?", Kdcab, noacDebet).
				First(&rek).Error; err != nil {
				trx.Rollback()
				helper.RespondMaskedError(c, false, err)
				return
			}
			curSldAkhir := 0.0
			if rek.SLDAKHIR != nil {
				curSldAkhir = *rek.SLDAKHIR
			}
			curSldPakai := 0.0
			if rek.SLDPAKAI != nil {
				curSldPakai = *rek.SLDPAKAI
			}
			newSldAkhir := helper.Round2(curSldAkhir - totalPelunasan)
			newSldPakai := helper.Round2(curSldPakai - totalPelunasan)

			if err := trx.Table("SHDBLC").
				Where("kdcab = ? and noac = ?", Kdcab, noacDebet).
				Updates(map[string]interface{}{
					"sldakhir": newSldAkhir,
					"sldpakai": newSldPakai,
				}).Error; err != nil {
				trx.Rollback()
				log.Printf("ERROR: Gagal memotong saldo tabungan pelunasan: %s", err.Error())
				helper.RespondMaskedError(c, false, err)
				return
			}
		}

		tglTrnNow := now
		usrId := "Via API"
		flagBayarLunas := "1"
		statusLunasEarly := "LUNAS DIPERCEPAT"

		// Update KRDANGS untuk seluruh angsuran yang belum lunas
		for idx, u := range unpaidList {
			nlPokokItem := 0.0
			if u.POKOKJDW != nil {
				nlPokokItem = *u.POKOKJDW
			}
			nlBungaItem := 0.0
			if idx == 0 {
				nlBungaItem = bungaBerjalan
			}
			nlDendaItem := 0.0
			if idx == 0 {
				nlDendaItem = totalDenda
			}

			if err := trx.Table("KRDANGS").
				Where("kdcab = ? and nofas = ? and angsurke = ?", Kdcab, Nofas, u.ANGSURKE).
				Updates(map[string]interface{}{
					"flagbayar": flagBayarLunas,
					"tglbyr":    tglTrnNow,
					"nlpokok":   nlPokokItem,
					"nlbunga":   nlBungaItem,
					"nldenda":   nlDendaItem,
					"nlsaldo":   0.0,
					"status":    statusLunasEarly,
					"usrid":     usrId,
					"dupd":      tglTrnNow,
				}).Error; err != nil {
				trx.Rollback()
				log.Printf("ERROR: Gagal mengupdate angsuran KRDANGS saat pelunasan: %s", err.Error())
				helper.RespondMaskedError(c, false, err)
				return
			}
		}

		// Update KRD100 master
		if err := trx.Table("KRD100").
			Where("kdcab = ? and nofas = ?", Kdcab, Nofas).
			Updates(map[string]interface{}{
				"sldakhir":  0.0,
				"Status":    "0",
				"Stskrd":    "LN",
				"tgl_lunas": tglTrnNow,
				"flagbayar": "1",
				"usrid":     usrId,
				"dupd":      tglTrnNow,
			}).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal mengupdate status lunas KRD100: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// Insert KRDH400
		stsTrn := "02"
		stsHdr := "1"
		krdH400 := obj.KRDH400{
			KDCAB:   Kdcab,
			NOFAS:   Nofas,
			NOREF:   noRef,
			TGLTRN:  tglTrnNow,
			CRBAYAR: &crBayar,
			NOACDB:  &noacDebet,
			STSTRN:  &stsTrn,
			KET:     &ketTrn,
			STATUS:  &stsHdr,
			USRID:   &usrId,
			DUPD:    &tglTrnNow,
		}
		if err := trx.Table("KRDH400").Create(&krdH400).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal mencatat header pelunasan KRDH400: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		// Insert KRD400 detail
		stsDet := "1"
		krd400 := obj.KRD400{
			KDCAB:     Kdcab,
			NOFAS:     Nofas,
			ANGSURKE:  unpaidList[0].ANGSURKE,
			NOREF:     noRef,
			TGLTRN:    tglTrnNow,
			TGLVAL:    &unpaidList[0].TGLVAL,
			NLPOKOK:   &sisaPokok,
			NLBUNGA:   &bungaBerjalan,
			NLDENDA:   &totalDenda,
			NLPINALTY: &penalty,
			STATUS:    &stsDet,
			USRID:     &usrId,
			DUPD:      &tglTrnNow,
		}
		if err := trx.Table("KRD400").Create(&krd400).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal mencatat detail pelunasan KRD400: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		if err := trx.Commit().Error; err != nil {
			log.Printf("ERROR: Gagal commit transaksi pelunasan: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		res := obj.ResPelunasanKredit{
			NOREF:           noRef,
			KDCAB:           Kdcab,
			NOFAS:           Nofas,
			TGL_LUNAS:       tglTrnNow.Format(helper.DateTimeFormat),
			CRBAYAR:         crBayar,
			NOAC_DEBET:      noacDebet,
			SISA_POKOK:      sisaPokok,
			BUNGA:           bungaBerjalan,
			DENDA:           totalDenda,
			PENALTY:         penalty,
			TOTAL_PELUNASAN: totalPelunasan,
			STATUS:          "LUNAS",
			KETERANGAN:      ketTrn,
		}

		helper.RespondMaskedOk(c, "Berhasil memproses pelunasan kredit", res)
	}
}

func GetRiwayatBayar() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqRiwayatBayar
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Nofas := strings.TrimSpace(ReqBody.NOFAS)
		Kdcab := strings.TrimSpace(ReqBody.KDCAB)

		if Nofas == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter nofas wajib diisi"))
			return
		}

		type queryRow struct {
			NOREF     string    `gorm:"column:noref"`
			TGLTRN    time.Time `gorm:"column:tgltrn"`
			STSTRN    string    `gorm:"column:ststrn"`
			KET       string    `gorm:"column:ket"`
			CRBAYAR   string    `gorm:"column:crbayar"`
			NOACDB    string    `gorm:"column:noacdb"`
			ANGSURKE  float64   `gorm:"column:angsurke"`
			NLPOKOK   float64   `gorm:"column:nlpokok"`
			NLBUNGA   float64   `gorm:"column:nlbunga"`
			NLDENDA   float64   `gorm:"column:nldenda"`
			NLPINALTY float64   `gorm:"column:nlpinalty"`
		}

		var rows []queryRow
		q := AppTx.Main.Debug().Table("KRDH400").
			Select(`
				rtrim(KRDH400.noref) as noref,
				KRDH400.tgltrn,
				isnull(rtrim(KRDH400.Ststrn), '') as ststrn,
				isnull(rtrim(KRDH400.Ket), '') as ket,
				isnull(rtrim(KRDH400.Crbayar), '') as crbayar,
				isnull(rtrim(KRDH400.Noacdb), '') as noacdb,
				isnull(KRD400.angsurke, 0) as angsurke,
				isnull(KRD400.nlpokok, 0) as nlpokok,
				isnull(KRD400.nlbunga, 0) as nlbunga,
				isnull(KRD400.nldenda, 0) as nldenda,
				isnull(KRD400.nlpinalty, 0) as nlpinalty
			`).
			Joins("LEFT JOIN KRD400 ON KRDH400.kdcab = KRD400.kdcab AND KRDH400.nofas = KRD400.nofas AND KRDH400.noref = KRD400.noref").
			Where("KRDH400.nofas = ?", Nofas)

		if Kdcab != "" {
			q = q.Where("KRDH400.kdcab = ?", Kdcab)
		}

		if err := q.Order("KRDH400.tgltrn desc, KRDH400.noref desc, KRD400.angsurke desc").Find(&rows).Error; err != nil {
			log.Printf("ERROR: Gagal mengambil riwayat pembayaran pinjaman: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		resData := make([]obj.ResRiwayatBayarItem, len(rows))
		for i, r := range rows {
			total := helper.Round2(r.NLPOKOK + r.NLBUNGA + r.NLDENDA + r.NLPINALTY)
			tglFormatted := r.TGLTRN.Format(helper.DateTimeFormat)

			resData[i] = obj.ResRiwayatBayarItem{
				NOREF:        r.NOREF,
				TGLTRN:       tglFormatted,
				STSTRN:       r.STSTRN,
				STSTRN_DESC:  helper.FormatStsTrn(r.STSTRN),
				KET:          r.KET,
				CRBAYAR:      r.CRBAYAR,
				CRBAYAR_DESC: helper.FormatCrBayar(r.CRBAYAR),
				NOACDB:       r.NOACDB,
				ANGSURKE:     r.ANGSURKE,
				NLPOKOK:      r.NLPOKOK,
				NLBUNGA:      r.NLBUNGA,
				NLDENDA:      r.NLDENDA,
				NLPINALTY:    r.NLPINALTY,
				TOTAL:        total,
			}
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan riwayat pembayaran pinjaman", resData)
	}
}

