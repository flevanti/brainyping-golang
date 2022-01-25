package checks

import (
	"awesomeProject/pkg/checks/httpcheck"
	"awesomeProject/pkg/dbHelper"
	"awesomeProject/pkg/queueHelper"
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
