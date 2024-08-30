package inmem

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func startHTTPServer(t *testing.T, register func(engine *gin.Engine) error) int {
	addr := "localhost:0"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	if err := register(router); err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           router.Handler(),
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
	}
	ctx := context.Background()
	t.Cleanup(func() {
		_ = server.Shutdown(ctx)
	})

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	listenerPort := listener.Addr().(*net.TCPAddr).Port

	c := http.Client{
		Timeout: 60 * time.Second,
	}
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "localhost:"+strconv.Itoa(listenerPort)+"/healthz", nil)
	healthCheck(c, req)

	return listenerPort
}

func healthCheck(c http.Client, req *http.Request) {
	for i := 0; i < 20; i++ {
		resp, err := c.Do(req)
		if err != nil {
			continue
		}

		_, err = io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		err = resp.Body.Close()
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
}
