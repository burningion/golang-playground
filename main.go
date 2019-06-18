package main

import (
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func sayHello(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	traceID := span.Context().TraceID()
	spanID := span.Context().SpanID()

	message := r.URL.Path

	// set a tag for the current path
	span.SetTag("url.path", message)

	message = strings.TrimPrefix(message, "/")

	// log with matching trace ID
	log.WithFields(log.Fields{
		"dd.trace_id": traceID,
		"dd.span_id":  spanID,
		"message":     message,
	}).Info("root url called with " + message)

	message = "Hello " + message

	w.Write([]byte(message))
}

func sayPong(w http.ResponseWriter, r *http.Request) {
	span, _ := tracer.SpanFromContext(r.Context())
	traceID := span.Context().TraceID()
	spanID := span.Context().SpanID()

	log.WithFields(log.Fields{
		"dd.trace_id": traceID,
		"dd.span_id":  spanID,
	}).Info("Ping / Pong request called")
	w.Write([]byte("pong"))
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.Info("starting up")

	// start the tracer with zero or more options
	tracer.Start()
	defer tracer.Stop()

	mux := httptrace.NewServeMux(httptrace.WithServiceName("test-go"), httptrace.WithAnalytics(true)) // init the http tracer
	mux.HandleFunc("/ping", sayPong)
	mux.HandleFunc("/", sayHello) // use the tracer to handle the urls

	err := http.ListenAndServe(":8080", mux) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
