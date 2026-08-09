package stevo

import (
	"bytes"
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

// CheckStevoPanel fetches instances from the Stevo MCP endpoint, upserts local
// state, and creates/resolves incidents as needed.
func CheckStevoPanel(panelID int64) {
	var p models.PapiPanel
	var lastChecked sql.NullTime
	err := db.DB.QueryRow(`SELECT id, user_id, name, provider, base_url, panel_token, check_interval_sec,
		status, last_checked, last_error, total_instances, connected_instances, channels, created_at
		FROM papi_panels WHERE id = ?`, panelID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Provider, &p.BaseURL, &p.PanelToken, &p.CheckIntervalSec,
		&p.Status, &lastChecked, &p.LastError, &p.TotalInstances, &p.ConnectedInstances,
		&p.Channels, &p.CreatedAt)
	if err != nil {
		log.Printf("[stevo] panel %d not found: %v", panelID, err)
		return
	}
	if lastChecked.Valid {
		p.LastChecked = &lastChecked.Time
	}

	instances, fetchErr := fetchStevoInstances(p.BaseURL, p.PanelToken)
	if fetchErr != nil {
		log.Printf("[stevo] panel %d fetch error: %v", panelID, fetchErr)
		_, _ = db.DB.Exec(`UPDATE papi_panels SET status = 'error', last_checked = datetime('now'), last_error = ? WHERE id = ?`,
			fetchErr.Error(), panelID)
		broadcastStevoUpdate(p.UserID, panelID, "error")
		return
	}

	total := len(instances)
	connected := 0
	for _, inst := range instances {
		if inst.Connected {
			connected++
		}
	}

	_, _ = db.DB.Exec(`UPDATE papi_panels SET status = 'ok', last_checked = datetime('now'), last_error = '',
		total_instances = ?, connected_instances = ? WHERE id = ?`,
		total, connected, panelID)

	userChannels := alerts.GetUserChannelsJSON(p.UserID)

	for _, inst := range instances {
		displayName := inst.Name
		if displayName == "" {
			displayName = inst.InstanceName
		}
		if displayName == "" {
			displayName = inst.ID
		}

		// Read previous status BEFORE upsert so we can detect transitions.
		var prevStatus string
		var prevConnected int
		_ = db.DB.QueryRow(`SELECT status, CASE WHEN status IN ('active','connected','open') THEN 1 ELSE 0 END
			FROM papi_instances WHERE panel_id = ? AND instance_id = ?`,
			panelID, inst.ID).Scan(&prevStatus, &prevConnected)

		localID := upsertInstance(p.ID, p.UserID, inst)

		if !inst.Connected {
			msg := fmt.Sprintf("Instância Stevo '%s' está %s (painel %s)", displayName, inst.Status, p.Name)
			alerts.CreateIncident(p.UserID, "papi_instance", localID, "papi_disconnected", "critical", msg, userChannels)
		} else if prevConnected == 0 && prevStatus != "" {
			// Was NOT connected before, now IS → resolve + notify.
			subject := "[P-mon] Instância Stevo reconectada"
			body := fmt.Sprintf("Instância Stevo '%s' voltou a ficar conectada (painel %s)", displayName, p.Name)
			alerts.ResolveIncidentAndNotify(p.UserID, "papi_instance", localID, "papi_disconnected", userChannels, subject, body)
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
				_, _ = db.DB.Exec(`UPDATE papi_instances SET status = 'REMOVED', updated_at = datetime('now') WHERE id = ?`, eID)
				if eStatus == "active" || eStatus == "connected" || eStatus == "open" {
					label := eName
					if label == "" {
						label = eInstID
					}
					if ePhone != "" {
						label += " (" + ePhone + ")"
					}
					userChannels := alerts.GetUserChannelsJSON(p.UserID)
					msg := fmt.Sprintf("Instância Stevo '%s' foi removida do painel %s", label, p.Name)
					alerts.CreateIncident(p.UserID, "papi_instance", eID, "papi_disconnected", "critical", msg, userChannels)
				}
			}
		}
	}

	broadcastStevoUpdate(p.UserID, panelID, "ok")
}

// upsertInstance inserts or updates a Stevo instance and returns its local ID.
func upsertInstance(panelID, userID int64, inst models.StevoInstance) int64 {
	phone := inst.PhoneNumber
	status := inst.Status
	displayName := inst.Name
	if displayName == "" {
		displayName = inst.InstanceName
	}

	var existingID int64
	err := db.DB.QueryRow(`SELECT id FROM papi_instances WHERE panel_id = ? AND instance_id = ?`,
		panelID, inst.ID).Scan(&existingID)

	if err != nil {
		res, _ := db.DB.Exec(`INSERT INTO papi_instances (panel_id, user_id, instance_id, name, phone_number, status, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
			panelID, userID, inst.ID, displayName, phone, status)
		if res != nil {
			id, _ := res.LastInsertId()
			return id
		}
		return 0
	}

	_, _ = db.DB.Exec(`UPDATE papi_instances SET name = ?, phone_number = ?, status = ?, updated_at = datetime('now') WHERE id = ?`,
		displayName, phone, status, existingID)
	return existingID
}

// fetchStevoInstances calls the Stevo MCP endpoint using tools/call list_instances.
func fetchStevoInstances(baseURL, apiToken string) ([]models.StevoInstance, error) {
	url := strings.TrimRight(baseURL, "/") + "/mcp"

	// MCP protocol: call the "list_instances" tool via tools/call
	reqBody := models.StevoMCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "list_instances",
			"arguments": map[string]interface{}{},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Stevo returned %d: %s", resp.StatusCode, string(body))
	}

	// Stevo MCP returns SSE (text/event-stream). Read the "data:" line.
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read Stevo response: %v", readErr)
	}

	var jsonData string
	lines := strings.Split(string(respBody), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			jsonData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
	if jsonData == "" {
		jsonData = strings.TrimSpace(string(respBody))
	}
	if jsonData == "" {
		return nil, fmt.Errorf("Stevo MCP response contained no data")
	}

	// Parse the JSON-RPC response
	var mcpResp models.StevoMCPResponse
	if err := json.Unmarshal([]byte(jsonData), &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to decode Stevo MCP response: %v (raw: %.200s)", err, jsonData)
	}
	if mcpResp.Error != nil {
		return nil, fmt.Errorf("Stevo MCP error %d: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	// The result contains content[].text which is a JSON string with instances.
	var mcpResult models.StevoMCPResult
	if err := json.Unmarshal(mcpResp.Result, &mcpResult); err != nil {
		return nil, fmt.Errorf("failed to parse Stevo MCP result: %v", err)
	}
	if len(mcpResult.Content) == 0 {
		return nil, fmt.Errorf("Stevo MCP result has no content")
	}

	// The text field contains a JSON string with the instances array.
	var instancesResp models.StevoInstancesResponse
	if err := json.Unmarshal([]byte(mcpResult.Content[0].Text), &instancesResp); err != nil {
		return nil, fmt.Errorf("failed to parse Stevo instances: %v (text: %.200s)", err, mcpResult.Content[0].Text)
	}
	return instancesResp.Instances, nil
}

func broadcastStevoUpdate(userID, panelID int64, status string) {
	if ws.DefaultHub != nil {
		ws.DefaultHub.Broadcast(userID, "papi_update", gin.H{
			"panel_id": panelID,
			"status":   status,
		})
	}
}

// isStevoConnected returns true if the Stevo instance status indicates an open/connected session.
func isStevoConnected(status string) bool {
	s := strings.ToUpper(status)
	return s == "OPEN" || s == "CONNECTED" || s == "ACTIVE"
}
