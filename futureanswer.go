/*

This file contains a pattern using future answers.

Quite often you need multiple bits of info before you can ask the final question. If network
requests are involved then time can add up.

Ask the questions and wait for them to become available.

*/
package main

import (
	"fmt"
	"time"
)

// FutureAns is a future answer to a question being asked which includes
// the requested data and/or any errors during production of the answer
type FutureAns struct {
	d string // data to be returned
	e error  // any error encountered
}

func main() {

	ts := time.Now()
	two := make([]string, 0)

	// ask both questions at the same time and then wait for an answer
	oneChan := askQ1()
	twoChan := askQ2()

	// block until answer becomes available
	oneAns := <-oneChan

	if oneAns.e != nil {
		fmt.Printf("error received from ask1() %v\n", oneAns.e)
	}

	// answers to q2 will be available at about the same time as the answers to Q1
	for twoAns := range twoChan {
		if twoAns.e != nil {
			fmt.Printf("error received from ask2() %v\n", twoAns.e)
			continue
		}
		two = append(two, twoAns.d)
	}

	// use the future answer
	fmt.Printf("Results from Q1: %s err: %v\n", oneAns.d, oneAns.e)
	fmt.Printf("Number of results from Q2: %d\n", len(two))
	for k, v := range two {
		fmt.Printf("results %d from Q2: %s\n", k, v)
	}
	fmt.Printf("All done. Run time: %s\n", time.Since(ts).String())
}

func askQ1() chan *FutureAns {

	c := make(chan *FutureAns)

	go func() {
		// single use channel so close after we put our answer on chan
		defer close(c)

		// do some network requests that takes time
		time.Sleep(time.Second)
		c <- &FutureAns{d: "message from Q1", e: nil}

		// exit goroutine & defer close the chan
	}()
	return c
}

func askQ2() chan *FutureAns {

	c := make(chan *FutureAns)

	go func() {
		// single use channel so close after we put our answer on chan
		defer close(c)

		// do some network requests that takes time
		time.Sleep(time.Second)
		c <- &FutureAns{d: "message 1 from Q2 ", e: nil}
		c <- &FutureAns{d: "message 2 from Q2 ", e: fmt.Errorf("sorry there was an error getting this result\n")}
		c <- &FutureAns{d: "message 3 from Q2 ", e: nil}

		// exit goroutine & close the chan to exit the range
	}()
	return c
}

/*

 */
