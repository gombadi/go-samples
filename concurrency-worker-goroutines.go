/*

This file contains a code snippet containing a workflow -
- read a large number of items into an incoming channel (fileChan)
- start x goroutines to pull items from the chan, process & put into results channel (resChan)
- final goroutine to read resChan and store in a structure
- after all the work is done pass the results back to the caller

*/
package main

/*
Code Snippet only so has not been test compiled for bug checking !!!
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// fillMap starts goroutines to download files from s3 and produce the map with all events
func fillMap(s3svc *s3.S3, s3Files []*s3.Object) (map[string]*AuthEvent, error) {

	// chan to stort incoming work with buffer of 25
	fileChan := make(chan string, 25)
	resChan := make(chan *AuthEvent)
	doneChan := make(chan struct{})

	// struct to store final result from system
	tm := make(map[string]*AuthEvent)

	var wg sync.WaitGroup

	// start the worker goroutines
	for x := 0; x <= 3; x++ {
		wg.Add(1)
		go processS3File(fileChan, resChan, &wg)
	}

	// start the results goroutine running
	go processResults(resChan, doneChan, tm)

	// fill the work chan with file names from the provided slice
	for _, s3File := range s3Files {
		// will fill upto the buffer size (25) then block and
		// allow other goroutines to start running
		fileChan <- *s3File.Key
	}

	// all filenames are in the chan so close it and let the goroutines drain it
	close(fileChan)

	// wait for the goroutines to do their work and exit
	wg.Wait()

	// Now we know the workers are gone and therefore no more work will be sent on the resChan
	// this will exit the results processor range over the resChan
	close(resChan)
	// wait for the results goroutine to shutdown then return for reporting
	<-doneChan
	close(doneChan)

	// return the processed results
	return tm, nil
}

// processS3File runs in a goroutine and reads filenames from the fileChan, pulls down the file, extracts the events
// and then sends then on the resChan
func processS3File(fileChan chan string, resChan chan *AuthEvent, wg *sync.WaitGroup) {

	defer wg.Done()

	s3svc := s3.New(session.New())
	var b bytes.Buffer // A Buffer needs no initialization.

	// read filenames from chan until it is closed
	for s3FileName := range fileChan {

		params := &s3.GetObjectInput{
			Bucket: aws.String(s3bucket),
			Key:    aws.String(s3FileName),
		}
		getObjOutput, err := s3svc.GetObject(params)
		if err != nil {
			fmt.Printf("s3.GetObject err - file: %s err: %v\n", s3FileName, err)
			continue
		}

		b.ReadFrom(getObjOutput.Body)
		authEvents := bytes.Split(b.Bytes(), []byte("\n"))

		for _, jsonae := range authEvents {
			// last \n does not need unmarshal
			if len(jsonae) < 10 {
				continue
			}
			ae := &AuthEvent{}
			err := json.Unmarshal(jsonae, &ae)
			if err != nil {
				fmt.Printf("unmarshal fail:%v\n", err)
				fmt.Printf("unmarshal json:\n%s\n", jsonae)
			} else {
				// push the decoded struct into the results channel
				resChan <- ae
			}
		}
		// reset to a clean buffer for each file
		b.Reset()
	}
	// exit goroutine and defer wg.Done()
}

// processResults runs in a goroutine and reads results from the workers and updates a global map
func processResults(resChan chan *AuthEvent, doneChan chan struct{}, tm map[string]*AuthEvent) {

	// read things from the resChan and add to the global map
	for ae := range resChan {
		tm[ae.Hash] = ae
	}

	// tell the world we are done
	doneChan <- struct{}{}
}

/*

 */
