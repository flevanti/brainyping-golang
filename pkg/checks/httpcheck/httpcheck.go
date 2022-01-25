package httpcheck

import (
	"awesomeProject/pkg/dbHelper"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

var userAgent string = "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148" //this will trick the server to think we are a browser adding some more redirects sometimes

func ProcessCheck(url string, method string, subType string) (dbHelper.CheckOutcomeRecord, error) {
	var outcome dbHelper.CheckOutcomeRecord
	var err error

	switch subType {
	case "GET":
	case "HEAD":
		outcome, err = subTypeGetHead(url, subType)
		break
	//case "HEAD":
	//	subTypeGetHead(url, "HEAD")
	//	break
	case "ROBOTSTXT":
		outcome, err = subTypeRobotstxt(url)
		break
	default:
		err = errors.New("method subtype not correct")
	}

	return outcome, err
}

func subTypeGetHead(url string, method string) (dbHelper.CheckOutcomeRecord, error) {
	var err error
	var cookieJar *cookiejar.Jar
	var client http.Client
	var request *http.Request
	var response *http.Response
	var timeout = time.Duration(5 * time.Second)
	var returnedValue dbHelper.CheckOutcomeRecord

	cookieJar, err = cookiejar.New(nil)
	if err != nil {
		returnedValue.ErrorInternal = "Error while preparing http cookie jar: " + err.Error()
		returnedValue.ErrorOriginal = err.Error()
		returnedValue.ErrorFriendly = "Error while preparing HTTP cookie jar for request"
		returnedValue.Message = returnedValue.ErrorFriendly
		return returnedValue, err
	}
	client = http.Client{
		Timeout: timeout,
		Jar:     cookieJar,
	}
	request, err = http.NewRequest(method, url, nil)
	if err != nil {
		returnedValue.ErrorInternal = "Error while preparing http request: " + err.Error()
		returnedValue.ErrorOriginal = err.Error()
		returnedValue.ErrorFriendly = "Error while preparing HTTP request"
		returnedValue.Message = returnedValue.ErrorFriendly
		return returnedValue, err
	}

	//request.Header.Set("User-Agent", userAgent)

	response, err = client.Do(request)

	if err != nil {
		returnedValue.ErrorInternal = "Error while performing http call: " + err.Error()
		returnedValue.ErrorOriginal = err.Error()
		returnedValue.ErrorFriendly = "Error while performing HTTP request"
		returnedValue.Message = returnedValue.ErrorFriendly
		return returnedValue, err
	}

	returnedValue.Message = fmt.Sprintf("%s ||%d", response.Status, response.StatusCode)

	redirectionsToListRecursive(response, &returnedValue.RedirectsHistory)
	returnedValue.Redirects = len(returnedValue.RedirectsHistory)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		returnedValue.ErrorOriginal = fmt.Sprintf("Status code not 2xx but %s", response.Status)
		returnedValue.ErrorInternal = returnedValue.ErrorOriginal
		returnedValue.ErrorFriendly = returnedValue.ErrorOriginal
		return returnedValue, errors.New(strings.ToLower(returnedValue.ErrorOriginal))
	}

	returnedValue.Success = true

	return returnedValue, nil
}

//func subTypeHead(url string) (bool, error) {
//	timeout := time.Duration(3 * time.Second)
//	client := http.Client{
//		Timeout: timeout,
//	}
//	resp, err := client.Head(url)
//	if err != nil {
//		return false, err
//	}
//	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
//		return false, errors.New(fmt.Sprintf("Status code returner not 2xx but %d", resp.StatusCode))
//	}
//	return true, nil
//}

func subTypeRobotstxt(url string) (dbHelper.CheckOutcomeRecord, error) {
	return dbHelper.CheckOutcomeRecord{}, nil
	//timeout := time.Duration(3 * time.Second)
	//client := http.Client{
	//	Timeout: timeout,
	//}
	//resp, err := client.Head(url)
	//if err != nil {
	//	return false, err
	//}
	//if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	//	return false, errors.New(fmt.Sprintf("Status code returner not 2xx but %d", resp.StatusCode))
	//}
	//return true, nil
}

func redirectionsToListRecursive(resp *http.Response, history *[]dbHelper.RedirectHistory) {
	var historyElement dbHelper.RedirectHistory

	if resp.Request.Response != nil {
		redirectionsToListRecursive(resp.Request.Response, history)
	}
	//extract information we want
	historyElement.URL = resp.Request.URL.String()
	historyElement.Status = resp.Status
	historyElement.StatusCode = resp.StatusCode

	//append the element to the redirections history
	*history = append(*history, historyElement)

}
