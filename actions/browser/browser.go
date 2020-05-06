package browser

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ServiceConfig .
type ServiceConfig struct {
	Name   string      `json:"name"`
	URL    string      `json:"url"`
	Method string      `json:"method"`
	Body   interface{} `json:"body,omitempty"`
}

const configPath = "/home/pi/.i3/config"

// CheckWebSocketCount checks the web socket count and restarts the browser if the count is 0
func CheckWebSocketCount(ctx context.Context, configs []ServiceConfig) (bool, error) {
	// talk to the provided service and get websocket status
	// if count is 0
	// do the chrome restart
	// else do nothing

	return false, nil
}

func restartBrowser() error {
	err := killChromium()
	if err != nil {
		return err
	}

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
	args := tokens[1:]

	//run the last line as a bash command with exec
	cmd := exec.Command(tokens[0], args...)
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
		// fmt.Printf("line - %s\n", line)
	}
	if err != nil && err != io.EOF {
		return "", err
	}
	return lastLine, nil
}
