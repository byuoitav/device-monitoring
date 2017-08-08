package microservicestatus

import (
	"bufio"
	"os"
)

const (
	StatusOK   = 0
	StatusSick = 1
	StatusDead = 2
)

type Status struct {
	Version    string `json:"version"`
	Status     int    `json:"statuscode"`
	StatusInfo string `json:"statusinfo"`
}

func GetVersion(fileLocation string) (string, error) {
	file, err := os.Open(fileLocation)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // only read first line
	if err := scanner.Err(); err != nil {
		return "", err
	}

	version := scanner.Text()

	return version, nil
}
