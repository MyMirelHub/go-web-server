package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func body(w http.ResponseWriter, r *http.Request) {
	len := r.ContentLength

	if len <= 2 {
		http.Error(w, "String lenght minimum of 2 allowed, try again", http.StatusBadRequest)
		return
	}

	body := make([]byte, len)
	r.Body.Read(body)

	log.Printf("string submitted: %s\n", string(body))

	for index, element := range repetition(string(body)) {
		if element > 2 {
			fmt.Fprintln(w, index, "=", element)
		}
	}
}

// logger
func loggingmiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("Received request from %s\n", r.RemoteAddr)
		log.Printf("%s %s", r.Method, time.Since(start))
	})
}

// Create a function that takes a string and returns a map
func repetition(st string) map[string]int {

	input := strings.SplitAfter(st, "")
	wc := make(map[string]int)
	for _, s := range input {
		_, matched := wc[s]
		if matched {
			wc[s] += 1
		} else {
			wc[s] = 1
		}
	}
	return wc
}

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kata_http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

// prometheusMiddleware implements mux.MiddlewareFunc
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		next.ServeHTTP(w, r)
		totalRequests.WithLabelValues(path).Inc()
	})
}

func init() {
	prometheus.Register(totalRequests)
}

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*5, "the duration for which the server gracefully wait for existing connections to finish")
	flag.Parse()

	r := mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Use(loggingmiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	// r.HandleFunc("/", body).Methods("PUT")
	r.Path("/").HandlerFunc(body).Methods("PUT")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	fmt.Printf("Starting server at port %s\n", port)

	srv := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      r,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// Wait for termination
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	signal.Notify(c, os.Interrupt)
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)

}
