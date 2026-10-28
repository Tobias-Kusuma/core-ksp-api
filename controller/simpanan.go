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

func GetProdukSimpanan() gin.HandlerFunc {
	return func(c *gin.Context) {
		AppTx, err := helper.Get_DB_ReqBody(c, 30, nil)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		ResData := make([]obj.ResProdukSimpanan, 0)
		querySelect := `
			rtrim(jenis) as jenis,
			rtrim(nmproduk) as nmproduk,
			case rtrim(jnsprd)
				when '1' then 'Tabungan'
				when '3' then 'RAB'
				else rtrim(jnsprd)
			end as jnsprd,
			case rtrim(jnstipe)
				when '1' then 'Harian'
				when '3' then 'Berjangka'
				when '4' then 'Pokok'
				when '5' then 'Wajib'
				else rtrim(jnstipe)
			end as jnstipe,
			nlbiaya,
			minsetor,
			wjbsetor,
			minsld,
			maxtarik,
			biayatutup,
			penalty,
			pajak,
			rtrim(status) as status
		`

		if err := AppTx.Main.Debug().Table("RET500").
			Select(querySelect).
			Where("status = ?", "1").
			Order("jenis asc").
			Scan(&ResData).Error; err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan daftar produk simpanan", ResData)
	}
}

func NewSimpanan() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqBukaSimpanan
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

		// 1. Cek keberadaan anggota di MSTCST
		var anggota struct {
			KDCAB string  `gorm:"column:kdcab"`
			CNO   string  `gorm:"column:cno"`
			CNM   *string `gorm:"column:cnm"`
		}
		if err := AppTx.Main.Debug().Table("MSTCST").
			Select("kdcab, cno, cnm").
			Where("kdcab = ? and cno = ?", Kdcab, Cno).
			First(&anggota).Error; err != nil {
			log.Printf("ERROR: Data anggota tidak ditemukan: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Data anggota tidak ditemukan"))
			return
		}

		NamaAnggota := ""
		if anggota.CNM != nil {
			NamaAnggota = strings.TrimSpace(*anggota.CNM)
		}

		// 2. Cek parameter produk simpanan di RET500
		var produk obj.RET500
		if err := AppTx.Main.Debug().Table("RET500").
			Where("jenis = ? and status = ?", Jenis, "1").
			First(&produk).Error; err != nil {
			log.Printf("ERROR: Produk simpanan tidak ditemukan atau tidak aktif: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Produk simpanan tidak ditemukan atau tidak aktif"))
			return
		}

		ByTutup := 0.00
		if produk.BIAYATUTUP != nil {
			ByTutup = *produk.BIAYATUTUP
		}

		MinSld := 0.00
		if produk.MINSLD != nil {
			MinSld = *produk.MINSLD
		}

		MaxTarik := 0.00
		if produk.MAXTARIK != nil {
			MaxTarik = *produk.MAXTARIK
		}

		NlSetor := ReqBody.NLSETOR
		if NlSetor <= 0 && produk.WJBSETOR != nil {
			NlSetor = *produk.WJBSETOR
		}

		// 3. Generate nomor rekening baru (noac):
		var last_noac int64
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Select("isnull(max(try_cast(right(noac, 5) as bigint)), 0) as last_noac").
			Where("kdcab = ? and jenis = ?", Kdcab, Jenis).
			Scan(&last_noac).Error; err != nil {
			log.Printf("ERROR: Gagal membuat nomor rekening: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Noac := fmt.Sprintf("%s%s%05d", Kdcab, Jenis, last_noac+1)

		NolRupiah := 0.00
		TglNow := time.Now()

		newSimpanan := obj.SHDBLC{
			KDCAB:    Kdcab,
			NOAC:     Noac,
			CNO:      Cno,
			ALIASNM:  &NamaAnggota,
			JENIS:    &Jenis,
			BLOKIR:   helper.StringPtr("N"),
			NLHOLD:   &NolRupiah,
			SLDAWAL:  &NolRupiah,
			SLDAKHIR: &NolRupiah,
			SLDPAKAI: &NolRupiah,
			MINSLD:   &MinSld,
			MAXTARIK: &MaxTarik,
			NLBIAYA:  &ByTutup,
			MUTDB:    &NolRupiah,
			MUTCR:    &NolRupiah,
			NLSETOR:  &NlSetor,
			TGLAWAL:  &TglNow,
			AUTODB:   helper.StringPtr("N"),
			KODEAO:   helper.StringPtr("0000"),
			STATUS:   helper.StringPtr("1"),
		}

		if err := AppTx.Main.Debug().Table("SHDBLC").Create(&newSimpanan).Error; err != nil {
			log.Printf("ERROR: Gagal membuka rekening simpanan: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		ResData := obj.ResNewSimpanan{
			NAMA: NamaAnggota,
			NOAC: Noac,
		}

		helper.RespondMaskedOk(c, "Berhasil membuka rekening simpanan baru", ResData)
	}
}

func SetorSimpanan() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqSetorSimpanan
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Noac := strings.TrimSpace(ReqBody.NOAC)
		Kdcab := strings.TrimSpace(ReqBody.KDCAB)

		if Noac == "" || Kdcab == "" || ReqBody.NOMINAL <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter kdcab, noac, dan nominal > 0 wajib diisi"))
			return
		}

		var rekening obj.SHDBLC
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("kdcab = ? and noac = ? and status = ?", Kdcab, Noac, "1").
			First(&rekening).Error; err != nil {
			log.Printf("ERROR: Rekening tidak ditemukan atau tidak aktif: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Rekening tidak ditemukan atau tidak aktif"))
			return
		}

		sldAkhir := 0.0
		if rekening.SLDAKHIR != nil {
			sldAkhir = *rekening.SLDAKHIR
		}
		newSldAkhir := sldAkhir + ReqBody.NOMINAL

		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("kdcab = ? and noac = ?", Kdcab, Noac).
			Updates(map[string]interface{}{
				"sldakhir": newSldAkhir,
				"sldpakai": newSldAkhir,
			}).Error; err != nil {
			log.Printf("ERROR: Gagal memproses setor simpanan: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		aliasNm := ""
		if rekening.ALIASNM != nil {
			aliasNm = *rekening.ALIASNM
		}

		ResData := obj.ResSetorSimpanan{
			KDCAB:      Kdcab,
			NOAC:       Noac,
			ALIASNM:    aliasNm,
			NOMINAL:    ReqBody.NOMINAL,
			SLDAKHIR:   newSldAkhir,
			SLDPAKAI:   newSldAkhir,
			KETERANGAN: ReqBody.KETERANGAN,
		}

		helper.RespondMaskedOk(c, "Berhasil melakukan setoran simpanan", ResData)
	}
}

func TarikSimpanan() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqTarikSimpanan
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Noac := strings.TrimSpace(ReqBody.NOAC)
		Kdcab := strings.TrimSpace(ReqBody.KDCAB)

		if Noac == "" || Kdcab == "" || ReqBody.NOMINAL <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter kdcab, noac, dan nominal > 0 wajib diisi"))
			return
		}

		var rekening obj.SHDBLC
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("kdcab = ? and noac = ? and status = ?", Kdcab, Noac, "1").
			First(&rekening).Error; err != nil {
			log.Printf("ERROR: Rekening tidak ditemukan atau tidak aktif: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Rekening tidak ditemukan atau tidak aktif"))
			return
		}

		sldPakai := 0.0
		if rekening.SLDPAKAI != nil {
			sldPakai = *rekening.SLDPAKAI
		}

		minSld := 0.0
		if rekening.MINSLD != nil {
			minSld = *rekening.MINSLD
		}

		if (sldPakai - ReqBody.NOMINAL) < minSld {
			helper.RespondMaskedError(c, true, fmt.Errorf("Saldo tidak mencukupi untuk penarikan"))
			return
		}

		newSldAkhir := sldPakai - ReqBody.NOMINAL

		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("kdcab = ? and noac = ?", Kdcab, Noac).
			Updates(map[string]interface{}{
				"sldakhir": newSldAkhir,
				"sldpakai": newSldAkhir,
			}).Error; err != nil {
			log.Printf("ERROR: Gagal memproses tarik simpanan: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		aliasNm := ""
		if rekening.ALIASNM != nil {
			aliasNm = *rekening.ALIASNM
		}

		ResData := obj.ResTarikSimpanan{
			KDCAB:      Kdcab,
			NOAC:       Noac,
			ALIASNM:    aliasNm,
			NOMINAL:    ReqBody.NOMINAL,
			SLDAKHIR:   newSldAkhir,
			SLDPAKAI:   newSldAkhir,
			KETERANGAN: ReqBody.KETERANGAN,
		}

		helper.RespondMaskedOk(c, "Berhasil melakukan penarikan simpanan", ResData)
	}
}

func TransferSimpanan() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqTransferSimpanan
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		NoacAsal := strings.TrimSpace(ReqBody.NOAC_ASAL)
		NoacTujuan := strings.TrimSpace(ReqBody.NOAC_TUJUAN)

		if NoacAsal == "" || NoacTujuan == "" || ReqBody.NOMINAL <= 0 {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter noac_asal, noac_tujuan, dan nominal > 0 wajib diisi"))
			return
		}

		if NoacAsal == NoacTujuan {
			helper.RespondMaskedError(c, true, fmt.Errorf("Rekening asal dan tujuan tidak boleh sama"))
			return
		}

		var rekAsal obj.SHDBLC
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("noac = ? and status = ?", NoacAsal, "1").
			First(&rekAsal).Error; err != nil {
			log.Printf("ERROR: Rekening asal tidak ditemukan atau tidak aktif: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Rekening asal tidak ditemukan atau tidak aktif"))
			return
		}

		var rekTujuan obj.SHDBLC
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("noac = ? and status = ?", NoacTujuan, "1").
			First(&rekTujuan).Error; err != nil {
			log.Printf("ERROR: Rekening tujuan tidak ditemukan atau tidak aktif: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Rekening tujuan tidak ditemukan atau tidak aktif"))
			return
		}

		sldAsal := 0.0
		if rekAsal.SLDPAKAI != nil {
			sldAsal = *rekAsal.SLDPAKAI
		}
		minSld := 0.0
		if rekAsal.MINSLD != nil {
			minSld = *rekAsal.MINSLD
		}

		if (sldAsal - ReqBody.NOMINAL) < minSld {
			helper.RespondMaskedError(c, true, fmt.Errorf("Saldo rekening asal tidak mencukupi"))
			return
		}

		trx := AppTx.Main.Begin()

		newSldAsal := sldAsal - ReqBody.NOMINAL
		if err := trx.Table("SHDBLC").Where("noac = ?", NoacAsal).
			Updates(map[string]interface{}{
				"sldakhir": newSldAsal,
				"sldpakai": newSldAsal,
			}).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal update saldo rekening asal: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		sldTujuan := 0.0
		if rekTujuan.SLDAKHIR != nil {
			sldTujuan = *rekTujuan.SLDAKHIR
		}
		newSldTujuan := sldTujuan + ReqBody.NOMINAL
		if err := trx.Table("SHDBLC").Where("noac = ?", NoacTujuan).
			Updates(map[string]interface{}{
				"sldakhir": newSldTujuan,
				"sldpakai": newSldTujuan,
			}).Error; err != nil {
			trx.Rollback()
			log.Printf("ERROR: Gagal update saldo rekening tujuan: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		if err := trx.Commit().Error; err != nil {
			log.Printf("ERROR: Gagal commit transaksi transfer: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		namaPengirim := ""
		if rekAsal.ALIASNM != nil {
			namaPengirim = *rekAsal.ALIASNM
		}
		namaPenerima := ""
		if rekTujuan.ALIASNM != nil {
			namaPenerima = *rekTujuan.ALIASNM
		}

		ResData := obj.ResTransferSimpanan{
			NOAC_ASAL:      NoacAsal,
			NAMA_PENGIRIM:  namaPengirim,
			NOAC_TUJUAN:    NoacTujuan,
			NAMA_PENERIMA:  namaPenerima,
			NOMINAL:        ReqBody.NOMINAL,
			SLDAKHIR_ASAL:  newSldAsal,
			SLDPAKAI_ASAL:  newSldAsal,
			KETERANGAN:     ReqBody.KETERANGAN,
			WAKTU_TRANSFER: time.Now().Format(helper.DateTimeFormat),
		}

		helper.RespondMaskedOk(c, "Berhasil melakukan transfer simpanan", ResData)
	}
}

func GetMutasi() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqGetMutasi
		_, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Noac := strings.TrimSpace(ReqBody.NOAC)
		if Noac == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter noac wajib diisi"))
			return
		}

		resMutasi := make([]obj.ResMutasiItem, 0)
		helper.RespondMaskedOk(c, "Berhasil mendapatkan mutasi simpanan", resMutasi)
	}
}

func GetDaftarRekening() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqDaftarRekening
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		Cno := strings.TrimSpace(ReqBody.CNO)

		if Kdcab == "" || Cno == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter kdcab dan cno wajib diisi"))
			return
		}

		var daftarRekening []obj.SHDBLC
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("kdcab = ? and cno = ?", Kdcab, Cno).
			Order("noac asc").
			Find(&daftarRekening).Error; err != nil {
			log.Printf("ERROR: Gagal mengambil daftar rekening: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan daftar rekening", daftarRekening)
	}
}

func GetDataRekening() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqDataRekening
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		Noac := strings.TrimSpace(ReqBody.NOAC)

		if Kdcab == "" || Noac == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter kdcab dan noac wajib diisi"))
			return
		}

		var resRekening obj.ResDataRekening
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Select("SHDBLC.*, rtrim(RET500.nmproduk) as nmproduk, rtrim(RET500.jnsprd) as jnsprd, rtrim(RET500.jnstipe) as jnstipe").
			Joins("LEFT JOIN RET500 ON SHDBLC.jenis = RET500.jenis").
			Where("SHDBLC.kdcab = ? and SHDBLC.noac = ?", Kdcab, Noac).
			First(&resRekening).Error; err != nil {
			log.Printf("ERROR: Data rekening tidak ditemukan: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Data rekening tidak ditemukan"))
			return
		}

		helper.RespondMaskedOk(c, "Berhasil mendapatkan data rekening", resRekening)
	}
}

func TutupRekening() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ReqBody obj.ReqTutupRekening
		AppTx, err := helper.Get_DB_ReqBody(c, 30, &ReqBody)
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		Kdcab := strings.TrimSpace(ReqBody.KDCAB)
		Noac := strings.TrimSpace(ReqBody.NOAC)

		if Kdcab == "" || Noac == "" {
			helper.RespondMaskedError(c, true, fmt.Errorf("Parameter kdcab dan noac wajib diisi"))
			return
		}

		var rekening obj.SHDBLC
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("kdcab = ? and noac = ? and status = ?", Kdcab, Noac, "1").
			First(&rekening).Error; err != nil {
			log.Printf("ERROR: Rekening tidak ditemukan atau sudah ditutup: %s", err.Error())
			helper.RespondMaskedError(c, true, fmt.Errorf("Rekening tidak ditemukan atau sudah ditutup"))
			return
		}

		saldoAkhir := 0.0
		if rekening.SLDAKHIR != nil {
			saldoAkhir = *rekening.SLDAKHIR
		}
		biayaTutup := 0.0
		if rekening.NLBIAYA != nil {
			biayaTutup = *rekening.NLBIAYA
		}
		danaDicairkan := saldoAkhir - biayaTutup
		if danaDicairkan < 0 {
			danaDicairkan = 0
		}

		now := time.Now()
		if err := AppTx.Main.Debug().Table("SHDBLC").
			Where("kdcab = ? and noac = ?", Kdcab, Noac).
			Updates(map[string]interface{}{
				"status":   "0",
				"tgltutup": now,
				"sldakhir": 0,
				"sldpakai": 0,
			}).Error; err != nil {
			log.Printf("ERROR: Gagal menutup rekening: %s", err.Error())
			helper.RespondMaskedError(c, false, err)
			return
		}

		aliasNm := ""
		if rekening.ALIASNM != nil {
			aliasNm = *rekening.ALIASNM
		}

		res := obj.ResTutupRekening{
			NOAC:           Noac,
			ALIASNM:        aliasNm,
			SALDO_TERAKHIR: saldoAkhir,
			BIAYA_TUTUP:    biayaTutup,
			DANA_DICAIRKAN: danaDicairkan,
			NOAC_TUJUAN:    ReqBody.NOAC_TUJUAN,
			TGL_TUTUP:      now.Format(helper.DateFormat),
			STATUS:         "TUTUP",
		}

		helper.RespondMaskedOk(c, "Berhasil menutup rekening simpanan", res)
	}
}
