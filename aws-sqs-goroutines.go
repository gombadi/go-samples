/*

This file contains sample code that starts a number of goroutines to read messages out of AWS SQS.

The SQS message queue names have been changed so it will not run correctly

*/
package main

import (
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// main is the application start point
func main() {

	// pull all messges from sqs
	sess := session.New()
	sqssvc := sqs.New(sess)

	cparams := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String("https://sqs.us-west-2.amazonaws.com/123456789098/sqs-queue-name"),
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
		},
	}
	cresp, err := sqssvc.GetQueueAttributes(cparams)
	if err != nil {
		log.Printf("error reading count from sqs: %v\n", err)
		os.Exit(1)
	}
	log.Printf("message count: %s\n", *cresp.Attributes["ApproximateNumberOfMessages"])

	var wg sync.WaitGroup

	for x := 0; x < 5; x++ {
		go getMessages(x, sqssvc, &wg)
		wg.Add(1)
	}

	wg.Wait()

	// delete/purge sqs messages
	purgeparams := &sqs.PurgeQueueInput{
		QueueUrl: aws.String("https://sqs.us-west-2.amazonaws.com/123456789098/sqs-queue-name"), // Required
	}
	_, err = sqssvc.PurgeQueue(purgeparams)
	if err != nil {
		log.Printf("error reading from sqs: %v\n", err)
		os.Exit(1)
	}
	log.Printf("queue purged\n")
}

func getMessages(x int, sqssvc *sqs.SQS, wg *sync.WaitGroup) {
	defer wg.Done()

	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String("https://sqs.us-west-2.amazonaws.com/123456789098/sqs-queue-name"),
		MaxNumberOfMessages: aws.Int64(10),
		VisibilityTimeout:   aws.Int64(30),
		WaitTimeSeconds:     aws.Int64(0),
	}

	resp, err := sqssvc.ReceiveMessage(params)
	if err != nil {
		log.Printf("error reading from sqs: %v\n", err)
		os.Exit(1)
	}

	log.Printf("routine: %d count: %v\n", x, len(resp.Messages))
	for _, message := range resp.Messages {
		log.Printf("routine: %d body %s\n", x, *message.Body)
	}
}

/*


 */
