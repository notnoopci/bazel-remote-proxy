package main

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

func TestEffectiveKeyWorks(t *testing.T) {
	cases := []struct{ expected, prefix, userPath string }{
		{"path", "", "path/"},
		{"prefix/path", "prefix", "path"},
		{"prefix/path", "prefix", "path/"},
		{"prefix", "prefix", ""},
	}

	for _, c := range cases {
		e := effectiveKey(c.prefix, c.userPath)
		assert.Equal(t, c.expected, e)
	}
}

func TestS3Director(t *testing.T) {

	sess := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Credentials:      credentials.NewStaticCredentials("myid", "secret", ""),
		S3ForcePathStyle: aws.Bool(true),
	}))

	root, _ := url.Parse("s3://mybucket/path/to/prefix")

	director, _ := newS3Director(sess, root)

	methods := []string{
		"GET",
		"PUT",
		"DELETE",
	}

	for _, method := range methods {
		t.Run(method+" case", func(t *testing.T) {
			r := httptest.NewRequest(method, "https://localhost:8080/cas/test", nil)
			director.Direct(r)

			assert.Equal(t, method, r.Method)
			assert.Equal(t, "s3.amazonaws.com", r.URL.Host)
			assert.Equal(t, "/mybucket/path/to/prefix/cas/test", r.URL.Path)
			assert.Contains(t, r.URL.Query().Get("X-Amz-Credential"), "myid")
			assert.Equal(t, "", r.Host)
		})
	}
}
