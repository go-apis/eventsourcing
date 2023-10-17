package apub

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func New(ctx context.Context, cfg *es.ProviderConfig) (es.Streamer, error) {
	if cfg.Stream.Type != "apub" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Stream.Type)
	}

	awscfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Stream.AWS.Region))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	snsClient := sns.NewFromConfig(awscfg)
	sqsClient := sqs.NewFromConfig(awscfg)

	out, err := sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(cfg.Stream.AWS.QueueName),
	})
	if err != nil {
		return nil, err
	}

	return NewStreamer(
		cfg.Service,
		snsClient,
		sqsClient,
		cfg.Stream.AWS.TopicARN,
		*out.QueueUrl,
		cfg.Stream.AWS.MaxNumberOfMessages,
		cfg.Stream.AWS.WaitTimeSeconds,
	)
}

func init() {
	es.RegisterStreamProviders("apub", New)
}
