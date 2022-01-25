package utilities

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

func FailOnError(err error) {
	if err != nil {
		log.Fatalln("\n\n🔴 ", err.Error())
	}
}

func GetMemoryStats(unit string) map[string]string {
	var memStats runtime.MemStats
	var memStatsToReturn = make(map[string]string)
	var unitValue uint64
	runtime.ReadMemStats(&memStats)

	if unit == "AUTO" {
		switch true {
		case memStats.Alloc > 1024*1024*1024*1024:
			unit = "TB"
			break
		case memStats.Alloc > 1024*1024*1024:
			unit = "GB"
			break
		case memStats.Alloc > 1024*1024:
			unit = "MB"
			break
		case memStats.Alloc > 1024:
			unit = "KB"
			break
		default:
			unit = "B"
			break
		}
	}

	switch unit {
	case "B":
	case "":
		unitValue = 1 // 1
		break
	case "KB":
		unitValue = 1024 // 1024
		break
	case "MB":
		unitValue = 1024 * 1024 // 1,048,576
		break
	case "GB":
		unitValue = 1024 * 1024 * 1024 // 1,073,741,824
		break
	case "TB":
		unitValue = 1024 * 1024 * 1024 * 1024 // 1,099,511,627,776
		break
	default:
		log.Fatalf("Unknown memory unit requested: [%s]", unit)
	}

	memStatsToReturn["Alloc"] = strconv.FormatUint(memStats.Alloc/unitValue, 10)
	memStatsToReturn["TotalAlloc"] = strconv.FormatUint(memStats.TotalAlloc/unitValue, 10)
	memStatsToReturn["Sys"] = strconv.FormatUint(memStats.Sys/unitValue, 10)
	memStatsToReturn["AllocUnit"] = strconv.FormatUint(memStats.Alloc/unitValue, 10) + unit
	memStatsToReturn["TotalAllocUnit"] = strconv.FormatUint(memStats.TotalAlloc/unitValue, 10) + unit
	memStatsToReturn["SysUnit"] = strconv.FormatUint(memStats.Sys/unitValue, 10) + unit
	memStatsToReturn["NumGC"] = strconv.FormatUint(memStats.Sys/unitValue, 10)
	memStatsToReturn["Unit"] = unit

	return memStatsToReturn

}

func LineCounter(r io.Reader) (int, error) {
	buf := make([]byte, bufio.MaxScanTokenSize)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil
		case err != nil:
			return count, err
		}
	}
}

func generateNotSeriousId() string {
	//to implement some funny id!
	return "NotYetAvailable🙂"
}

func GenerateId() string {
	//wrapper, one day we will implement a serious id generator...
	return generateNotSeriousId()
}

func RetrieveHostName() string {
	h, _ := os.Hostname()
	return h
}

func RetrievePublicIP() string {
	req, err := http.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer req.Body.Close()
	ip, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return ""
	}

	return string(ip)

}
