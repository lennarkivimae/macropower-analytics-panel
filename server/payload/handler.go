package payload

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MacroPower/macropower-analytics-panel/server/cacher"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Handler is the handler for incoming payloads.
type Handler struct {
	logger log.Logger
	ch     chan Payload
}

// NewHandler creates a new Handler.
func NewHandler(cache *cacher.Cacher, buffer int, sessionLog bool, variableLog bool, raw bool, logger log.Logger) *Handler {
	ch := make(chan Payload, buffer)
	go startProcessor(cache, ch, sessionLog, variableLog, raw, logger)

	return &Handler{
		logger: logger,
		ch:     ch,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := Payload{}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	h.ch <- p

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "")
}

// startProcessor starts a receiver and optional logger for the Payload channel.
func startProcessor(cache *cacher.Cacher, c <-chan Payload, sessionLog bool, variableLog bool, raw bool, logger log.Logger) {
	for p := range c {
		if p.Dashboard.UID != "new" {
			ProcessPayload(cache, p, logger)
		}
		if sessionLog {
			LogPayload(p, variableLog, logger, raw)
		}
	}
}

// processPayload is a receiver for Payloads.
func ProcessPayload(cache *cacher.Cacher, p Payload, logger log.Logger) {
	switch p.Type {
	case "start":
		addStart(cache, p)
	case "heartbeat":
		addHeartbeat(cache, p)
	case "end":
		addEnd(cache, p)
	default:
		addHeartbeat(cache, p)
		_ = level.Warn(logger).Log(
			"msg", "Session has invalid type, defaulted to heartbeat",
			"uuid", p.UUID,
			"type", p.Type,
		)
	}
}

// LogPayload writes a log describing the Payload.
func LogPayload(p Payload, logVars bool, logger log.Logger, raw bool) {
	if !logVars {
		p.Variables = p.Variables[:0]
	}

	if raw {
		level.Info(logger).Log("msg", "Received session data", "data", p)
		return
	}

	h := p.Host
	bi := h.BuildInfo
	li := h.LicenseInfo
	u := p.User
	tr := p.TimeRange

	var theme string
	if u.LightTheme {
		theme = "light"
	} else {
		theme = "dark"
	}

	var role string
	if u.IsGrafanaAdmin {
		role = "admin"
	} else if u.HasEditPermissionInFolders {
		role = "editor"
	} else {
		role = "user"
	}

	labels := []interface{}{
		"msg", "Received session data",
		"uuid", p.UUID,
		"type", p.Type,
		"has_focus", p.HasFocus,
		"host", fmt.Sprintf("%s//%s:%s", h.Protocol, h.Hostname, h.Port),
		"build", fmt.Sprintf("(commit=%s, edition=%s, env=%s, version=%s)", bi.Commit, bi.Edition, bi.Env, bi.Version),
		"license", fmt.Sprintf("(state=%s, expiry=%d, license=%t)", li.StateInfo, li.Expiry, li.HasLicense),
		"dashboard_name", p.Dashboard.Name,
		"dashboard_uid", p.Dashboard.UID,
		"dashboard_timezone", p.TimeZone,
		"user_id", u.ID,
		"user_login", u.Login,
		"user_email", u.Email,
		"user_name", u.Name,
		"user_theme", theme,
		"user_role", role,
		"user_locale", u.Locale,
		"user_timezone", u.Timezone,
		"time_from", tr.From,
		"time_to", tr.To,
		"time_from_raw", tr.Raw.From,
		"time_to_raw", tr.Raw.To,
		"timeorigin", p.TimeOrigin,
		"time", p.Time,
	}

	for _, v := range p.Variables {
		var variableValues []string
		for _, value := range v.Values {
			variableValues = append(variableValues, value.(string))
		}
		d := fmt.Sprintf("(label=%s, type=%s, multi=%t, count=%d, values=[%s])", v.Label, v.Type, v.Multi, len(v.Values), strings.Join(variableValues, ","))
		labels = append(labels, v.Name, d)
	}

	_ = level.Info(logger).Log(labels...)
}
