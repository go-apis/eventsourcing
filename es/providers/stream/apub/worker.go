package apub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/contextcloud/eventstore/es"
)

type Worker interface {
	Start(ctx context.Context) error
	Close() error
}

type worker struct {
	sqsClient *sqs.Client
	cfg       es.Config
	input     *sqs.ReceiveMessageInput
	callback  es.EventCallback
}

func (w *worker) handle(ctx context.Context, msg types.Message) error {
	if msg.Body == nil {
		return nil
	}

	data := []byte(*msg.Body)
	with, err := es.UnmarshalEvent(ctx, w.cfg, data)
	if err != nil {
		return err
	}

	// nothing todo
	if with == nil {
		return nil
	}

	return w.callback(ctx, with.Event)
}
func (w *worker) poll(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			output, err := w.sqsClient.ReceiveMessage(ctx, w.input)
			if err != nil {
				fmt.Printf("failed to fetch sqs message %v", err)
			}
			for _, msg := range output.Messages {
				if err := w.handle(ctx, msg); err != nil {
					fmt.Printf("failed to handle sqs message %v", err)
					continue
				}
				w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      w.input.QueueUrl,
					ReceiptHandle: msg.ReceiptHandle,
				})
			}
		}
	}
}

func (w *worker) Start(ctx context.Context) error {
	go w.poll(ctx)
	return nil
}

func (w *worker) Close() error {
	return nil
}

func NewWorker(
	sqsClient *sqs.Client,
	cfg es.Config,
	input *sqs.ReceiveMessageInput,
	callback es.EventCallback,
) (Worker, error) {
	return &worker{
		sqsClient: sqsClient,
		cfg:       cfg,
		input:     input,
		callback:  callback,
	}, nil
}
