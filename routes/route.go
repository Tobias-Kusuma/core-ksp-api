package routes

import (
	"fmt"

	"github.com/Tobias-Kusuma/core-ksp-api/controller"
	db "github.com/Tobias-Kusuma/core-ksp-api/database"
	"github.com/Tobias-Kusuma/core-ksp-api/helper"
	"github.com/gin-gonic/gin"
)

/*
	SECURING ENDPOINT:
	1. lv 0 (public) -> tidak menggunakan middleware apapun
	2. lv 1 (lv 0 + otentikasi user) -> menggunakan JWT authentication
	3. lv 2 (lv 1 + otentikasi timestamp request) -> mengecek apakah request masih belum expired
	4. lv 3 (lv 2 + otentikasi request id + otentikasi body signature) -> cek apakah ada replayed atau modified request
*/

func DefaultRoutes(routerEngine *gin.Engine) {
	// terbuka untuk publik tanpa authentication dan authorization
	defaultGroup := routerEngine.Group("")
	{
		// Index route
		defaultGroup.GET("/", func(c *gin.Context) {
			helper.RespondMaskedOk(c, "Selamat datang di API Mobile - EasyBank", nil)
		})

		// Cek DB Health
		defaultGroup.GET("/db-health", func(c *gin.Context) {
			if db.Main() == nil {
				helper.RespondMaskedError(c, true, fmt.Errorf("Tidak berhasil terhubung dengan database"))
			} else {
				helper.RespondMaskedOk(c, "Berhasil terhubung dengan database", nil)
			}
		})
	}
}

func AnggotaRoute(routerEngine *gin.Engine) {
	// terbuka untuk publik tanpa authentication dan authorization
	anggotaGroup := routerEngine.Group("/anggota")
	{
		// pendaftaran anggota
		anggotaGroup.POST("/pendaftaran", controller.DaftarAnggotaBaru())

		// get daftar anggota
		anggotaGroup.POST("/getdaftaranggota", controller.GetDaftarAnggota())

		// get data anggota
		anggotaGroup.POST("/getdataanggota", controller.GetDataAnggota())
	}
}

func SimpananRoute(routerEngine *gin.Engine) {
	// terbuka untuk publik tanpa authentication dan authorization
	simpananGroup := routerEngine.Group("/simpanan")
	{
		// get produk simpanan
		simpananGroup.POST("/getproduksimpanan", controller.GetProdukSimpanan())

		// pembukaan rekening simpanan
		simpananGroup.POST("/newsimpanan", controller.NewSimpanan())

		// setor simpanan
		simpananGroup.POST("/setorsimpanan", controller.SetorSimpanan())

		// tarik simpanan
		simpananGroup.POST("/tariksimpanan", controller.TarikSimpanan())

		// transfer simpanan
		simpananGroup.POST("/transfersimpanan", controller.TransferSimpanan())

		// get mutasi simpanan
		simpananGroup.POST("/getmutasi", controller.GetMutasi())

		// get daftar rekening nasabah
		simpananGroup.POST("/getdaftarrekening", controller.GetDaftarRekening())

		// get data detail rekening
		simpananGroup.POST("/getdatarekening", controller.GetDataRekening())

		// penutupan rekening simpanan
		simpananGroup.POST("/tutuprekening", controller.TutupRekening())
	}
}

func PinjamanRoute(routerEngine *gin.Engine) {
	pinjamanGroup := routerEngine.Group("/pinjaman")
	{
		// get daftar produk pinjaman
		pinjamanGroup.POST("/getprodukpinjaman", controller.GetProdukPinjaman())

		// pinjaman baru
		pinjamanGroup.POST("/pinjamanbaru", controller.PinjamanBaru())

		// get jadwal angsuran
		pinjamanGroup.POST("/getjadwalangsuran", controller.GetJadwalAngsuran())

		// get daftar pinjaman nasabah
		pinjamanGroup.POST("/getdaftarpinjaman", controller.GetDaftarPinjaman())

		// get data detail pinjaman
		pinjamanGroup.POST("/getdatapinjaman", controller.GetDataPinjaman())

		// get tagihan angsuran berjalan & denda
		pinjamanGroup.POST("/gettagihan", controller.GetTagihan())

		// simulasi pinjaman / kalkulator kredit
		pinjamanGroup.POST("/simulasi", controller.SimulasiPinjaman())

		// pembayaran cicilan / angsuran pinjaman
		pinjamanGroup.POST("/bayarangsuran", controller.BayarAngsuran())

		// pelunasan pinjaman dipercepat
		pinjamanGroup.POST("/pelunasankredit", controller.PelunasanKredit())

		// get riwayat pembayaran angsuran & pelunasan
		pinjamanGroup.POST("/getriwayatbayar", controller.GetRiwayatBayar())
	}
}
