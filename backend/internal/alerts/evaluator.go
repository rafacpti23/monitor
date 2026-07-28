package alerts

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"
)

// occurrenceKey uniquely identifies a rule's evaluation streak.
type occurrenceKey struct {
	RuleID int64
}

var (
	occMu       sync.Mutex
	occurrences = map[occurrenceKey]int{}
	lastFired   = map[occurrenceKey]time.Time{}
)

// EvaluateServerMetrics runs all enabled server alert rules against a fresh metric sample.
func EvaluateServerMetrics(userID, serverID int64, p *models.AgentPayload, diskTotal, diskUsed int64) {
	rows, err := db.DB.Query(`SELECT id, user_id, monitor_type, monitor_id, alert_type, comparison, threshold, occurrences, cooldown_min, status, channels
		FROM alert_rules WHERE user_id = ? AND monitor_type = 'server' AND monitor_id = ? AND status = 'enabled'`, userID, serverID)
	if err != nil {
		return
	}
	defer rows.Close()

	rules := []models.AlertRule{}
	for rows.Next() {
		var r models.AlertRule
		if err := rows.Scan(&r.ID, &r.UserID, &r.MonitorType, &r.MonitorID, &r.AlertType, &r.Comparison,
			&r.Threshold, &r.Occurrences, &r.CooldownMin, &r.Status, &r.Channels); err == nil {
			rules = append(rules, r)
		}
	}
	rows.Close()

	for _, r := range rules {
		matched := false
		msg := ""

		switch r.AlertType {
		case "cpu":
			matched = compareFloat(p.CPUPercent, r.Comparison, r.Threshold)
			msg = fmt.Sprintf("CPU at %.1f%% (rule: %s %s)", p.CPUPercent, r.Comparison, r.Threshold)
		case "ram":
			var usage float64
			if p.RAMTotalBytes > 0 {
				usage = float64(p.RAMUsedBytes) / float64(p.RAMTotalBytes) * 100
			}
			matched = compareFloat(usage, r.Comparison, r.Threshold)
			msg = fmt.Sprintf("RAM at %.1f%% (rule: %s %s)", usage, r.Comparison, r.Threshold)
		case "disk":
			var usage float64
			if diskTotal > 0 {
				usage = float64(diskUsed) / float64(diskTotal) * 100
			}
			matched = compareFloat(usage, r.Comparison, r.Threshold)
			msg = fmt.Sprintf("Disk at %.1f%% (rule: %s %s)", usage, r.Comparison, r.Threshold)
		case "load":
			matched = compareFloat(p.LoadAvg[0], r.Comparison, r.Threshold)
			msg = fmt.Sprintf("Load1 at %.2f (rule: %s %s)", p.LoadAvg[0], r.Comparison, r.Threshold)
		case "service_down":
			down := !serviceRunning(p.Services, r.Threshold) && !pm2Online(p.PM2Processes, r.Threshold)
			matched = down
			msg = fmt.Sprintf("Service '%s' is not running", r.Threshold)
		case "docker_down":
			matched = !dockerRunning(p.DockerContainers, r.Threshold)
			msg = fmt.Sprintf("Docker container '%s' is not running", r.Threshold)
		}

		evaluateRule(r, matched, "server", serverID, msg)
	}
}

// EvaluateCheckResult evaluates rules for a generic check result (latency etc.).
func EvaluateCheckResult(userID, checkID int64, ok bool, latencyMs int) {
	rows, err := db.DB.Query(`SELECT id, user_id, monitor_type, monitor_id, alert_type, comparison, threshold, occurrences, cooldown_min, status, channels
		FROM alert_rules WHERE user_id = ? AND monitor_type = 'check' AND monitor_id = ? AND status = 'enabled'`, userID, checkID)
	if err != nil {
		return
	}
	rules := []models.AlertRule{}
	for rows.Next() {
		var r models.AlertRule
		if err := rows.Scan(&r.ID, &r.UserID, &r.MonitorType, &r.MonitorID, &r.AlertType, &r.Comparison,
			&r.Threshold, &r.Occurrences, &r.CooldownMin, &r.Status, &r.Channels); err == nil {
			rules = append(rules, r)
		}
	}
	rows.Close()

	for _, r := range rules {
		matched := false
		msg := ""
		switch r.AlertType {
		case "ping_latency":
			matched = compareFloat(float64(latencyMs), r.Comparison, r.Threshold)
			msg = fmt.Sprintf("Latency %dms (rule: %s %s)", latencyMs, r.Comparison, r.Threshold)
		}
		evaluateRule(r, matched, "check", checkID, msg)
	}
}

// evaluateRule tracks consecutive occurrences and fires when the threshold streak is reached.
func evaluateRule(r models.AlertRule, matched bool, monitorType string, monitorID int64, msg string) {
	key := occurrenceKey{RuleID: r.ID}

	occMu.Lock()
	if matched {
		occurrences[key]++
	} else {
		occurrences[key] = 0
	}
	count := occurrences[key]
	last := lastFired[key]
	occMu.Unlock()

	if !matched || count < r.Occurrences {
		return
	}

	// Respect cooldown.
	if !last.IsZero() && time.Since(last) < time.Duration(r.CooldownMin)*time.Minute {
		return
	}

	occMu.Lock()
	lastFired[key] = time.Now()
	occurrences[key] = 0
	occMu.Unlock()

	severity := "warning"
	if r.AlertType == "nodata" || r.AlertType == "website_down" || r.AlertType == "service_down" || r.AlertType == "docker_down" {
		severity = "critical"
	}

	CreateIncident(r.UserID, monitorType, monitorID, r.AlertType, severity, msg, r.Channels)
}

// CreateIncident creates an incident (if none is active) and dispatches notifications.
func CreateIncident(userID int64, monitorType string, monitorID int64, alertType, severity, msg, channelsJSON string) {
	var existingID int64
	err := db.DB.QueryRow(`SELECT id FROM incidents WHERE user_id = ? AND monitor_type = ? AND monitor_id = ? AND alert_type = ? AND status IN ('active','acknowledged')`,
		userID, monitorType, monitorID, alertType).Scan(&existingID)
	if err == nil && existingID > 0 {
		return
	}

	res, err := db.DB.Exec(`INSERT INTO incidents (user_id, monitor_type, monitor_id, alert_type, severity, status, message)
		VALUES (?, ?, ?, ?, ?, 'active', ?)`, userID, monitorType, monitorID, alertType, severity, msg)
	if err != nil {
		log.Printf("Failed to create incident: %v", err)
		return
	}
	incidentID, _ := res.LastInsertId()
	log.Printf("Incident #%d created (user %d): %s", incidentID, userID, msg)

	DispatchNotifications(userID, incidentID, channelsJSON, fmt.Sprintf("[P-mon] %s alert", alertType), msg)
}

// ResolveIncidentAndNotify closes the active incident for the given monitor+alert
// and, if one existed, sends a "back to normal" notification through channelsJSON.
// Returns true when an active incident was found and resolved.
func ResolveIncidentAndNotify(userID int64, monitorType string, monitorID int64, alertType, channelsJSON, subject, body string) bool {
	var incidentID int64
	err := db.DB.QueryRow(`SELECT id FROM incidents WHERE user_id = ? AND monitor_type = ? AND monitor_id = ? AND alert_type = ? AND status IN ('active','acknowledged')`,
		userID, monitorType, monitorID, alertType).Scan(&incidentID)
	if err != nil || incidentID == 0 {
		return false
	}
	_, _ = db.DB.Exec(`UPDATE incidents SET status = 'resolved', resolved_at = datetime('now'), end_time = datetime('now') WHERE id = ?`, incidentID)
	log.Printf("Incident #%d resolved (user %d): %s", incidentID, userID, body)
	if channelsJSON != "" && channelsJSON != "[]" {
		DispatchNotifications(userID, incidentID, channelsJSON, subject, body)
	}
	return true
}

// DispatchNotifications resolves channel references and sends messages.
// channelsJSON may be a JSON array of channel IDs (numbers) or channel type names.
func DispatchNotifications(userID, incidentID int64, channelsJSON, subject, body string) {
	if channelsJSON == "" {
		channelsJSON = "[]"
	}
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(channelsJSON), &raw); err != nil {
		return
	}

	for _, item := range raw {
		var chID int64
		var chType string
		if err := json.Unmarshal(item, &chID); err == nil {
			sendToChannelByID(userID, incidentID, chID, subject, body)
			continue
		}
		if err := json.Unmarshal(item, &chType); err == nil {
			sendToChannelsByType(userID, incidentID, chType, subject, body)
		}
	}
}

func sendToChannelByID(userID, incidentID, channelID int64, subject, body string) {
	var chType, config string
	var enabled bool
	err := db.DB.QueryRow(`SELECT type, config, enabled FROM alert_channels WHERE id = ? AND user_id = ?`, channelID, userID).
		Scan(&chType, &config, &enabled)
	if err != nil || !enabled {
		return
	}
	sendVia(userID, incidentID, channelID, chType, config, subject, body)
}

func sendToChannelsByType(userID, incidentID int64, chType, subject, body string) {
	rows, err := db.DB.Query(`SELECT id, config FROM alert_channels WHERE user_id = ? AND type = ? AND enabled = 1`, userID, chType)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var chID int64
		var config string
		if err := rows.Scan(&chID, &config); err == nil {
			sendVia(userID, incidentID, chID, chType, config, subject, body)
		}
	}
}

// sendVia performs the actual send and records the result in notifications_sent.
func sendVia(userID, incidentID, channelID int64, chType, configJSON, subject, body string) {
	var cfg map[string]string
	_ = json.Unmarshal([]byte(configJSON), &cfg)
	if cfg == nil {
		cfg = map[string]string{}
	}

	var err error
	switch chType {
	case "email":
		err = SendEmail(cfg["to"], subject, body)
	case "whatsapp":
		// PAPI provider: config has instance, api_key, jid.
		err = SendWhatsAppPAPI(cfg["instance"], cfg["api_key"], cfg["jid"], subject+"\n"+body)
	case "webhook":
		payload, _ := json.Marshal(map[string]interface{}{
			"subject": subject,
			"body":    body,
		})
		err = SendWebhook(cfg["url"], string(payload))
	default:
		err = fmt.Errorf("unknown channel type %q", chType)
	}

	success := 0
	if err == nil {
		success = 1
	} else {
		log.Printf("Notification via channel %d (%s) failed: %v", channelID, chType, err)
	}
	if incidentID > 0 {
		_, _ = db.DB.Exec(`INSERT INTO notifications_sent (incident_id, channel_id, success) VALUES (?, ?, ?)`, incidentID, channelID, success)
	}
}

// compareFloat evaluates `val <comparison> threshold`.
func compareFloat(val float64, comparison, thresholdStr string) bool {
	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		return false
	}
	switch comparison {
	case ">=":
		return val >= threshold
	case "<=":
		return val <= threshold
	case ">":
		return val > threshold
	case "<":
		return val < threshold
	case "==":
		return val == threshold
	case "!=":
		return val != threshold
	default:
		return false
	}
}

func serviceRunning(services []models.AgentService, name string) bool {
	for _, s := range services {
		if s.Name == name && s.Running {
			return true
		}
	}
	return false
}

func pm2Online(procs []models.AgentPM2Process, name string) bool {
	for _, p := range procs {
		if p.Name == name && p.Status == "online" {
			return true
		}
	}
	return false
}

func dockerRunning(containers []models.AgentContainer, name string) bool {
	for _, c := range containers {
		if c.Name == name && c.Status == "running" {
			return true
		}
	}
	return false
}
