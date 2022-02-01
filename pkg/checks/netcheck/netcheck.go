package netcheck

import (
	"brainyping/pkg/queuehelper"
)

func ProcessRequest(check *queuehelper.CheckRecordQueued) {
	switch check.Record.SubType {
	case "NET":
		break
	default:

	}
	return

}
