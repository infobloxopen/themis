package pip_service

import (
	"fmt"
)

var pIPServiceConnection = map[string]string {
	"mcafee-ts": "127.0.0.1:5368",
}

func GetPIPConnection(service string) (string, error) {
	conn, ok := pIPServiceConnection[service]
	if !ok {
		return "", fmt.Errorf("Cannot find PIP Service '%s'", service)
	}

	return conn, nil
}
