package worker

import (
	"bytes"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"io"
	"net/http"
	"strconv"
)

type Client struct {
	AnalyticsUrl string
	GrafanaUrl   string
	Token        string
	Logger       log.Logger
}

func (api *Client) getResponseBody(res *http.Response) []byte {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		level.Info(api.Logger).Log(
			"status", "error",
			"message", "Failed to get response body",
			"error", err,
		)

		return []byte{}
	}

	return body
}

func (api *Client) prepare(method string, endpoint string, payload []byte) (*http.Request, error) {
	bearer := "Bearer " + api.Token

	req, err := http.NewRequest(method, api.GrafanaUrl+endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, errors.New("Failed to create new HTTP Request")
	}

	req.Header.Add("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func (api *Client) Get(endpoint string) ([]byte, error) {
	req, err := api.prepare("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Request failed with status: " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (api *Client) Post(endpoint string, payload []byte) ([]byte, error) {
	req, err := api.prepare("POST", endpoint, payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Request failed with status: " + strconv.Itoa(res.StatusCode))
	}
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
