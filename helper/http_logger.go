package helper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	db "github.com/Tobias-Kusuma/core-ksp-api/database"
	"github.com/Tobias-Kusuma/core-ksp-api/obj"
	"github.com/gin-gonic/gin"
)

const (
	maxHTTPLogBodyBytes = 64 << 10
	httpLogQueueSize    = 128
	httpLogTimeout      = 5 * time.Second
	logTruncatedText    = "\n[body log dipotong]"
	LogApiCtxKey        = "log-api"
)

type httpLogTask struct {
	request *obj.LOG_HTTP
	logid   *int64
	updates map[string]any
	done    chan struct{}
}

type httpLogQueue struct {
	tasks chan httpLogTask
}

// Satu worker menjaga urutan insert request sebelum update response.
func newHTTPLogQueue(process func(httpLogTask)) *httpLogQueue {
	queue := &httpLogQueue{tasks: make(chan httpLogTask, httpLogQueueSize)}
	go func() {
		for task := range queue.tasks {
			if task.done != nil {
				close(task.done)
				continue
			}
			process(task)
		}
	}()
	return queue
}

var httpLogs = newHTTPLogQueue(saveHTTPLog)

func (queue *httpLogQueue) enqueue(task httpLogTask) bool {
	select {
	case queue.tasks <- task:
		return true
	default:
		log.Printf("ERROR: Antrean log HTTP penuh, log dilewati")
		return false
	}
}

// Dipanggil setelah server selesai menangani request, sebelum proses berhenti.
func FlushHTTPLogs(ctx context.Context) error {
	done := make(chan struct{})

	select {
	case httpLogs.tasks <- httpLogTask{done: done}:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func saveHTTPLog(task httpLogTask) {
	ctx, cancel := context.WithTimeout(context.Background(), httpLogTimeout)
	defer cancel()

	if task.request != nil {
		if err := db.Main().WithContext(ctx).Table("LOG_HTTP").Create(task.request).Error; err != nil {
			log.Printf("ERROR: Gagal menyimpan request API ke log dengan error: %s", err.Error())
			return
		}
		*task.logid = task.request.LOGID
		return
	}

	// logid hanya dibaca/ditulis worker, bukan goroutine request.
	if task.logid == nil || *task.logid <= 0 {
		return
	}

	result := db.Main().WithContext(ctx).Table("LOG_HTTP").Where("logid = ?", *task.logid).Updates(task.updates)

	if result.Error != nil {
		log.Printf("ERROR: Gagal menyimpan response API ke log dengan error: %s untuk logid: %d", result.Error.Error(), *task.logid)
	} else if result.RowsAffected == 0 {
		log.Printf("ERROR: Log HTTP tidak ditemukan untuk logid: %d", *task.logid)
	}
}

type responseLogWriter struct {
	gin.ResponseWriter
	body      bytes.Buffer
	truncated bool
}

func (w *responseLogWriter) capture(data []byte) {
	remaining := maxHTTPLogBodyBytes - w.body.Len()
	if len(data) > remaining {
		w.truncated = true
		data = data[:remaining]
	}

	w.body.Write(data)
}

func (w *responseLogWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.capture(data[:n])
	return n, err
}

func (w *responseLogWriter) WriteString(data string) (int, error) {
	n, err := w.ResponseWriter.WriteString(data)
	data = data[:n]
	remaining := maxHTTPLogBodyBytes - w.body.Len()

	if len(data) > remaining {
		w.truncated = true
		data = data[:remaining]
	}

	w.body.WriteString(data)
	return n, err
}

type requestLogBody struct {
	io.Reader
	io.Closer
}

func dumpIncomingRequest(c *gin.Context) (string, error) {
	var bodyBytes []byte
	var bodyErr error

	if c.Request.Body != nil {
		body := c.Request.Body
		// Baca cuplikan saja; satu byte tambahan untuk mendeteksi pemotongan log.
		bodyBytes, bodyErr = io.ReadAll(io.LimitReader(body, maxHTTPLogBodyBytes+1))
		// Kembalikan cuplikan di depan stream asli agar controller membaca body lengkap.
		c.Request.Body = &requestLogBody{
			Reader: io.MultiReader(bytes.NewReader(bodyBytes), body),
			Closer: body,
		}
	}

	dumpReq, err := httputil.DumpRequest(c.Request, false)
	if err != nil {
		return "", err
	}

	truncated := len(bodyBytes) > maxHTTPLogBodyBytes
	if truncated {
		bodyBytes = bodyBytes[:maxHTTPLogBodyBytes]
	}

	dump := string(dumpReq)
	if len(bodyBytes) > 0 {
		dump += "\r\n" + string(bodyBytes)
	}
	if truncated {
		dump += logTruncatedText
	}

	return dump, bodyErr
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		timeNow := time.Now()
		dumpReq, err := dumpIncomingRequest(c)
		clientIp := c.ClientIP()
		endpoint := c.FullPath()

		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		LogApi := obj.LOG_HTTP{
			REQUEST_ENDPOINT: &endpoint,
			REQUEST_TEXT:     &dumpReq,
			REQUEST_TIME:     &timeNow,
			CLIENT_IP:        &clientIp,
		}

		logid := new(int64)
		if httpLogs.enqueue(httpLogTask{request: &LogApi, logid: logid}) {
			c.Set(LogApiCtxKey, logid)
		}

		if err != nil {
			log.Printf("ERROR: Gagal membaca request API dengan error: %s", err.Error())

			statusCode := http.StatusBadRequest
			var netErr net.Error

			if errors.As(err, &netErr) && netErr.Timeout() {
				statusCode = http.StatusRequestTimeout
			}

			c.AbortWithStatusJSON(statusCode, gin.H{"error": http.StatusText(statusCode)})
			return
		}

		c.Next()
	}
}

// Daftarkan sebelum gin.Recovery() dan RequestLogger() agar response error ikut tertangkap.
func ResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := &responseLogWriter{ResponseWriter: c.Writer}
		c.Writer = writer

		defer func() {
			logApiAny, exists := c.Get(LogApiCtxKey)
			if !exists {
				return
			}

			logid, ok := logApiAny.(*int64)
			if !ok || logid == nil {
				log.Printf("ERROR: Logid tidak valid di context, tipe data: %T", logApiAny)
				return
			}

			timeNow := time.Now()
			statusCode := c.Writer.Status()
			response := writer.body.String()

			if writer.truncated {
				response += logTruncatedText
			}

			httpLogs.enqueue(httpLogTask{logid: logid, updates: map[string]any{
				"response_code": fmt.Sprintf("%d - %s", statusCode, http.StatusText(statusCode)),
				"response_text": response,
				"response_time": timeNow,
			}})
		}()

		c.Next()
	}
}
