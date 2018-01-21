package main

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"

	"github.com/notnoopci/bazel-remote-proxy/circleci"
	pb "github.com/notnoopci/bazel-remote-proxy/circleci/protocol"
)

func newCircleCIDirector(url *url.URL) (director, error) {

	config, err := circleci.ReadCircleCIConfig()
	if err != nil {
		logrus.WithError(err).Error("error reading circleci config")
		return nil, err
	}

	ctx := context.Background()

	outerConn, err := circleci.NewGrpcConnection(
		ctx,
		config.ServiceURL,
		config.AuthCA,
		config.AuthClientCert,
		config.AuthClientKey,
	)

	if err != nil {
		logrus.WithFields(logrus.Fields{"address": config.ServiceURL, "error": err}).Error("error connecting to Outer")
		return nil, err
	}

	processor := pb.NewEventProcessorClient(outerConn)
	provider := circleci.NewStorageProvider(ctx, processor)
	awsCredentials := credentials.NewCredentials(provider)

	session := session.New(&aws.Config{
		Region:      aws.String(config.AWSRegion),
		Credentials: awsCredentials,
	})

	path := "global/caches/bazel"
	if url.Path != "" && url.Path != "/" {
		path = url.Path
	}

	delegated, err := url.Parse(config.StorageRoot + "/" + path)
	if err != nil {
		return nil, err
	}

	return newS3Director(session, delegated)
}
