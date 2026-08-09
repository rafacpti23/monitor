package papi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"p-mon-backend/internal/alerts"
	"p-mon-backend/internal/db"
	"p-mon-backend/internal/ws"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

// CheckPanel fetches instances from the PAPI panel, upserts local state, and
// creates/resolves incidents as needed.
func CheckPanel(panelID int64) {
	var p models.PapiPanel
	var lastChecked sql.NullTime
	err := db.DB.QueryRow(`SELECT id, user_id, name, base_url, panel_token, check_interval_sec,
		status, last_checked, last_error, total_instances, connected_instances, channels, created_at
		FROM papi_panels WHERE id = ?`, panelID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.BaseURL, &p.PanelToken, &p.CheckIntervalSec,
		&p.Status, &lastChecked, &p.LastError, &p.TotalInstances, &p.ConnectedInstances,
		&p.Channels, &p.CreatedAt)
	if err != nil {
		log.Printf("[papi] panel %d not found: %v", panelID, err)
		return
	}
	if lastChecked.Valid {
		p.LastChecked = &lastChecked.Time
	}

	instances, fetchErr := fetchInstances(p.BaseURL, p.PanelToken)
	if fetchErr != nil {
		log.Printf("[papi] panel %d fetch error: %v", panelID, fetchErr)
		_, _ = db.DB.Exec(`UPDATE papi_panels SET status = 'error', last_checked = datetime('now'), last_error = ? WHERE id = ?`,
			fetchErr.Error(), panelID)
		broadcastPapiUpdate(p.UserID, panelID, "error")
		return
	}

	total := len(instances)
	connected := 0
	for _, inst := range instances {
		if isPapiConnected(inst.Status) {
			connected++
		}
	}

	_, _ = db.DB.Exec(`UPDATE papi_panels SET status = 'ok', last_checked = datetime('now'),
		last_error = '', total_instances = ?, connected_instances = ? WHERE id = ?`,
		total, connected, panelID)

	// Fetch user's enabled channel types once (used for all incident calls below).
	userChannels := alerts.GetUserChannelsJSON(p.UserID)

	// Upsert each instance and track status transitions.
	for _, inst := range instances {
		instStatus := strings.ToUpper(inst.Status)
		var localID int64
		var prevStatus string
		err := db.DB.QueryRow(`SELECT id, status FROM papi_instances WHERE panel_id = ? AND instance_id = ?`,
			panelID, inst.ID).Scan(&localID, &prevStatus)

		if err == sql.ErrNoRows {
			// New instance, insert.
			res, insErr := db.DB.Exec(`INSERT INTO papi_instances
				(panel_id, user_id, instance_id, name, phone_number, status, last_seen, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`,
				panelID, p.UserID, inst.ID, inst.Name, inst.PhoneConnected, instStatus)
			if insErr != nil {
				continue
			}
			localID, _ = res.LastInsertId()
			prevStatus = "" // new, treat as no previous
		} else if err == nil {
			// Existing instance, update.
			_, _ = db.DB.Exec(`UPDATE papi_instances SET name = ?, phone_number = ?, status = ?,
				last_seen = datetime('now'), updated_at = datetime('now') WHERE id = ?`,
				inst.Name, inst.PhoneConnected, instStatus, localID)
		} else {
			continue
		}

		// Status transitions → incidents.
		label := inst.Name
		if label == "" {
			label = inst.ID
		}
		if inst.PhoneConnected != "" {
			label += " (" + inst.PhoneConnected + ")"
		}

		if !isPapiConnected(instStatus) {
			// Instance is NOT connected → create critical incident (deduplicated by CreateIncident).
			msg := fmt.Sprintf("Instância PAPI '%s' está %s (painel %s)", label, instStatus, p.Name)
			alerts.CreateIncident(p.UserID, "papi_instance", localID, "papi_disconnected", "critical", msg, userChannels)
		} else if prevStatus != "" && !isPapiConnected(prevStatus) {
			// Was NOT connected, now IS connected → resolve + notify "voltou".
			subject := "[P-mon] Instância PAPI reconectada"
			body := fmt.Sprintf("Instância PAPI '%s' voltou a ficar conectada (painel %s)", label, p.Name)
			alerts.ResolveIncidentAndNotify(p.UserID, "papi_instance", localID, "papi_disconnected", userChannels, subject, body)
		} else if isPapiConnected(instStatus) && prevStatus == "" {
			// New instance, already connected — no action needed.
		}
	}

	// Mark instances no longer returned by the API as disappeared.
	seen := map[string]bool{}
	for _, inst := range instances {
		seen[inst.ID] = true
	}
	existingRows, err := db.DB.Query(`SELECT id, instance_id, name, phone_number, status FROM papi_instances WHERE panel_id = ?`, panelID)
	if err == nil {
		defer existingRows.Close()
		for existingRows.Next() {
			var eID int64
			var eInstID, eName, ePhone, eStatus string
			if err := existingRows.Scan(&eID, &eInstID, &eName, &ePhone, &eStatus); err != nil {
				continue
			}
			if !seen[eInstID] {
				// Instance no longer in API response — mark as removed.
				_, _ = db.DB.Exec(`UPDATE papi_instances SET status = 'REMOVED', updated_at = datetime('now') WHERE id = ?`, eID)
				if isPapiConnected(eStatus) {
					label := eName
					if label == "" {
						label = eInstID
					}
					if ePhone != "" {
						label += " (" + ePhone + ")"
					}
					msg := fmt.Sprintf("Instância PAPI '%s' foi removida do painel %s", label, p.Name)
					alerts.CreateIncident(p.UserID, "papi_instance", eID, "papi_disconnected", "critical", msg, userChannels)
				}
			}
		}
	}

	broadcastPapiUpdate(p.UserID, panelID, "ok")
}

// fetchInstances calls PAPI /api/v1/instances and returns parsed results.
func fetchInstances(baseURL, panelToken string) ([]models.PapiAPIInstance, error) {
	url := strings.TrimRight(baseURL, "/") + "/api/v1/instances"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-panel-token", panelToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PAPI returned %d: %s", resp.StatusCode, string(body))
	}

	var apiResp models.PapiAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode PAPI response: %v", err)
	}
	if !apiResp.Success {
		return nil, fmt.Errorf("PAPI API returned success=false")
	}
	return apiResp.Instances, nil
}

func broadcastPapiUpdate(userID, panelID int64, status string) {
	if ws.DefaultHub != nil {
		ws.DefaultHub.Broadcast(userID, "papi_update", gin.H{
			"panel_id": panelID,
			"status":   status,
		})
	}
}

// isPapiConnected returns true if the PAPI instance status indicates a connected session.
func isPapiConnected(status string) bool {
	s := strings.ToUpper(strings.TrimSpace(status))
	return s == "CONNECTED" || s == "ACTIVE" || s == "OPEN"
}
