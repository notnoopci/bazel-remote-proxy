package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

var (
	port    = flag.Int("port", 8080, "Port the HTTP server listens to")
	backend = flag.String("backend", "", "uri of backend storage service, e.g. s3://my-bazel-cache/prefix")
)

type director interface {
	Direct(*http.Request)
}

func main() {
	flag.Parse()

	if *backend == "" {
		flag.Usage()
		os.Exit(1)
		return
	}

	session := session.New()

	backendURL, err := url.Parse(*backend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "passed backend isn't a uri: %v\n  %v\n", *backend, err)
		flag.Usage()
		os.Exit(1)
	}

	var d director
	switch backendURL.Scheme {
	case "s3":
		d, _ = newS3Director(session, backendURL)
	default:
		fmt.Fprintf(os.Stderr, "only S3 is supported currently\n")
		flag.Usage()
		os.Exit(1)
	}

	handler := &httputil.ReverseProxy{
		Director: d.Direct,
	}

	addr := "127.0.0.1:" + strconv.Itoa(*port)
	s := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	logrus.WithField("address", addr).Info("starting proxy and listening")
	logrus.Fatal(s.ListenAndServe())
}
