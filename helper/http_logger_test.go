package helper

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func testHTTPLogs(t *testing.T, process func(httpLogTask)) {
	t.Helper()
	previous := httpLogs
	httpLogs = newHTTPLogQueue(process)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := FlushHTTPLogs(ctx); err != nil {
			t.Errorf("flush logs: %v", err)
		}
		close(httpLogs.tasks)
		httpLogs = previous
	})
}

func flushTestHTTPLogs(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := FlushHTTPLogs(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPLoggerResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name     string
		path     string
		body     string
		status   int
		endpoint string
		panic    bool
	}{
		{name: "normal", path: "/items/123", body: "payload", status: 200, endpoint: "/items/:id"},
		{name: "missing route", path: "/missing", status: 404, endpoint: "/missing"},
		{name: "large body stays complete", path: "/items/123", body: strings.Repeat("x", (3<<20)+1), status: 200, endpoint: "/items/:id"},
		{name: "panic", path: "/items/123", status: 500, endpoint: "/items/:id", panic: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var tasks []httpLogTask
			testHTTPLogs(t, func(task httpLogTask) {
				if task.request != nil {
					*task.logid = 42
				}
				tasks = append(tasks, task)
			})
			router := gin.New()
			router.Use(ResponseLogger(), gin.RecoveryWithWriter(io.Discard), RequestLogger(), ContextCleaner())
			router.NoRoute(func(c *gin.Context) { c.String(http.StatusNotFound, "404 page not found") })
			var businessCtx context.Context
			router.POST("/items/:id", func(c *gin.Context) {
				var cancel context.CancelFunc
				businessCtx, cancel = context.WithCancel(c.Request.Context())
				c.Set("cancelCtx", cancel)
				if test.panic {
					panic("test panic")
				}
				body, err := io.ReadAll(c.Request.Body)
				if err != nil || string(body) != test.body {
					t.Errorf("handler body = %q, err = %v", body, err)
				}
				c.String(http.StatusOK, "ok")
			})
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body)))
			flushTestHTTPLogs(t)
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
			if len(tasks) != 2 || tasks[0].request == nil || tasks[1].updates == nil {
				t.Fatalf("expected ordered insert and update, got %d tasks", len(tasks))
			}
			if *tasks[0].request.REQUEST_ENDPOINT != test.endpoint {
				t.Errorf("endpoint = %q", *tasks[0].request.REQUEST_ENDPOINT)
			}
			if tasks[1].updates["response_code"] != httpStatus(test.status) {
				t.Errorf("logged status = %v", tasks[1].updates["response_code"])
			}
			if tasks[1].updates["response_text"] != recorder.Body.String() {
				t.Errorf("logged response differs from client response")
			}
			if len(tasks[1].updates) != 3 {
				t.Errorf("update contains fields beyond response")
			}
			if businessCtx != nil && !errors.Is(businessCtx.Err(), context.Canceled) {
				t.Error("business context was not canceled")
			}
		})
	}
}

func httpStatus(code int) string {
	return map[int]string{200: "200 - OK", 404: "404 - Not Found", 413: "413 - Request Entity Too Large", 500: "500 - Internal Server Error"}[code]
}

func TestHTTPLoggerDoesNotWaitForDatabase(t *testing.T) {
	blocked := make(chan struct{})
	testHTTPLogs(t, func(task httpLogTask) { <-blocked })
	defer close(blocked)
	router := gin.New()
	router.Use(ResponseLogger(), gin.RecoveryWithWriter(io.Discard), RequestLogger())
	router.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	done := make(chan struct{})
	go func() {
		defer close(done)
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("request blocked by log database")
	}
}

func TestHTTPLoggerQueueFull(t *testing.T) {
	queue := &httpLogQueue{tasks: make(chan httpLogTask, 1)}
	if !queue.enqueue(httpLogTask{}) || queue.enqueue(httpLogTask{}) {
		t.Fatal("full queue should reject new task without blocking")
	}
}

func TestHTTPLoggerCaptureLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var requestText, responseText string
	testHTTPLogs(t, func(task httpLogTask) {
		if task.request != nil {
			requestText = *task.request.REQUEST_TEXT
			*task.logid = 42
		} else {
			responseText = task.updates["response_text"].(string)
		}
	})
	payload := strings.Repeat("x", maxHTTPLogBodyBytes+1)
	router := gin.New()
	router.Use(ResponseLogger(), gin.RecoveryWithWriter(io.Discard), RequestLogger())
	router.POST("/", func(c *gin.Context) { c.String(http.StatusOK, payload) })
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload)))
	flushTestHTTPLogs(t)
	if recorder.Body.String() != payload {
		t.Fatal("client response was truncated")
	}
	if !strings.HasSuffix(requestText, logTruncatedText) || !strings.HasSuffix(responseText, logTruncatedText) {
		t.Fatal("truncated logs must carry a marker")
	}
	if len(responseText) != maxHTTPLogBodyBytes+len(logTruncatedText) {
		t.Fatal("response log exceeded capture limit")
	}
}

func TestHTTPLoggerSkipsUpdateWithoutInsertedID(t *testing.T) {
	// Tidak boleh menyentuh DB ketika insert gagal/tidak menghasilkan identitas.
	logid := int64(0)
	saveHTTPLog(httpLogTask{logid: &logid, updates: map[string]any{"response_text": "ok"}})
}

func TestHTTPLoggerSlowRequestBody(t *testing.T) {
	var status string
	testHTTPLogs(t, func(task httpLogTask) {
		if task.request != nil {
			*task.logid = 42
		} else {
			status = task.updates["response_code"].(string)
		}
	})
	router := gin.New()
	router.Use(ResponseLogger(), gin.RecoveryWithWriter(io.Discard), RequestLogger())
	router.POST("/", func(c *gin.Context) { c.String(http.StatusOK, "unexpected") })
	server := httptest.NewUnstartedServer(router)
	server.Config.ReadTimeout = 50 * time.Millisecond
	server.Start()
	defer server.Close()
	conn, err := net.DialTimeout("tcp", server.Listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	// Kirim body sebagian saja agar ReadTimeout server terpicu.
	if _, err := io.WriteString(conn, "POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 10\r\n\r\nx"); err != nil {
		t.Fatal(err)
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	flushTestHTTPLogs(t)
	if response.StatusCode != http.StatusRequestTimeout || status != "408 - Request Timeout" {
		t.Fatalf("client status = %d, logged status = %q", response.StatusCode, status)
	}
}
