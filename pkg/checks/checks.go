package checks

import (
	"brainyping/pkg/checks/httpcheck"
	"brainyping/pkg/dbHelper"
	"brainyping/pkg/queueHelper"
	"time"
)

func ProcessCheck(check *queueHelper.CheckRecordQueued) error {
	var checkResponse dbHelper.CheckOutcomeRecord
	var err error
	var checkStart time.Time = time.Now()
	switch check.Record.Type {
	case "HTTP":
		checkResponse, err = httpcheck.ProcessCheck(check.Record.Host, check.Record.Type, check.Record.SubType)
		break
	case "NET":
		//netcheck.ProcessRequest(check)
		break
	default:
	}

	checkResponse.CreatedUnix = time.Now().Unix()
	checkResponse.TimeSpent = time.Since(checkStart).Microseconds()
	check.RecordOutcome = checkResponse

	return err

}
