# localstack-sqs-service
## Subscriber of Pub-sub model implementation in golang using AWS SQS service with localstack

#### In a pub-sub model using SNS and SQS, messages are published to an SNS topic, and then delivered to subscribed SQS queues. So, in order to read messages in a pub-sub model, you need to read messages from the SQS queue that is subscribed to the SNS topic.

#### When a message is published to an SNS topic, the message is forwarded to all SQS queues that are subscribed to that topic. When a message is received on an SQS queue, it is treated as a separate copy of the message, independent from any other copies of the message that may have been delivered to other queues.

#### So in summary, to read messages in a pub-sub model using SNS and SQS, you need to read messages from the subscribed SQS queue, not from the SNS topic directly.
