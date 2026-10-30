package auth

import (
	"bytes"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Tobias-Kusuma/core-ksp-api/cache"
	"github.com/Tobias-Kusuma/core-ksp-api/helper"
	"github.com/Tobias-Kusuma/core-ksp-api/obj"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// authentication session internal
var (
	jwtKey     []byte
	signSecret []byte
)

// generate access token
func GenerateAccessToken( /*deviceid string, phone string, userid string,*/ kdcab string, cno string, cnm string) (string, error) {
	claims := obj.CustomClaims{
		// DEVICEID:  deviceid,
		// NOHP:      phone,
		// ID:        userid,
		KDCAB:     kdcab,
		CNO:       cno,
		CNM:       cnm,
		TOKENTYPE: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * 7 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "github.com/Tobias-Kusuma/core-ksp-api",
			Subject:   cnm,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

// generate access token
func GenerateRefreshToken( /*deviceid string, phone string, userid string,*/ kdcab string, cno string, cnm string) (string, error) {
	claims := obj.CustomClaims{
		// DEVICEID:  deviceid,
		// NOHP:      phone,
		// ID:        userid,
		KDCAB:     kdcab,
		CNO:       cno,
		CNM:       cnm,
		TOKENTYPE: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * 21 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "github.com/Tobias-Kusuma/core-ksp-api",
			Subject:   cnm,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

// Otentikasi menggunakan JWT
// ============================================
func SetJwtKey() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Println("No .env file found")
	// }

	jwtKey = []byte(os.Getenv("JWTKEY"))
}

// validasi token string
func ValidateToken(tokenStr string) (*obj.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &obj.CustomClaims{}, func(token *jwt.Token) (any, error) {
		// pastikan dulu kalau algoritma sesuai
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}

		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*obj.CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// otentikasi token akses
func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// cek auth header
		if authHeader == "" {
			err := fmt.Errorf("Tidak ada header Authorization yang ditemukan")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		// cek prefix value
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			err := fmt.Errorf("Authorization format harus token Bearer")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		// ambil token
		tokenStr := strings.TrimPrefix(authHeader, bearerPrefix)

		// validasi token
		claims, err := ValidateToken(tokenStr)
		if err != nil {
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, fmt.Errorf("Token tidak valid!"))
			return
		}

		if claims.TOKENTYPE != "access" {
			err := fmt.Errorf("Invalid, bukan token dengan tipe akses")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		if /* claims.DEVICEID == "" || claims.NOHP == "" || claims.ID == "" || */ claims.KDCAB == "" || claims.CNO == "" || claims.CNM == "" {
			err := fmt.Errorf("Invalid, detail identitas yang dibutuhkan tidak lengkap")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		// simpan data user ke context biar bisa dipakai handler berikutnya
		user := &obj.DetailCust{
			// DEVICEID: claims.DEVICEID,
			// NOHP:     claims.NOHP,
			// ID:       claims.ID,
			KDCAB: claims.KDCAB,
			CNO:   claims.CNO,
			CNM:   claims.CNM,
		}

		c.Set("token-string", tokenStr) // digunakan untuk verifikasi signature
		c.Set("detail-cust", user)

		c.Next()
	}
}

// Otentikasi menggunakan timestamp
// ============================================
func TimestampAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tsHeader := c.GetHeader("X-Timestamp")

		if tsHeader == "" {
			err := fmt.Errorf("Tidak ada header Timestamp yang ditemukan")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		now := time.Now()
		reqTime, err := time.Parse(helper.RFC3339, tsHeader)
		if err != nil {
			err := fmt.Errorf("Timestamp harus dalam format RFC3339")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		// hitung selisih (absolute), kalau lebih dari 30 detik maka abort
		diff := now.Sub(reqTime)
		if diff < 0 {
			diff = -diff
		}

		if diff > (1 * 30 * time.Second) {
			err := fmt.Errorf("Timestamp sudah expired")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		c.Next()
	}
}

// Otentikasi Request ID
// ============================================
func RequestIdAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqIdHeader := c.GetHeader("X-Request-Id")

		if reqIdHeader == "" {
			err := fmt.Errorf("Tidak ada header Request Id yang ditemukan")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		// buat key untuk cek duplikasi di cache
		cacheKey := "request-id:" + reqIdHeader

		_, isFound, err := cache.RedisGet[string](cacheKey)
		if isFound && err == nil {
			err := fmt.Errorf("Terjadi duplikasi request!")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		cache.RedisSet(cacheKey, reqIdHeader, 1*5*time.Minute)

		c.Set("requestid", reqIdHeader)
		c.Next()
	}
}

// Otentikasi Body Signature
// ============================================
func SetSignSecret() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Println("No .env file found")
	// }

	signSecret = []byte(os.Getenv("SIGNSECRET"))
}

func SignatureAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		signAuth := c.GetHeader("X-Signature")

		if signAuth == "" {
			err := fmt.Errorf("Tidak ada header Signature yang ditemukan")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		// ambil dan baca body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, fmt.Errorf("Request body tidak dapat dibaca"))
			return
		}

		// harus set lagi ke request biar bisa dibaca handler
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// ambil token dri context
		tokenAny, exists := c.Get("token-string")
		if !exists {
			// override jadi string kosong misal tidak ada
			tokenAny = ""
		}

		// pastikan string
		token, ok := tokenAny.(string)
		if !ok {
			err := fmt.Errorf("Token akses JWT tidak valid")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		// ambil token dri context
		requestId, exists := c.Get("requestid")
		if !exists {
			// override jadi string kosong misal tidak ada
			requestId = ""
		}

		// pastikan string
		reqId, ok := requestId.(string)
		if !ok {
			err := fmt.Errorf("Token akses JWT tidak valid")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		// create string to sign (method + bodyhash + jwt access token + request id)
		bodyHash := helper.HashBody(bodyBytes)
		method := c.Request.Method
		stringToSign := fmt.Sprintf("%s:%s:%s:%s", method, bodyHash, token, reqId)

		// create expected signature
		expected := helper.GenerateSignature(stringToSign, signSecret)

		// // UNTUK DEBUGGING
		// log.Println("========== SIGNATURE DEBUG ==========")
		// log.Printf("raw body     : %q", string(bodyBytes))
		// log.Printf("method       : %s", method)
		// log.Printf("token        : %s", token)
		// log.Printf("body hash    : %s", bodyHash)
		// log.Printf("stringToSign : %s", stringToSign)
		// log.Printf("expected     : %s", expected)
		// log.Printf("header sign  : %s", signAuth)

		// log.Printf("signSecret raw : %q", string(signSecret))
		// log.Printf("signSecret len : %d", len(signSecret))
		// log.Println("====================================")

		// validate signature
		if subtle.ConstantTimeCompare([]byte(expected), []byte(signAuth)) != 1 {
			err := fmt.Errorf("Signature tidak valid")
			log.Printf("ERROR: %s", err)
			helper.RespondErrorUnauthorized(c, err)
			return
		}

		c.Next()
	}
}
