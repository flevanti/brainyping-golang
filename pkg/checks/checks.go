package checks

import (
	"time"

	"brainyping/pkg/checks/httpcheck"
	"brainyping/pkg/dbhelper"
	"brainyping/pkg/queuehelper"
)

func ProcessCheckFromQueue(check *queuehelper.CheckRecordQueued) error {
	var checkResponse dbhelper.CheckOutcomeRecord
	var err error
	var checkStart time.Time = time.Now()
	switch check.Record.Type {
	case "HTTP":
		checkResponse, err = httpcheck.ProcessCheck(check.Record.Host, check.Record.SubType, check.Record.UserAgent)
		break
	case "NET":
		// netcheck.ProcessRequest(check)
		break
	default:
	}

	if err != nil {
		return err
	}

	checkResponse.CreatedUnix = time.Now().Unix()
	checkResponse.TimeSpent = time.Since(checkStart).Microseconds()
	check.RecordOutcome = checkResponse

	return nil
}

func ProcessHTTPCheckFromCli(subType string, url string, userAgent string) (dbhelper.CheckOutcomeRecord, error) {
	var checkResponse dbhelper.CheckOutcomeRecord
	var err error
	var checkStart time.Time = time.Now()

	checkResponse, err = httpcheck.ProcessCheck(url, subType, userAgent)
	if err != nil {
		return dbhelper.CheckOutcomeRecord{}, err
	}
	checkResponse.CreatedUnix = time.Now().Unix()
	checkResponse.TimeSpent = time.Since(checkStart).Microseconds()

	return checkResponse, nil

}
