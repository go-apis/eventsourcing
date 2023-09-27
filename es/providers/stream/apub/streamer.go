package apub

import (
	"context"
	"fmt"

	"github.com/contextcloud/eventstore/es"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"go.opentelemetry.io/otel"
)

type streamer struct {
	snsClient           *sns.Client
	sqsClient           *sqs.Client
	topicArn            string
	queueUrl            string
	maxNumberOfMessages int
	waitTimeSeconds     int

	worker  Worker
	service string
}

func (s *streamer) Start(ctx context.Context, cfg es.Config, callback es.EventCallback) error {
	_, span := otel.Tracer("apub").Start(ctx, "Start")
	defer span.End()

	if cfg == nil {
		return fmt.Errorf("cfg is required")
	}
	if callback == nil {
		return fmt.Errorf("callback is required")
	}

	if s.worker != nil {
		return fmt.Errorf("streamer already started")
	}

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueUrl),
		MaxNumberOfMessages: int32(s.maxNumberOfMessages),
		WaitTimeSeconds:     int32(s.waitTimeSeconds),
	}

	worker, err := NewWorker(s.sqsClient, cfg, input, callback)
	if err != nil {
		return err
	}

	s.worker = worker
	s.service = cfg.GetProviderConfig().Service
	return s.worker.Start(ctx)
}

func (s *streamer) Publish(ctx context.Context, evts ...*es.Event) error {
	if len(s.service) == 0 {
		return fmt.Errorf("streamer not started")
	}

	_, span := otel.Tracer("apub").Start(ctx, "Publish")
	defer span.End()

	messages := make([]*sns.PublishInput, len(evts))
	for i, evt := range evts {
		messageDeduplicationId := fmt.Sprintf("%s:%s:%s:%d", evt.Namespace, evt.AggregateType, evt.AggregateId.String(), evt.Version)
		messageGroupId := fmt.Sprintf("%s:%s:%s", evt.Namespace, evt.AggregateType, evt.AggregateId.String())
		data, err := es.MarshalEvent(ctx, s.service, evt)
		if err != nil {
			return err
		}
		messages[i] = &sns.PublishInput{
			Message:                aws.String(string(data)),
			TopicArn:               aws.String(s.topicArn),
			MessageDeduplicationId: aws.String(messageDeduplicationId),
			MessageGroupId:         aws.String(messageGroupId),
		}
	}

	for _, msg := range messages {
		_, err := s.snsClient.Publish(ctx, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("apub").Start(ctx, "Close")
	defer span.End()

	if s.worker == nil {
		return nil
	}

	if err := s.worker.Close(); err != nil {
		return err
	}
	s.worker = nil
	return nil
}

func NewStreamer(
	snsClient *sns.Client,
	sqsClient *sqs.Client,
	topicArn string,
	queueUrl string,
	maxNumberOfMessages int,
	waitTimeSeconds int,
) (es.Streamer, error) {
	return &streamer{
		snsClient:           snsClient,
		sqsClient:           sqsClient,
		topicArn:            topicArn,
		queueUrl:            queueUrl,
		maxNumberOfMessages: maxNumberOfMessages,
		waitTimeSeconds:     waitTimeSeconds,
	}, nil
}
