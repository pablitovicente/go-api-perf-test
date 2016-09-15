package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type requestsStats struct {
	totalOk               int
	totalNotOk            int
	medianTimeAllRequests float64
	totalBytesTransferred int64
	allRequestTimes       []float64
}

// The shared network client
var netClient = &http.Client{
	Timeout: time.Second * 4,
}

// Sync wait group so Main waits for all go routines to complete
var wg sync.WaitGroup

// For locking access to the stats
var statsLock = &sync.Mutex{}

// The Function that will take care of making all the request
func fetchURL(url string, theStats *requestsStats) {
	start := time.Now()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		//log.Fatalln(err)
		fmt.Println("Requester Instantiation failed....")
	}
	req.Header.Set("User-Agent", "Load-Testing-0.1")

	res, err := netClient.Do(req)
	end := time.Since(start).Seconds()
	defer wg.Done()

	if err != nil {
		//fmt.Println("**********************************************************")
		//fmt.Println(err)
		//fmt.Println("Requester failed....")
		//fmt.Println("**********************************************************")
		statsLock.Lock()
		defer statsLock.Unlock()
		theStats.totalNotOk++
		fmt.Print("-")
	} else {
		//fmt.Println(res.StatusCode)
		//fmt.Println(res.ContentLength)
		//fmt.Println(end)
		//fmt.Println("**********************************************************")
		fmt.Print("+")
		statsLock.Lock()
		defer statsLock.Unlock()
		theStats.allRequestTimes = append(theStats.allRequestTimes, end)
		if res.StatusCode == 200 {
			theStats.totalOk++
			theStats.totalBytesTransferred += res.ContentLength
		} else {
			theStats.totalNotOk++
		}
		// if err != nil {
		// 	//log.Fatalln(err)
		// 	fmt.Println("Some other error....")
		// }
	}
}

func calculateMedianTimes(theStats *requestsStats) {
	sort.Float64s(theStats.allRequestTimes)
	middle := len(theStats.allRequestTimes) / 2
	result := theStats.allRequestTimes[middle]
	if len(theStats.allRequestTimes)%2 == 0 {
		result = (result + theStats.allRequestTimes[middle-1]) / 2
	}
	theStats.medianTimeAllRequests = result
}

func validateURL(url string) {
	req, err := http.NewRequest("GET", url, nil)
	res, err := netClient.Do(req)
	if err != nil || res.StatusCode == 404 {
		fmt.Printf("The Url %s is not valid or is not responding, the test can't continue.", url)
		os.Exit(1)
	}
}

func main() {
	var theStats requestsStats

	targetURLAndNumberOfHits := os.Args[1:]
	totalRequestsToMake, _ := strconv.ParseInt(targetURLAndNumberOfHits[1], 10, 0)
	targetURL := targetURLAndNumberOfHits[0] //"http://127.0.0.1:3000/index.html"
	validateURL(targetURL)
	if targetURLAndNumberOfHits[1] == "" {
		fmt.Println("You need to specify the number of requests to execute")
		os.Exit(1)
	}

	// Start counting time
	start := time.Now()
	for i := 0; i < int(totalRequestsToMake); i++ {
		// Increment the WaitGroup counter.
		wg.Add(1)
		go fetchURL(targetURL, &theStats)
	}

	// Wait for all HTTP fetches to complete.
	wg.Wait()
	// Once that all routines are done calculate how much it took
	timeItTookToFinishAllRequests := time.Since(start).Seconds()
	// Calculate the Median of all request times
	calculateMedianTimes(&theStats)

	fmt.Println("")
	fmt.Println("")
	fmt.Println("************************************************************************************************************************")
	fmt.Println("************************************************************************************************************************")
	fmt.Println("Total OK requests: ")
	fmt.Println(theStats.totalOk)
	fmt.Println("Total NOT OK requests: ")
	fmt.Println(theStats.totalNotOk)
	fmt.Println("Total Time for finishing all requests: ")
	fmt.Println(timeItTookToFinishAllRequests)
	//fmt.Println("All the times array: ")
	//fmt.Println(theStats.allRequestTimes)
	//fmt.Println(len(theStats.allRequestTimes))
	fmt.Println("Median time to complete OK requests: ")
	fmt.Println(theStats.medianTimeAllRequests)
	fmt.Println("Total Transferred Mega Bytes:")
	fmt.Println(theStats.totalBytesTransferred / 1024 / 1024)
	fmt.Println("************************************************************************************************************************")
	fmt.Println("************************************************************************************************************************")

}
