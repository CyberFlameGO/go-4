package test

import (
	"flag"
	"log"
	"testing"
	"time"

	"bou.ke/monkey"
)

func init() {
	if flag.Lookup("test.v") == nil {
		log.Fatal("package go/test should not be used outside unit tests")
	}
}

var fakeNow = time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC)
var ticker *time.Ticker

func fakeTimer(start time.Time, ticker *time.Ticker, increment time.Duration) {
	fakeNow = start
	go func() {
		for range ticker.C {
			fakeNow = fakeNow.Add(increment)
		}
	}()
}

func FakeTime(multiplier int) {
	ticker = time.NewTicker(time.Millisecond)

	fakeTimer(time.Now().AddDate(0, -1, 0), ticker, 24*time.Hour/100)

}
func TestMP(t *testing.T) {

	second := time.NewTicker(100 * time.Millisecond)

	// Set up time to go at approximately 10 days/second.

	monkey.Patch(time.Now, func() time.Time { return fakeNow })
	log.Println(time.Now())

	for i := 0; i < 10; i++ {
		<-second.C
		log.Println(time.Now())
	}

	monkey.Unpatch(time.Now)
	log.Println(time.Now())

	t.Fail()
}
