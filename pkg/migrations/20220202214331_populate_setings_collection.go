package migrations

import (
	"brainyping/pkg/dbhelper"

	"github.com/flevanti/bisonmigration"
	"go.mongodb.org/mongo-driver/mongo"
)

//
// Please return an error if you want the migration to fail and the migration process to stop.
// Migration failed will continue to be pending ( or won't be rolled back if it was a down process)
// Don't exit, panic or try any other way to stop the process.
//
// just return a nice error
//

type SettingType struct {
	Key   string `bson:"key"`
	Value string `bson:"value"`
	Descr string `bson:"descr"`
}

func getListOfSettings() []SettingType {
	userAgent := "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"
	settings := []SettingType{
		{Key: "WRK_HTTP_TIMEOUT_MS", Value: "10000", Descr: "Timeout used during http checks"},
		{Key: "WRK_GOROUTINES", Value: "30", Descr: "number of goroutines to start for each worker"},
		{Key: "WRK_THROTTLE_RPS", Value: "50", Descr: "limit the number of request that each worker performs every second"},
		{Key: "WRK_BUF_CH_SIZE", Value: "100", Descr: "size of buffered channel used to pass requests from the queue to the goroutines"},
		{Key: "WRK_GRACE_PERIOD_MS", Value: "10000", Descr: "time since last request processed that each go routine wait before stopping. Used for gracefully stopping the application"},
		{Key: "WRK_WRKS_READY_TIMEOUT_MS", Value: "15000", Descr: "time for all the goroutines started to be ready to work"},
		{Key: "WRK_HTTP_USER_AGENT", Value: userAgent, Descr: "user agent used during HTTP requests"},
		{Key: "QUEUE_PREFETCH_COUNT", Value: "100", Descr: "number of messages to consume each request"},
		{Key: "QUEUENAME_REQUEST", Value: "brainypingqueue", Descr: "queue name used for sending/receiving checks requests"},
		{Key: "QUEUENAME_RESPONSE", Value: "brainypingresponsequeue", Descr: "queue name used for sending/receiving processed checks responses "},
		{Key: "BL_RPS_SPREAD", Value: "10", Descr: "number of requests per seconds spread each second. this is used when bulk loading checks to avoid having all the checks/requests happening in the same second"},
		{Key: "BL_OWNERUID", Value: "BL-OWNERUID", Descr: "user id of the owner of the check loaded during a bulk load"},
		{Key: "BL_BULK_SAVE_SIZE", Value: "1000", Descr: "number of records to include in each saving operation"},
		{Key: "RC_BUF_CH_SIZE", Value: "100", Descr: "size of the buffered channel used to pass responses from the queue consumer to the response collector process"},
		{Key: "RC_GRACE_PERIOD_MS", Value: "5000", Descr: "time to wait before closing the application to make sure all buffered responses are saved"},
		{Key: "RC_BULK_SAVE_SIZE", Value: "100", Descr: "number of records to include in each saving operation"},
		{Key: "RC_SAVE_AUTO_FLUSH_MS", Value: "10000", Descr: "max time waited before buffered responses are saved even if the limit is not reached. This is to avoid having records only staged in memory for too long"},
	}

	return settings
}

func getSettingsInterface() []interface{} {
	var recordsIntf []interface{}
	for _, v := range getListOfSettings() {
		recordsIntf = append(recordsIntf, v)
	}
	return recordsIntf
}

func up_20220202214331(db *mongo.Client) error {
	settings := getSettingsInterface()

	return dbhelper.SaveManyRecords(dbhelper.GetDatabaseName(), dbhelper.TablenameSettings, &settings)
}

func down_20220202214331(db *mongo.Client) error {
	for _, v := range getListOfSettings() {
		_, err := dbhelper.DeleteRecordsByFieldValue(dbhelper.GetDatabaseName(), dbhelper.TablenameSettings, "key", v.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

//
//
// DON'T TOUCH ANYTHING BEYOND THIS POINT
//
//

//
// this is adding the migration to the migration engine
//
func init() {
	bisonmigration.RegisterMigration(20220202214331, "populate_settings_collection", "*DEFAULT*", up_20220202214331, down_20220202214331)
}
