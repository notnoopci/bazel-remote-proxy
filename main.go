package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

var (
	bind    = flag.String("bind", "127.0.0.1:7643", "address and port to bind to")
	backend = flag.String("backend", "", "uri of backend storage service, e.g. s3://my-bazel-cache/prefix")
)

type director interface {
	Direct(*http.Request)
}

func main() {
	var err error
	logrus.WithFields(logrus.Fields{
		"version":   Version,
		"GitCommit": GitCommit,
	}).Info("version info")

	flag.Parse()

	if *backend == "" {
		flag.Usage()
		os.Exit(1)
		return
	}

	backendURL, err := url.Parse(*backend)
	if err != nil {
		fmt.Fprintf(os.Stderr, "passed backend isn't a uri: %v\n  %v\n", *backend, err)
		flag.Usage()
		os.Exit(1)
	}

	var d director

	switch backendURL.Scheme {
	case "s3":
		d, err = newS3Director(session.New(), backendURL)
	case "circleci":
		d, err = newCircleCIDirector(backendURL)
	default:
		fmt.Fprintf(os.Stderr, "only S3 and circleci are supported currently\n")
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing backend: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	handler := &httputil.ReverseProxy{
		Director: d.Direct,
	}

	addr := *bind
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	s := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	logrus.WithField("address", addr).Info("starting proxy and listening")
	logrus.Fatal(s.ListenAndServe())
}
