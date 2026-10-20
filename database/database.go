package database

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// main connection string
func connString(prefix string) (string, error) {
	dbuser := os.Getenv(prefix + "_DBUSER")
	dbpassword := os.Getenv(prefix + "_DBPASSWORD")
	dbserver := os.Getenv(prefix + "_DBSERVER")
	dbport := os.Getenv(prefix + "_DBPORT")
	dbname := os.Getenv(prefix + "_DBNAME")

	if dbuser == "" || dbpassword == "" || dbserver == "" || dbport == "" || dbname == "" {
		return "", fmt.Errorf("env database %s belum lengkap", prefix)
	}

	// URL-encode username and password to handle special characters
	encodedUser := url.QueryEscape(dbuser)
	encodedPassword := url.QueryEscape(dbpassword)

	// full connection string
	destination := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s&encrypt=disable",
		encodedUser, encodedPassword, dbserver, dbport, dbname,
	)

	return destination, nil
}

type Pools struct {
	Main *gorm.DB
}

var pools Pools

func openDb(conn string, maxOpen, maxIdle int, life time.Duration) (*gorm.DB, error) {
	gormdb, err := gorm.Open(sqlserver.Open(conn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormdb.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(life)

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return gormdb, nil
}

func InitDB() error {
	// get all connection string
	mainConn, err := connString("MAIN")
	if err != nil {
		return fmt.Errorf("ERROR: Gagal mendapatkan main db connection string, dengan detail: %w\n", err)
	}

	// docConn, err := connString("DOC")
	// if err != nil {
	// 	return fmt.Errorf("ERROR: Gagal mendapatkan doc db connection string, dengan detail: %w\n", err)
	// }

	// logConn, err := connString("LOG")
	// if err != nil {
	// 	return fmt.Errorf("ERROR: Gagal mendapatkan log db connection string, dengan detail: %w\n", err)
	// }

	// open all connection
	pools.Main, err = openDb(mainConn, 200, 75, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("ERROR: Gagal terkoneksi dengan mobile database, dengan detail: %w\n", err)
	}

	// pools.Doc, err = openDb(docConn, 20, 10, 30*time.Minute)
	// if err != nil {
	// 	return fmt.Errorf("ERROR: Gagal terkoneksi dengan doc database, dengan detail: %w\n", err)
	// }

	// pools.Log, err = openDb(logConn, 30, 15, 30*time.Minute)
	// if err != nil {
	// 	return fmt.Errorf("ERROR: Gagal terkoneksi dengan log database, dengan detail: %w\n", err)
	// }

	return nil
}

func Main() *gorm.DB {
	if pools.Main == nil {
		panic("main db is nil: call db.InitDB() and ensure it returns nil error")
	}

	return pools.Main
}

// func Doc() *gorm.DB {
// 	if pools.Doc == nil {
// 		panic("doc db is nil: call db.InitDB() and ensure it returns nil error")
// 	}

// 	return pools.Doc
// }

// func Log() *gorm.DB {
// 	if pools.Log == nil {
// 		panic("log db is nil: call db.InitDB() and ensure it returns nil error")
// 	}

// 	return pools.Log
// }
