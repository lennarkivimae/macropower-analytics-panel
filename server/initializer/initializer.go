package initializer

import (
	"github.com/MacroPower/macropower-analytics-panel/server/cacher"
	"github.com/MacroPower/macropower-analytics-panel/server/payload"
	"github.com/MacroPower/macropower-analytics-panel/server/worker"
	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"time"
)

func InitializeMetricsForDashboards(api worker.Client, logger log.Logger, cache *cacher.Cacher) {
	dashboards, hasErrored := api.GetDashboards()
	if hasErrored {
		return
	}

	for _, dashboard := range dashboards {
		payloadData := createDashboardPayload(dashboard.Uid, dashboard.Title, api.GrafanaUrl)

		payload.ProcessPayload(cache, payloadData, logger)
	}
}

func createDashboardPayload(dashboardUid string, dashboardName string, grafanaUrl string) payload.Payload {
	id := uuid.New()
	currentTime := int(time.Now().Unix())

	options := payload.OptionsInfo{
		PostStart:         false,
		PostEnd:           false,
		PostHeartbeat:     false,
		HeartbeatInterval: 60,
		HeartbeatAlways:   false,
	}

	host := payload.HostInfo{
		Hostname:  grafanaUrl,
		Port:      "",
		Protocol:  "https",
		BuildInfo: payload.HostBuildInfo{},
		LicenseInfo: payload.HostLicenseInfo{
			HasLicense: false,
			Expiry:     0,
			StateInfo:  "",
		},
	}

	dashboard := payload.DashboardInfo{
		Name: dashboardName,
		UID:  dashboardUid,
	}

	user := payload.UserInfo{
		IsSignedIn:                 false,
		ID:                         0,
		Login:                      payload.ANALYTICS_USER,
		Email:                      "",
		Name:                       payload.ANALYTICS_USER,
		LightTheme:                 false,
		OrgCount:                   0,
		OrgID:                      0,
		OrgName:                    "",
		OrgRole:                    "",
		IsGrafanaAdmin:             false,
		Timezone:                   "",
		Locale:                     "",
		HasEditPermissionInFolders: false,
	}

	timeRange := payload.TimeRangeInfo{
		From: currentTime,
		To:   currentTime,
		Raw:  payload.TimeRangeRawInfo{},
	}

	return payload.Payload{
		UUID:       id.String(),
		Type:       "heartbeat",
		HasFocus:   false,
		Options:    options,
		Host:       host,
		Dashboard:  dashboard,
		User:       user,
		Variables:  []payload.VariablesInfo{},
		TimeRange:  timeRange,
		TimeZone:   "server",
		TimeOrigin: currentTime,
		Time:       currentTime,
	}
}
