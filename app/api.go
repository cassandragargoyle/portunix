package app

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Info struct to represent the JSON structure
type SHInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Vendor  string `json:"vendor"`
}

func DaemonGetInfoJson(address string, port string) (string, error) {
	url := getUrl(address, port)
	url += "/rest/info"

	response, err := SendGet(url)

	if err == nil {
		if false {
			fmt.Println(response)
		}
	}
	return response, err
}

func DaemonGetInfo(address string, port string) (SHInfo, error) {
	var info SHInfo
	response, err := DaemonGetInfoJson(address, port)
	if IsHtml(response) {
		return info, errors.New("This is not SweeHome daemon server.")
	}
	err = json.Unmarshal([]byte(response), &info)
	if err != nil {
		fmt.Println("Error:", err)
		return info, err
	}
	return info, nil
}

func getUrl(address string, port string) string {
	var url string
	if len(port) > 0 {
		url = "http://" + address + ":" + port
	} else {
		url = "http://" + address
	}
	return url
}
