/*
This file contains code that can be dropped into an AWS based project to monitor and advise if a spot instance
is going to be terminated soon. AWS set metadata 2 minutes before they terminate a spot instance.

*/
package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

// uncomment to test compile this file
//func main() {}

// NewSpotter creates and starts the spot instance termination check running and returns
// a chan that will receive notification that this instance is about to be
// terminated. If the done chan is supplied it will shutdown if anything is
// received on this chan
func NewSpotter(done chan struct{}) chan struct{} {
	c := make(chan struct{})
	// start the spot termination checker running
	go run(done, c)
	return c
}

// run runs in a goroutine
func run(done chan struct{}, c chan struct{}) {

	termChan := make(chan bool)
	tick := time.NewTicker(time.Second * 15)

	dowhile := true
	for dowhile == true {
		select {
		case <-tick.C:
			// check if this spot instance is terminating soon
			go spotRunning(termChan)
		case dowhile = <-termChan:
			// if anything received on the termChan then exit the for loop and send term signal
		case <-done:
			// done channel closed so exit the select
			dowhile = false
		}
	}
	tick.Stop()
	c <- struct{}{}
}

// spotRunning runs in a goroutine and returns false if the spot instance
// is about to be terminated
func spotRunning(termChan chan bool) {
	// get our external ip address so we can add it to the results
	c := ec2metadata.New(session.New())

	_, err := c.GetMetadata("spot/termination-time")
	if err == nil {
		// normally this option is not present so returns an error
		// non err so spot instance is terminating soon
		termChan <- false
	}
}

/*

 */
