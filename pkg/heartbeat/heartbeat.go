package heartbeat

import (
	"fmt"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/initapp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This package is currently writing the HBs in the same app database
// In the future it could be helpful to send the pulse to a monitoring system at least in addition to the current local db
// The local db could still be useful if we need that info handy or BI tooling but for a status monitoring/alerting system it would be beneficial to rely on an external service.

type HeartBeatType struct {
	hostName              string
	hostNameFriendly      string
	appRole               string
	region                string
	subRegion             string
	lastHBTime            time.Time
	uptimeTime            time.Duration
	uptimeSinceTime       time.Time
	uptimeSinceHuman      string
	uptimeSinceUnix       int64
	pulseSequence         int64
	pulseFrequency        time.Duration
	pulseFrequencySeconds int64
	chDone                chan bool
	pulsing               bool
	dbClient              *mongo.Client
	dbName                string
	dbCollection          string
	publicIp              string
	statusListeningPort   string
	appVersion            [][]string
}

type HeartBeatDBType struct {
	HostName              string     `bson:"hostname"`
	HostNameFriendly      string     `bson:"hostnamefriendly"`
	AppRole               string     `bson:"approle"`
	Region                string     `bson:"region"`
	SubRegion             string     `bson:"subregion"`
	LastHBUnix            int64      `bson:"lasthbunix"`
	LastHB                string     `bson:"lasthb"`
	UptimeSeconds         int64      `bson:"uptimeseconds"`
	UptimeHuman           string     `bson:"uptimehuman"`
	UptimeSinceHuman      string     `bson:"uptimesincehuman"`
	UptimeSinceUnix       int64      `bson:"uptimesinceunix"`
	PulseSequence         int64      `bson:"pulsesequence"`
	PulseFrequencySeconds int64      `bson:"pulsefrequencyseconds"`
	PublicIp              string     `bson:"publicip"`
	StatusListeningPort   string     `bson:"statuslisteningport"`
	AppVersion            [][]string `bson:"appversion"`
}

func (hb *HeartBeatType) Stop() {
	hb.chDone <- true
	hb.pulsing = false
}

func (hb *HeartBeatType) Start() {
	if hb.pulsing {
		return
	}
	hb.pulsing = true
	hb.sendPulse() // send a pulse immediately
	go hb.pulseRoutine()
}

func (hb *HeartBeatType) pulseRoutine() {
	ticker := time.NewTicker(hb.pulseFrequency)
	for {
		select {
		case <-hb.chDone:
			return
		case <-ticker.C:
			hb.sendPulse()
		}
	}
}

func (hb *HeartBeatType) sendPulse() {
	dbRecord := HeartBeatDBType{}

	// update values in hb client
	hb.pulseSequence++
	hb.lastHBTime = time.Now()
	hb.uptimeTime = time.Since(hb.uptimeSinceTime)

	// create the db record
	dbRecord.HostName = hb.hostName
	dbRecord.HostNameFriendly = hb.hostNameFriendly
	dbRecord.AppRole = hb.appRole
	dbRecord.Region = hb.region
	dbRecord.SubRegion = hb.subRegion
	dbRecord.LastHBUnix = hb.lastHBTime.Unix()
	dbRecord.LastHB = hb.lastHBTime.Format(time.RFC850)
	dbRecord.UptimeSeconds = hb.uptimeTime.Milliseconds() / 1000
	dbRecord.UptimeHuman = hb.uptimeTime.String()
	dbRecord.UptimeSinceHuman = hb.uptimeSinceHuman
	dbRecord.UptimeSinceUnix = hb.uptimeSinceUnix
	dbRecord.PulseSequence = hb.pulseSequence
	dbRecord.PulseFrequencySeconds = hb.pulseFrequencySeconds
	dbRecord.PublicIp = hb.publicIp
	dbRecord.StatusListeningPort = hb.statusListeningPort
	dbRecord.AppVersion = hb.appVersion

	t := true
	opts := options.UpdateOptions{}
	opts.Upsert = &t
	err := dbhelper.UpdateRecord(hb.dbClient, hb.dbName, hb.dbCollection, bson.M{"hostname": dbRecord.HostName, "approle": dbRecord.AppRole}, bson.M{"$set": dbRecord}, &opts)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func New(hostName string, hostNameFriendly string, appRole string, region string, subRegion string, frequency time.Duration, dbClient *mongo.Client, dbName string, dbCollection string, statusListeningPort string, publicIp string) *HeartBeatType {
	hb := HeartBeatType{}

	hb.dbClient = dbClient
	hb.dbName = dbName
	hb.dbCollection = dbCollection

	hb.hostName = hostName
	hb.hostNameFriendly = hostNameFriendly
	hb.appRole = appRole
	hb.region = region
	hb.subRegion = subRegion
	hb.uptimeSinceTime = time.Now()
	hb.pulseFrequency = frequency
	hb.publicIp = publicIp
	hb.statusListeningPort = statusListeningPort
	hb.appVersion = initapp.GetAppVersion()

	hb.uptimeSinceUnix = hb.uptimeSinceTime.Unix()
	hb.uptimeSinceHuman = hb.uptimeSinceTime.Format(time.RFC850)
	hb.pulseFrequencySeconds = hb.pulseFrequency.Milliseconds() / 1000 // used milliseconds() because seconds() returns a float64, and I'm too lazy

	return &hb
}
