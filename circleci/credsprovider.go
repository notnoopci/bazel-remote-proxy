package circleci

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/cenk/backoff"
	pb "github.com/notnoopci/bazel-remote-proxy/circleci/protocol"
	"github.com/sirupsen/logrus"
)

const providerName = "CirclecIProvider"

type eventProcessorProvider struct {
	credentials.Expiry

	eventProcessor pb.EventProcessorClient
	context        context.Context
}

func (p *eventProcessorProvider) Retrieve() (credentials.Value, error) {
	result := credentials.Value{ProviderName: providerName}
	expiration := time.Now()

	logrus.Debug("about to fetch storage credentials")

	op := func() error {
		creds, err := p.eventProcessor.StorageCredentials(p.context, &pb.CredentialsRequest{})
		if err != nil || creds == nil {
			logrus.WithError(err).Warn("error retrieving storage credentials")
			return err
		}

		if creds.S3Credentials == nil {
			logrus.WithError(err).Error("error retrieving storage credentials")
			return backoff.Permanent(errors.New("no storage credentials found"))
		}

		s3creds := creds.S3Credentials

		result = credentials.Value{
			AccessKeyID:     s3creds.AccessKeyId,
			SecretAccessKey: s3creds.SecretAccessKey,
			SessionToken:    s3creds.SessionToken,

			ProviderName: providerName,
		}

		expiration = time.Unix(s3creds.ExpirationTimestamp/1000, 0)
		logrus.WithField("expiration", s3creds.ExpirationTimestamp).Info("S3 credentials retrieved")

		return nil
	}
	notify := func(err error, d time.Duration) {
		logrus.WithError(err).Warn("error fetching storage credentials, retrying")
	}

	b := backoff.WithContext(
		backoff.WithMaxTries(backoff.NewExponentialBackOff(), 10),
		p.context,
	)
	err := backoff.RetryNotify(op, b, notify)
	if err != nil {
		logrus.WithError(err).Error("error fetching storage credentials")
		return result, err
	}

	logrus.WithField("creds", result).Debug("retrieved credentials")
	p.SetExpiration(expiration, 10*time.Second)
	return result, err
}

func NewStorageProvider(ctx context.Context, eventProcessor pb.EventProcessorClient) credentials.Provider {
	return &eventProcessorProvider{
		eventProcessor: eventProcessor,
		context:        ctx,
	}
}
