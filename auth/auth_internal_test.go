package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Tobias-Kusuma/core-ksp-api/cache"
	"github.com/Tobias-Kusuma/core-ksp-api/obj"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateTokensPreservesSessionWithoutRedis(t *testing.T) {
	oldKey := jwtKey
	oldRedis := cache.Redis
	jwtKey = []byte("session-test-key")
	// Generator tidak boleh membaca atau membuat ulang sesi, termasuk setelah logout.
	cache.Redis = nil
	t.Cleanup(func() {
		jwtKey = oldKey
		cache.Redis = oldRedis
	})

	for _, tokenType := range []string{"access", "refresh"} {
		t.Run(tokenType, func(t *testing.T) {
			generate := GenerateAccessToken
			if tokenType == "refresh" {
				generate = GenerateRefreshToken
			}
			if _, err := generate("001", "123", "BUDI", ""); err == nil {
				t.Fatal("token dengan session kosong harus ditolak")
			}
			token, err := generate("001", "123", "BUDI", "session-login")
			if err != nil {
				t.Fatal(err)
			}
			claims, err := ValidateToken(token)
			if err != nil {
				t.Fatal(err)
			}
			if claims.SESSIONID != "session-login" || claims.TOKENTYPE != tokenType {
				t.Fatalf("claims sesi atau tipe token tidak sesuai: %+v", claims)
			}
		})
	}
}

func TestRejectTokenWithoutSession(t *testing.T) {
	oldKey := jwtKey
	jwtKey = []byte("session-test-key")
	t.Cleanup(func() { jwtKey = oldKey })

	for _, tokenType := range []string{"access", "refresh"} {
		claims := obj.CustomClaims{
			KDCAB: "001", CNO: "123", CNM: "BUDI", TOKENTYPE: tokenType,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			},
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ValidateToken(token); err == nil {
			t.Fatalf("%s token tanpa session harus ditolak", tokenType)
		}
		if tokenType == "access" {
			router := gin.New()
			router.GET("/protected", JwtAuth(), func(c *gin.Context) {
				t.Error("handler tidak boleh dipanggil untuk token tanpa session")
			})
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, ingin 401", response.Code)
			}
		}
	}
}
