package worker

import (
	"bytes"
	"encoding/json"
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
	Filter       string
}

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
