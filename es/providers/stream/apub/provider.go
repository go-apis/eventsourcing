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

func New(ctx context.Context, cfg es.StreamConfig) (es.Streamer, error) {
	if cfg.Type != "apub" {
		return nil, fmt.Errorf("invalid data provider type: %s", cfg.Type)
	}

	awscfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.AWS.Region))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	snsClient := sns.NewFromConfig(awscfg)
	sqsClient := sqs.NewFromConfig(awscfg)

	out, err := sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(cfg.AWS.QueueName),
	})
	if err != nil {
		return nil, err
	}

	return NewStreamer(
		snsClient,
		sqsClient,
		cfg.AWS.TopicARN,
		*out.QueueUrl,
		cfg.AWS.MaxNumberOfMessages,
		cfg.AWS.WaitTimeSeconds,
	)
}

func init() {
	es.RegisterStreamProviders("apub", New)
}
