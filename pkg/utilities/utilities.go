package utilities

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

func FailOnError(err error) {
	if err != nil {
		fmt.Print("\n\n\n")
		pc, filename, line, _ := runtime.Caller(1)
		log.Printf("[error] %s[%s:%d] %v\n", runtime.FuncForPC(pc).Name(), filename, line, err)
		pc, filename, line, _ = runtime.Caller(2)
		log.Printf("[error] %s[%s:%d]\n", runtime.FuncForPC(pc).Name(), filename, line)
		pc, filename, line, _ = runtime.Caller(3)
		log.Printf("[error] %s[%s:%d]\n", runtime.FuncForPC(pc).Name(), filename, line)
		pc, filename, line, _ = runtime.Caller(4)
		log.Printf("[error] %s[%s:%d]\n", runtime.FuncForPC(pc).Name(), filename, line)

		log.Fatal("Bye bye")

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

func generateNotSeriousId() string {
	// to implement some funny id!
	return "NotYetAvailable🙂"
}

func GenerateId() string {
	// wrapper, one day we will implement a serious id generator...
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

func PrintTable(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.AppendBulk(data)
	table.Render()
}

func PrintTableOneColumn(headers string, data []string) {
	var dataProcessed [][]string
	for _, v := range data {
		dataProcessed = append(dataProcessed, []string{v})
	}
	PrintTable([]string{headers}, dataProcessed)
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J") // clear the screen...
}

func CalculateSpeedPerSecond(sampling time.Duration) func(int64) float64 {
	var lastTotal int64
	var lastTime = time.Now()
	var lastSpeed float64
	var samplingInternal = sampling
	funcToReturn := func(newTotal int64) float64 {
		timeElapsed := time.Since(lastTime)
		if timeElapsed < samplingInternal {
			return lastSpeed
		}
		delta := newTotal - lastTotal
		lastSpeed = float64(delta) / float64(timeElapsed) * float64(time.Second)
		lastTotal = newTotal
		lastTime = time.Now()
		return lastSpeed
	}
	return funcToReturn
}

func ReadUserInput(textToShow string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(textToShow, "> ")
	text, err := reader.ReadString('\n')
	FailOnError(err)

	return strings.Trim(text, " \n\t")
}

func ReadUserInputWithOptions(textToShow string, options []string, backCommand string) string {
	reader := bufio.NewReader(os.Stdin)
	backCommandText := ""
	backCommand = strings.Trim(backCommand, " ")
	if backCommand != "" {
		backCommandText = fmt.Sprintf("[`%s` to go back] ", backCommand)
	}
	for {
		fmt.Printf("%s %s>  ", textToShow, backCommandText)
		text, err := reader.ReadString('\n')
		FailOnError(err)
		text = strings.Trim(text, " \n\t")
		for _, v := range options {
			if text == v {
				return text
			}
		}
		if backCommand != "" && text == backCommand {
			return backCommand
		}
	}
}

func ReadUserInputConfirm(text string) bool {
	for {
		r := ReadUserInput(text)
		switch r {
		case "y", "Y", "yes", "YES", "Yes":
			return true
		case "n", "N", "no", "NO", "No":
			return false
		}
	}
}
