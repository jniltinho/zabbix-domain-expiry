package output

import (
	"encoding/json"
	"fmt"
	"os"
)

const Version = "2.0.0"

const (
	StateOK       = 0
	StateWarning  = 1
	StateCritical = 2
	StateUnknown  = 3
)

type Result struct {
	State            string `json:"state"`
	DaysLeft         int    `json:"days_left"`
	DaysSinceExpired int    `json:"days_since_expired"`
	ExpireDate       string `json:"expire_date"`
	Message          string `json:"message"`
}

type VersionInfo struct {
	Version string `json:"version"`
}

func StateName(code int) string {
	switch code {
	case StateOK:
		return "OK"
	case StateWarning:
		return "WARNING"
	case StateCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func Exit(code int, daysLeft int, expireDate, message string) {
	daysSinceExpired := 0
	if daysLeft < 0 {
		daysSinceExpired = -daysLeft
	}

	if expireDate == "" {
		expireDate = "unknown"
	}

	result := Result{
		State:            StateName(code),
		DaysLeft:         daysLeft,
		DaysSinceExpired: daysSinceExpired,
		ExpireDate:       expireDate,
		Message:          message,
	}

	data, err := json.Marshal(result)
	if err != nil {
		fmt.Fprint(os.Stdout, `{"state":"UNKNOWN","days_left":0,"days_since_expired":0,"expire_date":"unknown","message":"State: UNKNOWN ; Failed to encode JSON"}`)
		os.Exit(0)
	}

	// Zabbix external checks treat any non-zero exit code as failure (NOT SUPPORTED),
	// even when valid JSON is printed. Domain status is conveyed via the JSON "state" field.
	fmt.Fprint(os.Stdout, string(data))
	os.Exit(0)
}

func PrintVersion() {
	data, _ := json.Marshal(VersionInfo{Version: Version})
	fmt.Println(string(data))
}