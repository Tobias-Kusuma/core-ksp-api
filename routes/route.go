package routes

import (
	"fmt"

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
