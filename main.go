package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tobias-Kusuma/core-ksp-api/auth"
	"github.com/Tobias-Kusuma/core-ksp-api/cache"
	db "github.com/Tobias-Kusuma/core-ksp-api/database"
	"github.com/Tobias-Kusuma/core-ksp-api/helper"
	"github.com/Tobias-Kusuma/core-ksp-api/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// load konfigurasi dari file .env jika ada
	if err := godotenv.Load(); err != nil {
		log.Println("📄 Notice: File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	// inisiasi redis client dan cek koneksi
	cache.InitRedisClient()
	if err := cache.PingRedis(3 * time.Second); err != nil {
		log.Printf("⚠️ WARNING: Gagal terhubung ke Redis (%s): %s", cache.Redis.Options().Addr, err)
	} else {
		log.Printf("✅ Terhubung ke Redis di %s", cache.Redis.Options().Addr)
	}

	// buka koneksi dengan database
	if err := db.InitDB(); err != nil {
		log.Fatal(err)
	}

	// inisiasi PORT
	port := os.Getenv("PORT")

	// set secret key & cred lainnya
	auth.SetJwtKey()
	auth.SetSignSecret()

	// inisiasi router
	router := gin.New()

	// tambah middleware
	router.Use(gin.Logger(), helper.ResponseLogger(), gin.Recovery(), helper.RequestLogger(), helper.ContextCleaner())

	// Tulis default error di dalam middleware agar body response ikut tercatat.
	router.NoRoute(func(c *gin.Context) { c.String(http.StatusNotFound, "404 page not found") })
	router.NoMethod(func(c *gin.Context) { c.String(http.StatusMethodNotAllowed, "405 method not allowed") })

	// routes yang dipakai
	routes.DefaultRoutes(router)
	routes.AnggotaRoute(router)
	routes.SimpananRoute(router)
	routes.PinjamanRoute(router)

	log.Printf("%s\n", "✅ All database connections have been successfully initialized")

	if port == "" {
		port = ":8080"
	}

	server := &http.Server{
		Addr:              port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// jalankan server HTTP dengan router Gin sebagai handler, jadi tidak perlu router.Run().
	// gunakan goroutine agar proses utama tetap bisa menunggu sinyal shutdown.
	// buffer 1 menampung hasil server meskipun proses utama sedang menangani shutdown.
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()

	log.Printf("%s\n", "✅ API Apps starting ...")

	// tunggu sampai server berhenti atau aplikasi menerima sinyal Ctrl+C / SIGTERM.
	select {
	case err := <-serverErr:
		// catat error server; ErrServerClosed menandakan server ditutup secara normal.
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("ERROR: HTTP server berhenti dengan error: %s", err)
		}

	case <-stopCtx.Done():
		// hentikan penerimaan koneksi baru dan tunggu request aktif selesai, maksimal 20 detik.
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("ERROR: Gagal shutdown HTTP server: %s", err)

			// paksa tutup koneksi jika graceful shutdown gagal atau melewati batas waktu.
			_ = server.Close()
		}

		cancel()
	}

	// tunggu antrean log HTTP selesai diproses, maksimal 10 detik sebelum aplikasi keluar.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := helper.FlushHTTPLogs(ctx); err != nil {
		log.Printf("ERROR: Antrean log HTTP belum selesai saat shutdown: %s", err)
	}

	// tutup koneksi Redis setelah proses penyelesaian log HTTP berakhir.
	_ = cache.CloseRedis()
}
