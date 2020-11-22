package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rsc.io/quote"
)

func main() {
	{
		// log start and end
		const srvName = "k8s-demo-web"
		log.Printf("%s started", srvName)
		defer log.Printf("%s stopped", srvName)
	}

	ctx, cancel := signalContext()
	defer cancel()

	var server *http.Server
	{
		mux := http.NewServeMux()
		mux.Handle("/hi", hiHandler())
		server = &http.Server{
			Addr:           ":8080",
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
	}

	// shutdown server when ctx is cancelled
	go func() {
		<-ctx.Done()
		timeout := 5 * time.Second
		shutdownCTX, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		server.Shutdown(shutdownCTX)
	}()

	log.Println("http: Server started")
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func signalContext() (ctx context.Context, cancel func()) {
	ctx, localCancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	cancel = func() {
		signal.Stop(c)
		localCancel()
	}

	go func() {
		select {
		case s, ok := <-c:
			if ok {
				log.Printf("signal caught: %s\n", s)
			}
			localCancel()
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}

func hiHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintf(w, quote.Go())

		status := http.StatusOK
		log.Printf("%d %s elapsed=%s\n", status, http.StatusText(status), time.Since(start))
	})
}
