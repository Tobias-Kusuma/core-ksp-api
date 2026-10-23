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
	db "github.com/Tobias-Kusuma/core-ksp-api/database"
	"github.com/Tobias-Kusuma/core-ksp-api/helper"
	"github.com/Tobias-Kusuma/core-ksp-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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
	router.Use(gin.Logger(), helper.ResponseLogger(), gin.Recovery(), helper.RequestLogger(), helper.ContextCleaner())
	// Tulis default error di dalam middleware agar body response ikut tercatat.
	router.NoRoute(func(c *gin.Context) { c.String(http.StatusNotFound, "404 page not found") })
	router.NoMethod(func(c *gin.Context) { c.String(http.StatusMethodNotAllowed, "405 method not allowed") })

	// routes yang dipakai
	routes.DefaultRoutes(router)
	routes.AnggotaRoute(router)

	log.Printf("%s\n", "✅ API Apps starting ...")
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

	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("ERROR: HTTP server berhenti dengan error: %s", err)
		}
	case <-stopCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("ERROR: Gagal shutdown HTTP server: %s", err)
			_ = server.Close()
		}
		cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := helper.FlushHTTPLogs(ctx); err != nil {
		log.Printf("ERROR: Antrean log HTTP belum selesai saat shutdown: %s", err)
	}
}
