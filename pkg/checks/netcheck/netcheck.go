package netcheck

import (
	"brainyping/pkg/queueHelper"
)

func ProcessRequest(check *queueHelper.CheckRecordQueued) {
	switch check.Record.SubType {
	case "NET":
		break
	default:

	}
	return

}
