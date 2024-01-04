package apub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/contextcloud/eventstore/es"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"go.opentelemetry.io/otel"
)

type Unsubscribe func(ctx context.Context) error

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
	unsubscribe  []Unsubscribe
}

func (s *streamer) createSubscription(ctx context.Context, suffix string) (*string, Unsubscribe, error) {
	queueName := s.service + ".fifo"
	if suffix != "" {
		queueName = fmt.Sprintf("%s-%s.fifo", s.service, suffix)
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

	unsubscribe := func(ctx context.Context) error {
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
	return createQueueRsp.QueueUrl, unsubscribe, nil
}

func (s *streamer) handle(queueUrl *string, handler es.EventHandler) func(context.Context, *types.Message) {
	return func(ctx context.Context, msg *types.Message) {
		var raw []byte
		if msg.Body != nil {
			raw = []byte(*msg.Body)
		}

		evt, err := es.GlobalRegistry.ParseEvent(ctx, raw)
		if err != nil && !errors.Is(err, es.ErrNotFound) {
			select {
			case s.errCh <- err:
			default:
				log.Printf("missed error in AWS event bus: %s", err)
			}
			return
		}

		if evt != nil {
			if err := handler.Handle(ctx, evt); err != nil {
				select {
				case s.errCh <- err:
				default:
					log.Printf("missed error in AWS event bus: %s", err)
				}
				return
			}
		}

		if _, err := s.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      queueUrl,
			ReceiptHandle: msg.ReceiptHandle,
		}); err != nil {
			err = fmt.Errorf("could not delete message: %w", err)
			select {
			case s.errCh <- err:
			default:
				log.Printf("missed error in AWS event bus: %s", err)
			}
		}
	}
}

func (s *streamer) loop(input *sqs.ReceiveMessageInput, handler es.EventHandler) {
	defer s.wg.Done()

	h := s.handle(input.QueueUrl, handler)

	for {
		select {
		case <-s.cctx.Done():
			return
		default:
			output, err := s.sqsClient.ReceiveMessage(s.cctx, input)
			if err != nil {
				err = fmt.Errorf("could not receive: %w", err)

				select {
				case s.errCh <- err:
				default:
					log.Printf("missed error in GCP event bus: %s", err)
				}

				// Retry the receive loop if there was an error.
				time.Sleep(time.Second)
				continue
			}

			for _, msg := range output.Messages {
				h(s.cctx, &msg)
			}

		}
		return
	}
}

func (s *streamer) AddHandler(ctx context.Context, name string, handler es.EventHandler) error {
	// Check handler existence.
	s.registeredMu.Lock()
	defer s.registeredMu.Unlock()

	if _, ok := s.registered[name]; ok {
		return fmt.Errorf("handler already registered: %s", name)
	}

	sub, unsubscribe, err := s.createSubscription(ctx, name)
	if err != nil {
		return err
	}

	if unsubscribe != nil {
		s.unsubscribe = append(s.unsubscribe, unsubscribe)
	}

	// Register handler.
	s.registered[name] = true
	s.wg.Add(1)

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            sub,
		MaxNumberOfMessages: int32(10),
		WaitTimeSeconds:     int32(20),
	}
	go s.loop(input, handler)
	return nil
}

func (s *streamer) Publish(ctx context.Context, evts ...*es.Event) error {
	if len(s.service) == 0 {
		return fmt.Errorf("streamer not started")
	}

	_, span := otel.Tracer("apub").Start(ctx, "Publish")
	defer span.End()

	for _, evt := range evts {
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
	for _, unsub := range s.unsubscribe {
		if err := unsub(ctx); err != nil {
			s.errCh <- err
		}
	}

	return nil
}

func NewStreamer(ctx context.Context, service string, cfg *es.AwsSnsConfig) (es.Streamer, error) {
	awscfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region), config.WithSharedConfigProfile(cfg.Profile))
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
	return &streamer{
		service:    service,
		snsClient:  snsClient,
		sqsClient:  sqsClient,
		config:     cfg,
		attributes: out.Attributes,
		cctx:       cctx,
		cancel:     cancel,
		registered: make(map[string]bool),
		errCh:      make(chan error, 100),
	}, nil
}
