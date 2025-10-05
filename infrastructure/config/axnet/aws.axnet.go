package axnet

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// AWSHttpTransport returns a Transport that's compatible with the AWS cloud environment
func AWSHttpTransport(maxIdleConns *int, idleConnTimeout *time.Duration) *http.Transport {
	var mic int = 10
	if maxIdleConns != nil {
		mic = *maxIdleConns
	}
	var ict time.Duration = 30
	if idleConnTimeout != nil {
		ict = *idleConnTimeout
	}
	return &http.Transport{
		MaxIdleConns:       mic,
		IdleConnTimeout:    ict * time.Second,
		DisableCompression: true,
		DialTLSContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: true})
		},
	}
}

// AWSHttpClient returns a Client that's compatible with the AWS cloud environment.
// See `AWSHttpTransport()`
func AWSHttpClient(maxIdleConns *int, idleConnTimeout *time.Duration) *http.Client {
	return &http.Client{Transport: AWSHttpTransport(maxIdleConns, idleConnTimeout)}
}

func PublishMessageSNS(
	topicArn, subject string,
	messageAttributes map[string]*sns.MessageAttributeValue,
	messageData []byte,
) error {
	cfg := &aws.Config{
		HTTPClient: AWSHttpClient(nil, nil),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *cfg,
	}))

	svc := sns.New(sess)

	if _, err := svc.Publish(&sns.PublishInput{
		TopicArn:          aws.String(topicArn),
		Subject:           aws.String(subject),
		Message:           aws.String(string(messageData)),
		MessageAttributes: messageAttributes,
	}); err != nil {
		return err
	}

	return nil
}

func WriteMessageSQS(sqsUrl string, messageData []byte) error {
	session := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sqs.New(session)

	if _, err := svc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &sqsUrl,
		MessageBody: aws.String(string(messageData)),
	}); err != nil {
		return err
	}

	return nil
}
