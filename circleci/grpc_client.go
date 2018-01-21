package circleci

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/url"

	"github.com/sirupsen/logrus"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func grpcTLSOption(caCert, cert, key []byte) (grpc.DialOption, error) {
	var err error

	// TLS config
	var tlsConfig tls.Config

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, errors.New("There was an error reading certificate")
	}

	tlsConfig.RootCAs = certPool
	if key != nil || cert != nil {
		keypair, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{keypair}
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Fatal("Failed to parse credentials")
	}

	creds := credentials.NewTLS(&tlsConfig)
	return grpc.WithTransportCredentials(creds), nil
}

func NewGrpcConnection(ctx context.Context, destUrl string, caCert, clientCert, clientKey []byte) (*grpc.ClientConn, error) {
	if destUrl == "" {
		return nil, errors.New("address is required")
	}

	url, err := url.Parse(destUrl)
	if err != nil {
		logrus.WithFields(logrus.Fields{"address": destUrl, "error": err}).Error("error connecting to address")
		return nil, err
	}

	var opts []grpc.DialOption

	if url.Scheme == "https" {
		tlsOpt, err := grpcTLSOption(caCert, clientCert, clientKey)
		if err != nil {
			logrus.WithError(err).Error("error configuring TLS")
			return nil, err
		}

		opts = append(opts, tlsOpt)
	}

	conn, err := grpc.DialContext(ctx, url.Host, opts...)
	if err != nil {
		logrus.WithFields(logrus.Fields{"address": destUrl, "error": err}).Error("error connecting to address")
		return nil, err
	}

	return conn, nil
}
