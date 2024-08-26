package worker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	Filter       string
}

const contentTypeJson = "application/json"

func (api *Client) GetDashboards() ([]DashboardsResponse, bool) {
	res, err := api.Get("/api/search")
	if err != nil {
		level.Info(api.Logger).Log(
			"status", "error",
			"message", "GetDashboards - Failed to get dashboards data",
			"error", err,
		)

		return nil, true
	}

	var response []DashboardsResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		level.Info(api.Logger).Log(
			"status", "error",
			"message", "GetDashboards - Failed to parse JSON response",
			"error", err,
		)

		return nil, true
	}
	return response, false
}

func (api *Client) GetDashboard(uid string) *Dashboard {
	res, err := api.Get("/api/dashboards/uid/" + uid)
	if err != nil {
		level.Info(api.Logger).Log(
			"status", "error",
			"message", "GetDashboard - api.Get failed",
			"error", err,
		)

		return nil
	}

	var dashboardData map[string]interface{}
	err = json.Unmarshal(res, &dashboardData)
	if err != nil {
		level.Info(api.Logger).Log(
			"status", "error",
			"message", "GetDashboard - Failed to parse dashboardData response",
			"error", err,
		)

		return nil
	}

	return &Dashboard{
		Uid:  uid,
		Data: dashboardData,
	}
}

func (api *Client) Get(endpoint string) ([]byte, error) {
	req, err := api.prepareRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if res.StatusCode != http.StatusOK {
		errorMessage := fmt.Sprintf("Request failed\nEndpoint: %s\nStatus: %s", endpoint, strconv.Itoa(res.StatusCode))

		return nil, errors.New(errorMessage)
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
	req, err := api.prepareRequest("POST", endpoint)
	if err != nil {
		return nil, err
	}

	setContentType(req, contentTypeJson)
	addBodyToRequest(req, payload)

	client := &http.Client{}
	res, err := client.Do(req)
	if res.StatusCode != http.StatusOK {
		errorMessage := fmt.Sprintf("Request failed\nEndpoint: %s\nStatus: %s", endpoint, strconv.Itoa(res.StatusCode))

		return nil, errors.New(errorMessage)
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

func (api *Client) prepareRequest(method string, endpoint string) (*http.Request, error) {
	req, err := http.NewRequest(method, api.GrafanaUrl+endpoint, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to create new HTTP Request\nMethod: %s\nEndpoint: %s", method, endpoint))
	}

	req.Header.Add("Authorization", "Bearer "+api.Token)
	req.Header.Add("Accept", contentTypeJson)

	return req, nil
}

func setContentType(req *http.Request, contentType string) {
	req.Header.Add("Content-Type", contentType)
}

func addBodyToRequest(req *http.Request, payload []byte) {
	req.Body = io.NopCloser(bytes.NewBuffer(payload))
	req.ContentLength = int64(len(payload))
}
