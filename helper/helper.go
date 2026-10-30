package helper

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"net/http"
	"net/mail"
	"net/url"

	db "github.com/Tobias-Kusuma/core-ksp-api/database"
	"github.com/Tobias-Kusuma/core-ksp-api/obj"
	"github.com/pariz/gountries"
	"golang.org/x/crypto/argon2"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"

	// go-mssqldb modern path:
	mssql "github.com/microsoft/go-mssqldb"
)

const (
	SQLServerParamLimit = 2100
	PostgresParamLimit  = 65535
	MySQLParamLimit     = 65000
	DateFormat          = "2006-01-02"
	DateTimeFormat      = "2006-01-02 15:04:05"
	RFC3339             = "2006-01-02T15:04:05-07:00"
)

const (
	argonMemory  uint32 = 64 * 1024 // 64 MB
	argonTime    uint32 = 2
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

var (
	columnCache sync.Map // map[reflect.Type]int
)

func getCallerOutsideHelper() string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(2, pcs) // skip runtime.Callers + callerOutsideHelper
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()

		// kalau sudah habis frame
		if frame.Function == "" && !more {
			return "unknown"
		}

		// buang frame runtime
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				return "unknown"
			}
			continue
		}

		// buang frame dari package/folder helper
		inHelperPkg := strings.Contains(frame.Function, "/helper.") || strings.Contains(frame.Function, "\\helper.")
		inHelperDir := strings.Contains(frame.File, string(filepath.Separator)+"helper"+string(filepath.Separator))

		if inHelperPkg || inHelperDir {
			if !more {
				return "unknown"
			}
			continue
		}

		// ketemu caller di luar helper
		return fmt.Sprintf("%s (%s:%d)", frame.Function, frame.File, frame.Line)
	}
}

// menghapus data dalam context begitu handler selesai memproses;
func ContextCleaner() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if cancel, ok := c.Get("cancelCtx"); ok {
				cancel.(context.CancelFunc)()
			}
		}()
		c.Next()
	}
}

// mengambil data user yang disimpan oleh middleware JwtAuth ke context;
func GetDetailCust(c *gin.Context) (*obj.DetailCust, error) {
	caller := getCallerOutsideHelper()

	if c == nil {
		return nil, fmt.Errorf("context tidak tersedia. Func called from %s", caller)
	}

	value, exists := c.Get("detail-cust")
	if !exists {
		return nil, fmt.Errorf("data user tidak ditemukan dalam context. Func called from %s", caller)
	}

	user, ok := value.(*obj.DetailCust)
	if !ok || user == nil {
		return nil, fmt.Errorf("data user dalam context tidak valid. Func called from %s", caller)
	}

	return user, nil
}

// mengembalikan instance dari session context/connection wrapper, dan parsing JSON ke instance reference;
// parms: context (mandatory), timeout dalam satuan detik, json body address-of operator (bisa nil);
func Get_DB_ReqBody(c *gin.Context, timeout int, ReqBody any) (*obj.AppTx, error) {
	caller := getCallerOutsideHelper()

	// reqbody bisa nil, namun kalau ada isi maka diparsing dari JSON menjadi struct instance
	if ReqBody != nil {
		if err := c.ShouldBindJSON(ReqBody); err != nil {
			return nil, fmt.Errorf("%s. Func called from %s", err.Error(), caller)
		}
	}

	// set timeout default
	if timeout <= 0 {
		timeout = 10
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(timeout)*time.Second)

	c.Set("cancelCtx", cancel)

	// prepare db (nanti akan pakai ctx yang sama)
	mainDb := db.Main()
	// docDb := db.Doc()
	// logDb := db.Log()

	// buat guard kalau koneksi ada yang rusak
	if mainDb == nil /*|| docDb == nil || logDb == nil*/ {
		cancel()
		return nil, fmt.Errorf("db pools not initialized (call db.InitDB() first). Func called from %s", caller)
	}

	return &obj.AppTx{
		C:    c,
		Ctx:  ctx,
		Main: mainDb.WithContext(ctx),
		// Doc:    docDb.WithContext(ctx),
		// Log:    logDb.WithContext(ctx),
	}, nil
}

// Fungsi replace untuk mengganti substring di dalam string
func Replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

// untuk handling string pointer
func StringPtr(s string) *string {
	return &s
}

// untuk handling float64 pointer
func Float64Ptr(f float64) *float64 {
	return &f
}

// untuk handling int64 pointer
func Int64Ptr(i int64) *int64 {
	return &i
}

// untuk detect type data
func escapeValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}

	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", val.Float())

	case reflect.Bool:
		if val.Bool() {
			return "1"
		}
		return "0"

	case reflect.String:
		escaped := strings.ReplaceAll(val.String(), "'", "''")
		return fmt.Sprintf("'%s'", escaped)

	default:
		// Cek apakah time.Time
		if t, ok := v.(time.Time); ok {
			return fmt.Sprintf("'%s'", t.Format("2006-01-02 15:04:05"))
		}
		// Fallback sebagai string
		escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

// FUNGSI-FUNGSI UNTUK KEPERLUAN LOG QUERY
// ============================================================

// format value sesuai tipe: string pakai quote, number tidak, bool -> 1/0
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		// escape single quote
		return "'" + strings.ReplaceAll(val, "'", "''") + "'"
	case fmt.Stringer:
		s := val.String()
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return fmt.Sprint(val)
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		// fallback: pakai string + quote
		s := fmt.Sprint(val)
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}
}

// parse "amount >=" -> ("amount", ">="), "status" -> ("status", "=")
func parseColumnAndOp(key string) (string, string) {
	parts := strings.Fields(key)
	if len(parts) == 0 {
		return "", "="
	}
	if len(parts) == 1 {
		return parts[0], "="
	}
	op := strings.ToLower(parts[len(parts)-1])
	col := strings.Join(parts[:len(parts)-1], " ")
	return col, op
}

// parseGormColumnTag ambil nama kolom dari tag gorm, contoh:
// `gorm:"column:nofas;primaryKey"` -> "nofas"
func parseGormColumnTag(tag string) string {
	if tag == "" {
		return ""
	}

	parts := strings.Split(tag, ";")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "column:") {
			return strings.TrimPrefix(p, "column:")
		}
	}

	return ""
}

// where clause
func buildWhereClause(where map[string]interface{}) string {
	var conditions []string
	for key, val := range where {
		col, op := parseColumnAndOp(key)

		// handle slice/array => IN (...)
		rv := reflect.ValueOf(val)
		if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
			// kalau user belum nulis "in", kita paksa jadi in
			if op == "=" {
				op = "in"
			}
			var items []string
			for i := 0; i < rv.Len(); i++ {
				items = append(items, formatValue(rv.Index(i).Interface()))
			}
			conditions = append(conditions,
				fmt.Sprintf("%s %s (%s)", col, op, strings.Join(items, ", ")),
			)
			continue
		}

		// handle nil => is null
		if val == nil {
			conditions = append(conditions, fmt.Sprintf("%s is null", col))
			continue
		}

		// operator biasa (=, >, <, >=, <=, !=, <>, like, dll.)
		conditions = append(conditions,
			fmt.Sprintf("%s %s %s", col, op, formatValue(val)),
		)
	}
	return strings.Join(conditions, " AND ")
}

// insert query dengan struct
func BuildInsertQueryFromStruct(table string, model interface{}) string {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	var fields, values []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !v.Field(i).CanInterface() {
			continue
		}
		fields = append(fields, field.Name)
		values = append(values, fmt.Sprintf("'%v'", v.Field(i).Interface()))
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(fields, ", "), strings.Join(values, ", "))
}

// update query dengan struct
func BuildUpdateQueryFromStruct(table string, where map[string]interface{}, model interface{}) string {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	var sets []string
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !v.Field(i).CanInterface() {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = '%v'", field.Name, v.Field(i).Interface()))
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(sets, ", "), buildWhereClause(where))
}

// delete query dengan struct
func BuildDeleteQueryFromStruct(table string, where map[string]interface{}) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", table, buildWhereClause(where))
}

// insert batches query dengan struct
func BuildInsertQueryFromStructBatches(table string, slice any) (string, error) {
	v := reflect.ValueOf(slice)
	t := reflect.TypeOf(slice)

	// harus slice
	if t.Kind() != reflect.Slice {
		return "", errors.New("data must be a slice of struct")
	}
	if v.Len() == 0 {
		return "", nil
	}

	elemType := t.Elem()
	// handle []*Struct
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	if elemType.Kind() != reflect.Struct {
		return "", errors.New("data must be a slice of struct")
	}

	var fields []string
	var fieldIndexes []int

	// ambil daftar kolom sekali saja dari tipe struct
	for i := 0; i < elemType.NumField(); i++ {
		sf := elemType.Field(i)

		// skip unexported field
		if sf.PkgPath != "" {
			continue
		}

		// prioritas: gorm:"column:xxx"
		colName := parseGormColumnTag(sf.Tag.Get("gorm"))

		// kalau kosong, fallback ke json:"xxx"
		if colName == "" {
			jsonTag := sf.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				colName = strings.Split(jsonTag, ",")[0]
			}
		}

		// kalau masih kosong, fallback ke nama field (lowercase)
		if colName == "" {
			colName = strings.ToLower(sf.Name)
		}

		fields = append(fields, colName)
		fieldIndexes = append(fieldIndexes, i)
	}

	if len(fields) == 0 {
		return "", errors.New("no exported fields to build query")
	}

	var rows []string

	// loop tiap item di slice
	for i := 0; i < v.Len(); i++ {
		itemVal := v.Index(i)

		// handle pointer ke struct
		if itemVal.Kind() == reflect.Ptr {
			if itemVal.IsNil() {
				return "", fmt.Errorf("nil struct at index %d", i)
			}
			itemVal = itemVal.Elem()
		}

		var values []string
		for _, fi := range fieldIndexes {
			fv := itemVal.Field(fi)

			var iv any
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					iv = nil
				} else {
					iv = fv.Elem().Interface()
				}
			} else {
				iv = fv.Interface()
			}

			values = append(values, escapeValue(iv))
		}

		rows = append(rows, fmt.Sprintf("(%s)", strings.Join(values, ", ")))
	}

	query := fmt.Sprintf(
		"insert into %s (%s)\nvalues\n%s",
		table,
		strings.Join(fields, ", "),
		strings.Join(rows, ",\n"),
	)

	return query, nil
}

// insert batches query dengan map
func BuildInsertQueryFromMapBatches(table string, data []map[string]interface{}) string {
	if len(data) == 0 {
		return ""
	}

	var fields []string
	for k := range data[0] {
		fields = append(fields, k)
	}

	var rows []string
	for _, row := range data {
		var values []string
		for _, field := range fields {
			v := row[field]
			values = append(values, escapeValue(v))
		}
		rows = append(rows, fmt.Sprintf("(%s)", strings.Join(values, ", ")))
	}

	return fmt.Sprintf(
		"insert into %s (%s)\nvalues\n%s",
		table,
		strings.Join(fields, ", "),
		strings.Join(rows, ",\n"),
	)
}

// insert query dengan map
func BuildInsertQueryFromMap(table string, data map[string]interface{}) string {
	var fields, values []string
	for k, v := range data {
		fields = append(fields, k)
		values = append(values, fmt.Sprintf("'%v'", v))
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(fields, ", "), strings.Join(values, ", "))
}

// update query dengan map
func BuildUpdateQueryFromMap(table string, where, updates map[string]interface{}) string {
	var sets []string
	for k, v := range updates {
		sets = append(sets, fmt.Sprintf("%s = '%v'", k, v))
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(sets, ", "), buildWhereClause(where))
}

// delete query dengan map
func BuildDeleteQueryFromMap(table string, where map[string]interface{}) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", table, buildWhereClause(where))
}

// fungsi insert log query
func LogHisQuery(db *gorm.DB, usrid string, modul string, funcname string, logquery string, logerror *string) {
	caller := getCallerOutsideHelper()

	ReqLogHis := obj.LogHis{
		DUPD:     time.Now(),
		USRID:    usrid,
		MODUL:    modul,
		FUNCNAME: funcname,
		LOGQUERY: logquery,
		LOGERROR: logerror,
	}

	if (logerror == nil) || (len(*logerror) == 0) {
		ReqLogHis.STATUS = "1"
	} else {
		ReqLogHis.STATUS = "0"
	}

	// Gunakan session baru agar tidak ikut rollback
	go func() {
		err := db.Session(&gorm.Session{NewDB: true}).Table("LOGHIS").Create(&ReqLogHis).Error
		if err != nil {
			log.Printf("Gagal menyimpan audit log untuk fungsi: %s dengan detail error:%s\n", caller, err.Error())
		}
	}()
}

// Struct dan fungsi untuk masking error secara global
// ==========================================================
type MaskedError struct {
	Number  int    // e.g. 2601, 2627, dll
	Code    string // e.g. NOT_FOUND, DUPE_UNIQUE, DUPE_PK, TOO_LONG, BAD_CONVERSION, GENERAL
	Message string // user-facing
}

// implement interface
func (e MaskedError) Error() string { return e.Message }

// MaskSQLError mengubah error mentah menjadi MaskedError
// rowsAffected opsional untuk query yang tidak mengembalikan error ketika data kosong.
func MaskSQLError(err error, rowsAffected ...int64) error {
	if err == nil {
		if len(rowsAffected) > 0 && rowsAffected[0] == 0 {
			return MaskedError{Number: 0, Code: "NOT_FOUND", Message: "Data tidak ditemukan."}
		}
		return nil
	}

	// data tidak ditemukan (termasuk error yang dibungkus)
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		return MaskedError{Number: 0, Code: "NOT_FOUND", Message: "Data tidak ditemukan."}
	}

	// 1) shortcut: GORM duplicated-key (jaga-jaga untuk driver lain)
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return MaskedError{Number: 0, Code: "DUPE_UNIQUE", Message: "Data sudah ada, tidak boleh dimasukkan lagi (duplikat)."}
	}

	// 2) deteksi spesifik SQL Server (mssql.Error)
	var me mssql.Error
	if errors.As(err, &me) {
		switch me.Number {
		case 2601:
			// duplicate unique index
			return MaskedError{Number: int(me.Number), Code: "DUPE_UNIQUE", Message: "Data sudah ada dan tidak boleh duplikat."}
		case 2627:
			// bisa PK atau UNIQUE tergantung isi message
			msg := strings.ToUpper(me.Message)
			if strings.Contains(msg, "PRIMARY KEY") {
				return MaskedError{Number: int(me.Number), Code: "DUPE_PK", Message: "Data sudah ada dan tidak boleh duplikat."}
			}
			return MaskedError{Number: int(me.Number), Code: "DUPE_UNIQUE", Message: "Data sudah ada dan tidak boleh duplikat."}
		case 8152, 2628:
			// string/binary data truncated (lama & baru)
			return MaskedError{Number: int(me.Number), Code: "TOO_LONG", Message: "Panjang karakter yang dimasukkan terlalu panjang, mohon periksa kembali."}
		case 241, 242, 245:
			// conversion/date parsing
			return MaskedError{Number: int(me.Number), Code: "BAD_CONVERSION", Message: "Format data tidak sesuai. Periksa kembali data yang dimasukkan."}
		case 8115:
			// conversion/date parsing
			return MaskedError{Number: int(me.Number), Code: "BAD_CONVERSION", Message: "Pengolahan data numerik invalid, mohon periksa kembali."}
		default:
			// fallback untuk kode lain
			return MaskedError{Number: int(me.Number), Code: "GENERAL", Message: "Terjadi kesalahan saat memproses data."}
		}
	}

	// 3) fallback berbasis teks (kalau bukan mssql.Error, atau lewat layer lain)
	up := strings.ToUpper(err.Error())

	// not found / 0 records
	if strings.Contains(up, "RECORD NOT FOUND") ||
		strings.Contains(up, "NO ROWS IN RESULT SET") ||
		strings.Contains(up, "0 RECORDS") {
		return MaskedError{Number: 0, Code: "NOT_FOUND", Message: "Data tidak ditemukan."}
	}

	// duplicate
	if strings.Contains(up, "DUPLICATE KEY") || strings.Contains(up, "UNIQUE") {
		if strings.Contains(up, "PRIMARY KEY") {
			return MaskedError{Number: 0, Code: "DUPE_PK", Message: "Data sudah ada dan tidak boleh duplikat."}
		}
		return MaskedError{Number: 0, Code: "DUPE_UNIQUE", Message: "Data sudah ada dan tidak boleh duplikat."}
	}

	// too long
	if strings.Contains(up, "TRUNCAT") || strings.Contains(up, "VALUE TOO LONG") {
		return MaskedError{Number: 0, Code: "TOO_LONG", Message: "Panjang karakter yang dimasukkan terlalu panjang, mohon periksa kembali."}
	}

	// conversion
	if strings.Contains(up, "CONVERSION FAILED") ||
		strings.Contains(up, "INVALID INPUT SYNTAX") ||
		strings.Contains(up, "FAILED TO CONVERT") ||
		strings.Contains(up, "ERROR CONVERTING DATA TYPE") ||
		strings.Contains(up, "DATE AND/OR TIME") {
		return MaskedError{Number: 0, Code: "BAD_CONVERSION", Message: "Format data tidak sesuai. Periksa kembali nilai input."}
	}

	// general
	return MaskedError{Number: 0, Code: "GENERAL", Message: "Terjadi kesalahan saat memproses data."}
}

// masking respon error secara general, kalau bukan dari sistem (custom message) maka set custom = TRUE;
func RespondMaskedError(c *gin.Context, custom bool, err error) {
	if custom {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})

	} else {
		if me, ok := MaskSQLError(err).(MaskedError); ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": me.Message})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Terjadi kesalahan saat memproses."})
		}
	}
}

// masking respon error 401 (unauthorized)
func RespondErrorUnauthorized(c *gin.Context, err error) {
	ResBody := map[string]interface{}{
		"message": err.Error(),
	}

	c.AbortWithStatusJSON(http.StatusUnauthorized, ResBody)
}

// helper function untuk standarisasi respon OK;
// =============================================
func RespondMaskedOk(c *gin.Context, message any, data interface{}) {
	ResBody := map[string]interface{}{
		"message": message,
		"data":    data,
	}

	c.JSON(http.StatusOK, ResBody)
}

// function untuk mendapatkan uuid v7
// =============================================
func mustV7() uuid.UUID {
	caller := getCallerOutsideHelper()

	u, err := uuid.NewV7()
	if err != nil {
		log.Fatalf("Gagal generate uuid v7 ketika dipanggil oleh: %s dengan detil error: %v", caller, err)
	}
	return u
}

func GetUUIDString() string {
	return mustV7().String()
}

// buat catat dan debug waktu eksekusi
// =============================================
type StepTimer struct {
	start time.Time
	last  time.Time
	steps []string
}

func NewStepTimer() *StepTimer {
	now := time.Now()
	return &StepTimer{
		start: now,
		last:  now,
	}
}

func (t *StepTimer) Mark(name string) {
	now := time.Now()
	dur := now.Sub(t.last)
	t.steps = append(t.steps, fmt.Sprintf("%-35s : %v", name, dur))
	t.last = now
}

func (t *StepTimer) Summary() string {
	total := time.Since(t.start)

	var b strings.Builder
	b.WriteString("\n========== EXECUTION TIME SUMMARY ==========\n")
	for _, s := range t.steps {
		b.WriteString(s)
		b.WriteString("\n")
	}
	b.WriteString("--------------------------------------------\n")
	b.WriteString(fmt.Sprintf("%-35s : %v\n", "TOTAL", total))
	b.WriteString("============================================\n")

	return b.String()
}

// untuk dapatkan maksimum batches dalam sekali insert;
// paramLimit = limit parameter DB (SQL Server = 2100, Postgres = 65535, dll);
// return => jumlah batch;
func GetOptimalBatchSize(model any, paramLimit int) int {
	col := getColumnCount(model)

	// log.Printf("jumlah kolom: %s\n", strconv.Itoa(col))

	if col <= 0 {
		return 1
	}

	batch := paramLimit / col
	batch = batch - (batch / 10)

	if batch < 1 {
		return 1
	}

	return batch
}

func getColumnCount(model any) int {
	t := reflect.TypeOf(model)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return 0
	}

	// cek cache
	if v, ok := columnCache.Load(t); ok {
		return v.(int)
	}

	count := countFieldsRecursive(t)

	columnCache.Store(t, count)

	return count
}

func countFieldsRecursive(t reflect.Type) int {
	total := 0

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// skip unexported
		if f.PkgPath != "" {
			continue
		}

		gormTag := f.Tag.Get("gorm")

		// skip ignored
		if strings.Contains(gormTag, "-") {
			continue
		}

		// embedded struct
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			total += countFieldsRecursive(f.Type)
			continue
		}

		total++
	}

	return total
}

// kebutuhan signature request;
// pertama hash body, lalu generate signature;
func HashBody(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

func GenerateSignature(stringToSign string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

// generate OTP secara random dengan memakai rand;
func GenerateOTP(length int) (string, error) {
	const digits = "19511017"

	otp := make([]byte, length)

	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}

	return string(otp), nil
}

// normalisasi nomor HP:
var phoneRegex = regexp.MustCompile(`[^0-9]`)

// trim spasi depan/belakang;
// hapus semua karakter selain angka;
func NormalizePhoneNumberString(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = phoneRegex.ReplaceAllString(phone, "")
	return phone
}

func NormalizePhoneNumber(phone string) (int, error) {
	return strconv.Atoi(NormalizePhoneNumberString(phone))
}

// untuk encoding sesuai requirement dari vendor;
func UrlEncodeSMS(text string) string {
	return strings.ReplaceAll(url.QueryEscape(text), "+", "%20")
}

// format agama jdi flagging
func FormatAgama(agama string) string {
	switch agama {
	case "ISLAM":
		return "I"
	case "KRISTEN":
		return "P"
	case "KATOLIK":
		return "K"
	case "HINDU":
		return "H"
	case "BUDDHA":
		return "B"
	case "KONGHUCU":
		return "U"
	default:
		return ""
	}
}

// format negara ke dalam ISO3166-1 Alpha-2
func FormatNegaraAlpha2(negara string) (string, error) {
	cariNegara := gountries.New()

	negaraObj, err := cariNegara.FindCountryByName(negara)
	if err != nil {
		return "", err
	}

	return negaraObj.Alpha2, nil
}

// hitung usia (tahun) antara dua object time.Time
func HitUmur(from, to time.Time) int {
	years := to.Year() - from.Year()

	// kalau ulang tahun tahun ini belum lewat
	if to.Month() < from.Month() ||
		(to.Month() == from.Month() && to.Day() < from.Day()) {
		years--
	}

	return years
}

// fungsi hitung tiering pendapatan
func HitTierPendapatan(gaji float64) string {
	switch {
	case gaji < 3_000_000.00:
		return "1"
	case gaji < 5_000_000.00:
		return "2"
	case gaji < 10_000_000.00:
		return "3"
	case gaji < 20_000_000.00:
		return "4"
	case gaji < 50_000_000.00:
		return "5"
	case gaji < 100_000_000.00:
		return "6"
	case gaji < 500_000_000.00:
		return "7"
	default:
		return "8"
	}
}

// generate password dengan algoritma argon2id
func GenerateHashedPass(password string) (string, error) {
	salt := make([]byte, argonSaltLen)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	encodedPassword := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		encodedSalt,
		encodedHash,
	)

	return encodedPassword, nil
}

// verifikasi password
func VerifyPassword(password, encodedPassword string) bool {
	parts := strings.Split(encodedPassword, "$")
	if len(parts) != 6 {
		return false
	}

	if parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}

	var memory uint32
	var timeCost uint32
	var threads uint8

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	newHash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(newHash, expectedHash) == 1
}

// bikin 10 digit angka random
func GenerateRandom10Digit() (string, error) {
	min := int64(1000000000) // 10 digit terkecil
	max := int64(9999999999) // 10 digit terbesar

	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", n.Int64()+min), nil
}

// filter slice of struct dengan gunakan kondisi like
func FilterGenOptionsLike(data obj.ResListGenOptions, keyword string) obj.ResListGenOptions {
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	switch keyword {
	case "":
		return data
	}

	result := make(obj.ResListGenOptions, 0)

	for _, item := range data {
		label := strings.ToLower(item.LABEL)
		value := strings.ToLower(item.VALUE)

		if strings.Contains(label, keyword) || strings.Contains(value, keyword) {
			result = append(result, item)
		}
	}

	return result
}

func IsValidEmail(email string) bool {
	// Hilangkan spasi di awal dan akhir
	email = strings.TrimSpace(email)

	// Email tidak boleh kosong
	if email == "" {
		return false
	}

	// Parsing alamat email
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	// Pastikan hanya berisi alamat email, tanpa display name
	if addr.Address != email {
		return false
	}

	return true
}

func Round2(val float64) float64 {
	return math.Round(val*100) / 100
}

func ParseDateFlexible(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now(), nil
	}
	formats := []string{
		DateFormat,
		DateTimeFormat,
		time.RFC3339,
		"2006-01-02 15:04:05.000",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("Format tanggal '%s' tidak valid (gunakan format YYYY-MM-DD)", s)
}

func CalculateInstallmentDate(startDate time.Time, monthIndex int, installmentDay int) time.Time {
	targetYear := startDate.Year()
	targetMonth := int(startDate.Month()) + monthIndex
	for targetMonth > 12 {
		targetYear++
		targetMonth -= 12
	}

	lastDayOfMonth := time.Date(targetYear, time.Month(targetMonth+1), 0, 0, 0, 0, 0, startDate.Location()).Day()

	day := installmentDay
	if day <= 0 || day > 31 {
		day = startDate.Day()
	}
	if day > lastDayOfMonth {
		day = lastDayOfMonth
	}
	return time.Date(targetYear, time.Month(targetMonth), day, 0, 0, 0, 0, startDate.Location())
}

func NormalizeRmsBunga(input string) string {
	s := strings.ToUpper(strings.TrimSpace(input))
	switch {
	case s == "2" || strings.Contains(s, "ANUITAS"):
		return "2"
	case s == "4" || strings.Contains(s, "FLAT"):
		return "4"
	case s == "5" || strings.Contains(s, "EFEKTIF"):
		return "5"
	default:
		if s == "" {
			return "4"
		}
		return s
	}
}

func RmsBungaLabel(rms string) string {
	switch NormalizeRmsBunga(rms) {
	case "2":
		return "ANUITAS"
	case "4":
		return "FLAT"
	case "5":
		return "EFEKTIF"
	default:
		return "FLAT"
	}
}

func NormalizeJnsAngsur(input string) string {
	s := strings.TrimSpace(input)
	if s == "1" || strings.EqualFold(s, "Bunga") {
		return "1"
	}
	return "2"
}

func JnsAngsurLabel(jns string) string {
	if NormalizeJnsAngsur(jns) == "1" {
		return "Bunga"
	}
	return "Pokok + Bunga"
}

type InstallmentScheduleRow struct {
	AngsurKe int
	TglVal   time.Time
	PokokJdw float64
	BungaJdw float64
	Total    float64
	OsdebJdw float64
	BcadHari float64
}

func GenerateLoanSchedule(plafon float64, tenor int, annualRate float64, rmsbunga string, jnsangsur string, tglAwal time.Time, tglAngsur int) ([]InstallmentScheduleRow, float64, float64, float64, float64) {
	monthlyRate := (annualRate / 100.0) / 12.0
	schedule := make([]InstallmentScheduleRow, tenor)

	var angsurDeb, angsPokok, angsBunga, totBng float64

	if jnsangsur == "1" {
		// Angsuran Bunga saja setiap bulan, Pokok dibayar di akhir tempo
		angsBunga = Round2(plafon * monthlyRate)
		angsPokok = 0
		angsurDeb = angsBunga
		totBng = Round2(angsBunga * float64(tenor))

		curOsdeb := plafon
		for i := 1; i <= tenor; i++ {
			tglVal := CalculateInstallmentDate(tglAwal, i, tglAngsur)
			var p float64
			b := angsBunga
			if i == tenor {
				p = curOsdeb
				curOsdeb = 0
			} else {
				p = 0
			}
			schedule[i-1] = InstallmentScheduleRow{
				AngsurKe: i,
				TglVal:   tglVal,
				PokokJdw: p,
				BungaJdw: b,
				Total:    Round2(p + b),
				OsdebJdw: Round2(curOsdeb),
				BcadHari: Round2(b / 30.0),
			}
		}
		return schedule, angsurDeb, angsPokok, angsBunga, totBng
	}

	// JnsAngsur == "2" (Pokok + Bunga)
	switch rmsbunga {
	case "4": // FLAT
		angsPokok = Round2(plafon / float64(tenor))
		angsBunga = Round2(plafon * monthlyRate)
		angsurDeb = Round2(angsPokok + angsBunga)
		totBng = Round2(angsBunga * float64(tenor))

		curOsdeb := plafon
		for i := 1; i <= tenor; i++ {
			tglVal := CalculateInstallmentDate(tglAwal, i, tglAngsur)
			p := angsPokok
			b := angsBunga
			if i == tenor {
				p = Round2(curOsdeb) // penyesuaian sisa saldo akhir agar tepat 0
			}
			curOsdeb = Round2(curOsdeb - p)
			if curOsdeb < 0 {
				curOsdeb = 0
			}
			schedule[i-1] = InstallmentScheduleRow{
				AngsurKe: i,
				TglVal:   tglVal,
				PokokJdw: p,
				BungaJdw: b,
				Total:    Round2(p + b),
				OsdebJdw: curOsdeb,
				BcadHari: Round2(b / 30.0),
			}
		}

	case "5": // EFEKTIF (Sliding Rate)
		basePokok := Round2(plafon / float64(tenor))
		curOsdeb := plafon
		totBng = 0

		for i := 1; i <= tenor; i++ {
			tglVal := CalculateInstallmentDate(tglAwal, i, tglAngsur)
			p := basePokok
			if i == tenor {
				p = Round2(curOsdeb)
			}
			b := Round2(curOsdeb * monthlyRate)
			totBng += b
			curOsdeb = Round2(curOsdeb - p)
			if curOsdeb < 0 {
				curOsdeb = 0
			}

			if i == 1 {
				angsPokok = p
				angsBunga = b
				angsurDeb = Round2(p + b)
			}

			schedule[i-1] = InstallmentScheduleRow{
				AngsurKe: i,
				TglVal:   tglVal,
				PokokJdw: p,
				BungaJdw: b,
				Total:    Round2(p + b),
				OsdebJdw: curOsdeb,
				BcadHari: Round2(b / 30.0),
			}
		}
		totBng = Round2(totBng)

	case "2": // ANUITAS
		if monthlyRate > 0 {
			factor := math.Pow(1.0+monthlyRate, float64(tenor))
			angsurDeb = Round2(plafon * (monthlyRate * factor) / (factor - 1.0))
		} else {
			angsurDeb = Round2(plafon / float64(tenor))
		}

		curOsdeb := plafon
		totBng = 0

		for i := 1; i <= tenor; i++ {
			tglVal := CalculateInstallmentDate(tglAwal, i, tglAngsur)
			b := Round2(curOsdeb * monthlyRate)
			var p float64
			if i == tenor {
				p = Round2(curOsdeb)
			} else {
				p = Round2(angsurDeb - b)
			}
			totBng += b
			curOsdeb = Round2(curOsdeb - p)
			if curOsdeb < 0 {
				curOsdeb = 0
			}

			if i == 1 {
				angsPokok = p
				angsBunga = b
			}

			schedule[i-1] = InstallmentScheduleRow{
				AngsurKe: i,
				TglVal:   tglVal,
				PokokJdw: p,
				BungaJdw: b,
				Total:    Round2(p + b),
				OsdebJdw: curOsdeb,
				BcadHari: Round2(b / 30.0),
			}
		}
		totBng = Round2(totBng)

	default: // Default FLAT
		angsPokok = Round2(plafon / float64(tenor))
		angsBunga = Round2(plafon * monthlyRate)
		angsurDeb = Round2(angsPokok + angsBunga)
		totBng = Round2(angsBunga * float64(tenor))

		curOsdeb := plafon
		for i := 1; i <= tenor; i++ {
			tglVal := CalculateInstallmentDate(tglAwal, i, tglAngsur)
			p := angsPokok
			b := angsBunga
			if i == tenor {
				p = Round2(curOsdeb)
			}
			curOsdeb = Round2(curOsdeb - p)
			if curOsdeb < 0 {
				curOsdeb = 0
			}
			schedule[i-1] = InstallmentScheduleRow{
				AngsurKe: i,
				TglVal:   tglVal,
				PokokJdw: p,
				BungaJdw: b,
				Total:    Round2(p + b),
				OsdebJdw: curOsdeb,
				BcadHari: Round2(b / 30.0),
			}
		}
	}

	return schedule, angsurDeb, angsPokok, angsBunga, totBng
}

func FormatCrBayar(cr string) string {
	switch strings.ToUpper(strings.TrimSpace(cr)) {
	case "R":
		return "Tabungan"
	case "T":
		return "Tunai"
	case "B":
		return "Bank"
	default:
		return cr
	}
}

func FormatStsTrn(sts string) string {
	switch strings.TrimSpace(sts) {
	case "01":
		return "Pembayaran Angsuran"
	case "02":
		return "Pelunasan Kredit"
	default:
		return sts
	}
}
