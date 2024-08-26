package restapi

import (
	"context"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

// create a channel to receive OS signals
var sigChan chan os.Signal = make(chan os.Signal, 1)
var router *Router

func TestCreateRoutes(t *testing.T) {
	router = &Router{}

	router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, routeContext *RouteContext) {
		w.WriteHeader(http.StatusOK)
	})

}
func TestCreateAndStartServer(t *testing.T) {
	// create a server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	// start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Error(err)
		}
	}()
	client := &http.Client{}
	client.Timeout = 10 * time.Second

	wg := &sync.WaitGroup{}

	// simulate a simple request
	wg.Add(1)
	go func() {
		resp, err := client.Get("http://localhost:8080/test")
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		} else {
			t.Log("Client request successful")
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		resp, err := client.Get("http://localhost:8080/test2")
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		} else {
			t.Log("Client request successful")
		}
		wg.Done()
	}()

	wg.Wait()
	sigChan <- syscall.SIGINT
	// wait for a signal
	<-sigChan

	// create a context with timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Server Shutdown Failed:%+v", err)
	}
}
