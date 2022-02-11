package settings

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"brainyping/pkg/dbhelper"
	"brainyping/pkg/utilities"
)

func GetSettStr(key string) string {
	v, _ := GetEnv(key)
	return v
}

func GetSettDuration(key string) time.Duration {
	var d time.Duration
	var i int
	var err error
	s, b := GetEnv(key)
	if b {
		i, err = strconv.Atoi(s)
		utilities.FailOnError(err)
		d = time.Duration(i)
	} else {
		i = 0
	}
	return d
}

func GetSettInt(key string) int {
	var i int
	var err error
	s, b := GetEnv(key)
	if b {
		i, err = strconv.Atoi(s)
		utilities.FailOnError(err)
	} else {
		i = 0
	}
	return i
}

func GetSettInt64(key string) int64 {
	var i int64
	var err error
	s, b := GetEnv(key)
	if b {
		i, err = strconv.ParseInt(s, 10, 64)
		utilities.FailOnError(err)
	} else {
		i = 0
	}
	return i
}

func GetEnv(key string) (string, bool) {
	v, b := os.LookupEnv(key)
	// we are strict, if key doesn't exist fail.... maybe we want to log this and continue... ?
	if !b {
		utilities.FailOnError(errors.New(fmt.Sprintf("Env variable key [%s] not found while looking for setting key", key)))
	}
	return v, b
}

func SettExists(key string) bool {
	_, b := os.LookupEnv(key)
	return b
}

func SaveNewSett(record dbhelper.SettingType) {
	err := dbhelper.SaveRecord(dbhelper.GetDatabaseName(), dbhelper.TablenameSettings, record)
	utilities.FailOnError(err)
}

func DeleteSettingByKey(key string) {
	_, err := dbhelper.DeleteRecordsByFieldValue(dbhelper.GetDatabaseName(), dbhelper.TablenameSettings, "key", key)
	utilities.FailOnError(err)
}
