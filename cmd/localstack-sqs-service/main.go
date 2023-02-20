package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// SQSClient represents the SQS client interface
type SQSClient interface {
	sqsiface.SQSAPI
}

func main() {
	// Create an AWS session on Localstack
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Endpoint:         aws.String("http://localhost:4566"),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		},
	})
	if err != nil {
		log.Fatalf("failed to create session: %v", err)
	}

	// Create an SNS client on Localstack
	snsClient := sns.New(sess)

	// Create an SQS client on Localstack
	sqsClient := sqs.New(sess)

	// Create a new SNS topic
	topic, err := snsClient.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String("test-topic"),
	})
	if err != nil {
		log.Fatalf("failed to create SNS topic: %v", err)
	}

	// Create a new SQS queue
	queue, err := sqsClient.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String("test-queue"),
	})
	if err != nil {
		log.Fatalf("failed to create SQS queue: %v", err)
	}

	// Subscribe the SQS queue to the SNS topic
	_, err = snsClient.Subscribe(&sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		TopicArn: topic.TopicArn,
		Endpoint: queue.QueueUrl,
	})
	if err != nil {
		log.Fatalf("failed to subscribe to SNS topic: %v", err)
	}

	// Start a goroutine to listen for messages on the SQS queue
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := sqsClient.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
					QueueUrl:            queue.QueueUrl,
					MaxNumberOfMessages: aws.Int64(10),
					VisibilityTimeout:   aws.Int64(60), // 1 minute
					WaitTimeSeconds:     aws.Int64(20),
				})
				if err != nil {
					log.Printf("failed to receive message: %v", err)
				}

				for _, message := range result.Messages {
					fmt.Println(*message.Body)
					_, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      queue.QueueUrl,
						ReceiptHandle: message.ReceiptHandle,
					})
					if err != nil {
						log.Printf("failed to delete message: %v", err)
					}
				}
			}
		}
	}()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a signal
	<-sigCh
	// Wait for a signal
	log.Println("Graceful shutdown signal received, waiting for all messages to be processed...")

	// Cancel the context to stop receiving new messages
	cancel()

	// Wait for the goroutine to exit
	wg.Wait()

	log.Println("All messages processed. Exiting gracefully.")
}
