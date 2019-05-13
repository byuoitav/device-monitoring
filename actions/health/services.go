package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

// ServiceCheckConfig .
type ServiceCheckConfig struct {
	Name   string      `json:"name"`
	URL    string      `json:"url"`
	Method string      `json:"method"`
	Body   interface{} `json:"body,omitempty"`
}

// ServiceCheckResponse .
type ServiceCheckResponse struct {
	ServiceCheckConfig `json:"request"`

	StatusCode int         `json:"status-code,omitempty"`
	Error      string      `json:"error,omitempty"`
	Body       interface{} `json:"response-body,omitempty"`
}

// CheckServices .
func CheckServices(ctx context.Context, checks []ServiceCheckConfig) []ServiceCheckResponse {
	wg := sync.WaitGroup{}
	resps := []ServiceCheckResponse{}
	respsMu := sync.Mutex{}

	for i := range checks {
		wg.Add(1)

		if len(checks[i].Method) == 0 {
			checks[i].Method = "GET"
		}

		go func(idx int) {
			defer wg.Done()
			resp := checkService(ctx, checks[idx])

			respsMu.Lock()
			resps = append(resps, resp)
			respsMu.Unlock()
		}(i)
	}

	wg.Wait()
	return resps
}

func checkService(ctx context.Context, check ServiceCheckConfig) ServiceCheckResponse {
	sresp := ServiceCheckResponse{
		ServiceCheckConfig: check,
	}

	body := bytes.NewReader(nil)
	if check.Body != nil {
		b, err := json.Marshal(check.Body)
		if err != nil {
			sresp.Error = fmt.Sprintf("unable to marshal body: %s", err)
			return sresp
		}

		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(check.Method, check.URL, body)
	if err != nil {
		sresp.Error = fmt.Sprintf("unable to build request: %s", err)
		return sresp
	}

	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		sresp.Error = fmt.Sprintf("unable to make reuqest: %s", err)
		return sresp
	}
	defer resp.Body.Close()

	sresp.StatusCode = resp.StatusCode

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		sresp.Error = fmt.Sprintf("unable to read response body: %s", err)
		return sresp
	}

	err = json.Unmarshal(b, &sresp.Body)
	if err != nil {
		sresp.Error = fmt.Sprintf("unable to parse response body: %s", err)
		return sresp
	}

	return sresp
}
