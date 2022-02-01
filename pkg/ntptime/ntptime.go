package ntptime

import (
	"github.com/beevik/ntp"
	"time"
)

var timeOffset time.Duration
var timeInitialised time.Time
var timeInitialisedNtp time.Time
var ntpServer = "time.google.com"
var initialised bool

func Now() time.Time {
	return time.Now().Add(timeOffset)
}
func GetOffset() time.Duration {
	return timeOffset
}
func GetTimeInitialised() time.Time {
	return timeInitialised
}
func GetTimeInitialisedNtp() time.Time {
	return timeInitialisedNtp
}
func GetNtpServer() string {
	return ntpServer
}

func getInitialisedFlag() bool {
	return initialised
}

func init() {
	initialise()
}

func initialise() {
	if initialised {
		return
	}
	timeInitialised = time.Now()
	ntpTime, _ := ntp.Query(ntpServer)
	timeInitialisedNtp = ntpTime.Time
	timeOffset = ntpTime.ClockOffset
	initialised = true
	return
}
