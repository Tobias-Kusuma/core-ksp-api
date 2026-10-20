package obj

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// struct untuk log query
type LogHis struct {
	DUPD     time.Time `gorm:"column:dupd"`
	USRID    string    `gorm:"column:usrid"`
	MODUL    string    `gorm:"column:modul"`
	FUNCNAME string    `gorm:"column:funcname"`
	STATUS   string    `gorm:"column:status"`
	LOGQUERY string    `gorm:"column:logquery"`
	LOGERROR *string   `gorm:"column:logerror"`
}

// struct untuk oper context dan instance db yang telah dibuka;
// C: gin (session) context;
// Ctx: generik timeout and cancel context untuk goroutine/worker/dll;
// Mobile: db instance untuk mobile operation;
// Doc: db instance untuk document operation;
// Log: db instance untuk log operation;
type AppTx struct {
	C      *gin.Context    // session context
	Ctx    context.Context // generic context
	Mobile *gorm.DB        // transaksi data mobile
	Doc    *gorm.DB        // transaksi data file
	Log    *gorm.DB        // transaksi log trail
}

// struct claim untuk otentikasi jwt
type CustomClaims struct {
	DEVICEID  string `json:"deviceid"`
	NOHP      string `json:"nohp"`
	ID        string `json:"id"`
	KDCAB     string `json:"kdcab"`
	CNO       string `json:"cno"`
	NAMA      string `json:"nama"`
	TOKENTYPE string `json:"token_type"`
	jwt.RegisteredClaims
}

type DetailCust struct {
	DEVICEID string `json:"deviceid" gorm:"column:deviceid"`
	NOHP     string `json:"nohp" gorm:"column:nohp"`
	ID       string `json:"id" gorm:"column:id"`
	KDCAB    string `json:"kdcab" gorm:"column:kdcab"`
	CNO      string `json:"cno" gorm:"column:cno"`
	NAMA     string `json:"nama" gorm:"column:nama"`
}
