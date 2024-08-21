package apub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-apis/eventsourcing/es"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"go.opentelemetry.io/otel"
)

type CleanUp func(ctx context.Context) error

func IsQueueDoesNotExist(err error) bool {
	for {
		if err == nil {
			return false
		}
		if _, ok := err.(*types.QueueDoesNotExist); ok {
			return true
		}
		err = errors.Unwrap(err)
	}
}

func queuePolicySNSToSQS(topicARN string) string {
	var buf strings.Builder
	err := json.NewEncoder(&buf).Encode(
		map[string]any{
			"Version": "2012-10-17",
			"Statement": map[string]any{
				"Sid":       "SNSTopicSendMessage",
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    "sqs:SendMessage",
				"Resource":  "*",
				"Condition": map[string]any{
					"ArnEquals": map[string]any{
						"aws:SourceArn": topicARN,
					},
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type streamer struct {
	service    string
	snsClient  *sns.Client
	sqsClient  *sqs.Client
	config     *es.AwsSnsConfig
	attributes map[string]string

	cctx   context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	errCh chan error

	registered   map[string]bool
	registeredMu sync.RWMutex
	cleanups     []CleanUp
}

func (s *streamer) createSubscription(ctx context.Context, suffix string) (*string, CleanUp, error) {
	queueName := s.service
	if s.config.QueueName != "" {
		queueName = s.config.QueueName
	}
	if strings.HasSuffix(queueName, ".fifo") {
		queueName = strings.TrimSuffix(queueName, ".fifo")
	}
	if suffix != "" {
		queueName = fmt.Sprintf("%s-%s", queueName, suffix)
	}
	if !strings.HasSuffix(queueName, ".fifo") {
		queueName = queueName + ".fifo"
	}

	queueUrlRsp, err := s.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err == nil {
		return queueUrlRsp.QueueUrl, nil, nil
	}
	if !IsQueueDoesNotExist(err) {
		return nil, nil, err
	}

	// create the queue.
	createQueueRsp, err := s.sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"FifoQueue": "true",
			"Policy":    queuePolicySNSToSQS(s.config.TopicArn),
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("creating SQS queue %q: %w", queueName, err)
	}

	// AWS docs say to wait at least 1 second after creating a queue
	timer := time.NewTimer(1 * time.Second)
	select {
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("waiting after creating SQS queue %q: %w", queueName, ctx.Err())
	case <-timer.C:
	}

	queueAttributes, err := s.sqsClient.GetQueueAttributes(ctx,
		&sqs.GetQueueAttributesInput{
			QueueUrl: createQueueRsp.QueueUrl,
			AttributeNames: []types.QueueAttributeName{
				types.QueueAttributeNameQueueArn,
			},
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("getting attributes for SQS queue %q: %w", queueName, err)
	}

	queueARNKey := string(types.QueueAttributeNameQueueArn)
	queueARN := queueAttributes.Attributes[queueARNKey]
	if queueARN == "" {
		return nil, nil, fmt.Errorf("SQS queue %q has empty ARN", queueName)
	}

	subscribeOutput, err := s.snsClient.Subscribe(ctx, &sns.SubscribeInput{
		Attributes: map[string]string{
			"RawMessageDelivery": "true",
		},
		Endpoint:              aws.String(queueARN),
		TopicArn:              aws.String(s.config.TopicArn),
		Protocol:              aws.String("sqs"),
		ReturnSubscriptionArn: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("subscribing SQS queue %q to SNS topic %q: %w", queueName, s.config.TopicArn, err)
	}

	cleanup := func(ctx context.Context) error {
		if _, err := s.snsClient.Unsubscribe(ctx, &sns.UnsubscribeInput{
			SubscriptionArn: subscribeOutput.SubscriptionArn,
		}); err != nil {
			return fmt.Errorf("unsubscribing SQS queue %q from SNS topic %q: %w", queueName, s.config.TopicArn, err)
		}

		if _, err := s.sqsClient.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: createQueueRsp.QueueUrl,
		}); err != nil {
			return fmt.Errorf("deleting SQS queue %q: %w", queueName, err)
		}
		return nil
	}
	return createQueueRsp.QueueUrl, cleanup, nil
}

func (s *streamer) loop(input *sqs.ReceiveMessageInput, handler es.MessageHandler) {
	defer s.wg.Done()

	h := func(msg *types.Message) error {
		var raw []byte
		if msg.Body != nil {
			raw = []byte(*msg.Body)
		}

		if err := handler(s.cctx, raw); err != nil {
			return fmt.Errorf("could not handle message: %w", err)
		}

		if _, err := s.sqsClient.DeleteMessage(s.cctx, &sqs.DeleteMessageInput{
			QueueUrl:      input.QueueUrl,
			ReceiptHandle: msg.ReceiptHandle,
		}); err != nil {
			return fmt.Errorf("could not delete message: %w", err)
		}
		return nil
	}

	for {
		select {
		case <-s.cctx.Done():
			return
		default:
			output, err := s.sqsClient.ReceiveMessage(s.cctx, input)
			if err != nil {
				s.errCh <- fmt.Errorf("could not receive: %w", err)
				time.Sleep(time.Second)
				continue
			}

			for _, msg := range output.Messages {
				if err := h(&msg); err != nil {
					s.errCh <- err
				}
			}
		}
	}
}

func (s *streamer) AddHandler(ctx context.Context, name string, handler es.MessageHandler) error {
	_, span := otel.Tracer("apub").Start(ctx, "AddHandler")
	defer span.End()

	// Check handler existence.
	s.registeredMu.Lock()
	defer s.registeredMu.Unlock()

	sub, cleanup, err := s.createSubscription(ctx, name)
	if err != nil {
		return err
	}

	if cleanup != nil {
		s.cleanups = append(s.cleanups, cleanup)
	}

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            sub,
		MaxNumberOfMessages: int32(10),
		WaitTimeSeconds:     int32(10),
	}

	go s.loop(input, handler)

	s.wg.Add(1)
	return nil
}

func (s *streamer) Publish(ctx context.Context, evt *es.Event) error {
	_, span := otel.Tracer("apub").Start(ctx, "Publish")
	defer span.End()

	messageDeduplicationId := fmt.Sprintf("%s:%s:%s:%s:%d", s.service, evt.Namespace, evt.AggregateType, evt.AggregateId.String(), evt.Version)
	messageGroupId := fmt.Sprintf("%s:%s:%s:%s", s.service, evt.Namespace, evt.AggregateType, evt.AggregateId.String())
	data, err := es.MarshalEvent(ctx, evt)
	if err != nil {
		return err
	}
	msg := &sns.PublishInput{
		Message:                aws.String(string(data)),
		TopicArn:               aws.String(s.config.TopicArn),
		MessageDeduplicationId: aws.String(messageDeduplicationId),
		MessageGroupId:         aws.String(messageGroupId),
	}
	if _, err := s.snsClient.Publish(ctx, msg); err != nil {
		return err
	}

	return nil
}

func (s *streamer) Errors() <-chan error {
	return s.errCh
}

func (s *streamer) Close(ctx context.Context) error {
	_, span := otel.Tracer("apub").Start(ctx, "Close")
	defer span.End()

	s.cancel()
	s.wg.Wait()

	// unsubscribe any ephemeral subscribers we created.
	for _, cleanup := range s.cleanups {
		if err := cleanup(ctx); err != nil {
			s.errCh <- err
		}
	}

	return nil
}

func NewStreamer(ctx context.Context, service string, cfg *es.AwsSnsConfig) (es.Streamer, error) {
	awsOpts := []func(*config.LoadOptions) error{}
	if cfg.Region != "" {
		awsOpts = append(awsOpts, config.WithRegion(cfg.Region))
	}
	if cfg.Profile != "" {
		awsOpts = append(awsOpts, config.WithSharedConfigProfile(cfg.Profile))
	}
	awscfg, err := config.LoadDefaultConfig(ctx, awsOpts...)
	if err != nil {
		return nil, err
	}

	snsClient := sns.NewFromConfig(awscfg)
	sqsClient := sqs.NewFromConfig(awscfg)

	out, err := snsClient.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: aws.String(cfg.TopicArn),
	})
	if err != nil {
		return nil, err
	}

	cctx, cancel := context.WithCancel(ctx)
	s := &streamer{
		service:    service,
		snsClient:  snsClient,
		sqsClient:  sqsClient,
		config:     cfg,
		attributes: out.Attributes,
		cctx:       cctx,
		cancel:     cancel,
		registered: make(map[string]bool),
		errCh:      make(chan error, 100),
	}
	return s, nil
}
