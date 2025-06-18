package browser

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ServiceConfig .
type ServiceConfig struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Method string `json:"method,omitempty"`
	Body   any    `json:"body,omitempty"`
}

type socketResponse struct {
	NumSockets uint `json:"ClientCount"`
}

const configPath = "/home/pi/.i3/config"

// CheckWebSocketCount checks the web socket count and restarts the browser if the count is 0
func CheckWebSocketCount(ctx context.Context, configs []ServiceConfig) (bool, error) {
	for _, config := range configs {
		if config.Method == "" {
			config.Method = "GET"
		}
		resp, err := makeRequest(config.Method, config.URL)
		if err != nil {
			return false, err
		}

		if resp.NumSockets == 0 {
			err = restartBrowser()
			return true, err
		}
	}
	return false, nil
}

func makeRequest(method, url string) (*socketResponse, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("bad response code - %v: %s", resp.StatusCode, b)
	}

	var responseBody socketResponse
	err = json.Unmarshal(b, &responseBody)
	if err != nil {
		return nil, err
	}

	return &responseBody, nil
}

func restartBrowser() error {
	err := killChromium()
	// if err != nil {
	// 	return err
	// }

	//find i3 config file
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}

	fileIn := bufio.NewReader(file)

	//get the last line
	lastLine, err := readLastLine(fileIn)
	if err != nil {
		return err
	}
	lastLine = strings.TrimSuffix(lastLine, "\n")

	//parse the last line
	regex, err := regexp.Compile(` \S*\'.*\'| \S+`)
	if err != nil {
		return err
	}
	tokens := regex.FindAllString(lastLine, -1)
	for i, tok := range tokens {
		tokens[i] = strings.TrimPrefix(tok, " ")
	}

	command := []string{"sudo", "-H", "-u", "pi"}
	command = append(command, tokens...)

	args := command[1:]

	//run the last line as a bash command with exec
	cmd := exec.Command(command[0], args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DISPLAY=:0")

	err = cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func killChromium() error {
	cmd := exec.Command("pkill", "-o", "chromium")
	return cmd.Run()
}

func readLastLine(r *bufio.Reader) (string, error) {
	lastLine := ""
	line, err := r.ReadString('\n')
	for err == nil {
		if len(line) > 0 {
			lastLine = line
		}
		line, err = r.ReadString('\n')
	}
	if err != nil && err != io.EOF {
		return "", err
	}
	return lastLine, nil
}
