package main

import (
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Define a type for a handler that provides tracing parameters as well
type tracedHandler func(tracer.Span, *log.Entry, http.ResponseWriter, *http.Request)

// Write a wrapper function that does the magic preparation before calling the traced handler.
// This returns a function that is suitable for passing to mux.HandleFunc
func withSpanAndLogger(t tracedHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		span, _ := tracer.SpanFromContext(r.Context())
		traceID := span.Context().TraceID()
		spanID := span.Context().SpanID()
		entry := log.WithFields(log.Fields{
			"dd.trace_id": traceID,
			"dd.span_id":  spanID,
		})
		t(span, entry, w, r)
	}
}

func sayHello(span tracer.Span, l *log.Entry, w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path

	// set a tag for the current path
	span.SetTag("url.path", message)

	message = strings.TrimPrefix(message, "/")

	// log with matching trace ID
	l.WithFields(log.Fields{
		"message": message,
	}).Info("root url called with " + message)

	message = "Hello " + message

	w.Write([]byte(message))
}

func sayPong(span tracer.Span, l *log.Entry, w http.ResponseWriter, r *http.Request) {
	l.Info("Ping / Pong request called")

	w.Write([]byte("pong"))
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("starting up")

	// start the tracer with zero or more options
	tracer.Start()
	defer tracer.Stop()

	mux := httptrace.NewServeMux(httptrace.WithServiceName("test-go"), httptrace.WithAnalytics(true)) // init the http tracer
	mux.HandleFunc("/ping", withSpanAndLogger(sayPong))
	mux.HandleFunc("/", withSpanAndLogger(sayHello)) // use the tracer to handle the urls

	err := http.ListenAndServe(":8080", mux) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
