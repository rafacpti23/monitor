package api

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"p-mon-backend/internal/db"
	"p-mon-backend/internal/papi"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

// papiPanelCols keeps SELECT column order in sync with scanPapiPanel.
const papiPanelCols = `id, user_id, name, base_url, panel_token, check_interval_sec,
	status, last_checked, last_error, total_instances, connected_instances, channels, created_at`

func scanPapiPanel(row interface{ Scan(...interface{}) error }, p *models.PapiPanel) error {
	var lastChecked sql.NullTime
	err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.BaseURL, &p.PanelToken, &p.CheckIntervalSec,
		&p.Status, &lastChecked, &p.LastError, &p.TotalInstances, &p.ConnectedInstances,
		&p.Channels, &p.CreatedAt)
	if err != nil {
		return err
	}
	if lastChecked.Valid {
		p.LastChecked = &lastChecked.Time
	}
	return nil
}

// ListPapiPanels — GET /api/v1/papi/panels
func ListPapiPanels(c *gin.Context) {
	userID := c.GetInt64("user_id")
	rows, err := db.DB.Query(`SELECT `+papiPanelCols+` FROM papi_panels WHERE user_id = ? ORDER BY id DESC`, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	panels := []models.PapiPanel{}
	for rows.Next() {
		var p models.PapiPanel
		if err := scanPapiPanel(rows, &p); err != nil {
			continue
		}
		// Hide token in list responses.
		p.PanelToken = ""
		panels = append(panels, p)
	}
	c.JSON(200, panels)
}

// GetPapiPanel — GET /api/v1/papi/panels/:id (includes instances)
func GetPapiPanel(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var p models.PapiPanel
	err := scanPapiPanel(
		db.DB.QueryRow(`SELECT `+papiPanelCols+` FROM papi_panels WHERE id = ? AND user_id = ?`, id, userID),
		&p,
	)
	if err != nil {
		c.JSON(404, gin.H{"error": "panel not found"})
		return
	}
	p.PanelToken = ""

	// Attach instances.
	rows, err := db.DB.Query(`SELECT id, panel_id, user_id, instance_id, name, phone_number, status, last_seen, updated_at
		FROM papi_instances WHERE panel_id = ? ORDER BY status = 'CONNECTED' DESC, name ASC`, id)
	if err == nil {
		defer rows.Close()
		insts := []models.PapiInstance{}
		for rows.Next() {
			var ins models.PapiInstance
			var lastSeen sql.NullTime
			if err := rows.Scan(&ins.ID, &ins.PanelID, &ins.UserID, &ins.InstanceID, &ins.Name,
				&ins.PhoneNumber, &ins.Status, &lastSeen, &ins.UpdatedAt); err == nil {
				if lastSeen.Valid {
					ins.LastSeen = &lastSeen.Time
				}
				insts = append(insts, ins)
			}
		}
		p.Instances = insts
	}
	c.JSON(200, p)
}

// papiPanelBody is the JSON body for create/update.
type papiPanelBody struct {
	Name             string   `json:"name"`
	BaseURL          string   `json:"base_url"`
	PanelToken       string   `json:"panel_token"`
	CheckIntervalSec int      `json:"check_interval_sec"`
	Channels         []string `json:"channels"`
}

func (b *papiPanelBody) normalize() {
	b.Name = strings.TrimSpace(b.Name)
	b.BaseURL = strings.TrimRight(strings.TrimSpace(b.BaseURL), "/")
	if b.BaseURL == "" {
		b.BaseURL = "https://papi.api.br"
	}
	if b.CheckIntervalSec < 30 {
		b.CheckIntervalSec = 60
	}
}

// CreatePapiPanel — POST /api/v1/papi/panels
func CreatePapiPanel(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var body papiPanelBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	body.normalize()
	if body.Name == "" || body.PanelToken == "" {
		c.JSON(400, gin.H{"error": "name and panel_token are required"})
		return
	}

	channelsJSON := "[]"
	if len(body.Channels) > 0 {
		b, _ := json.Marshal(body.Channels)
		channelsJSON = string(b)
	}

	res, err := db.DB.Exec(`INSERT INTO papi_panels
		(user_id, name, base_url, panel_token, check_interval_sec, status, channels)
		VALUES (?, ?, ?, ?, ?, 'pending', ?)`,
		userID, body.Name, body.BaseURL, body.PanelToken, body.CheckIntervalSec, channelsJSON)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()

	// Kick off first check async (best-effort; scheduler will pick it up too).
	go papi.CheckPanel(id)

	c.JSON(201, gin.H{"id": id})
}

// UpdatePapiPanel — PUT /api/v1/papi/panels/:id
func UpdatePapiPanel(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body papiPanelBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	body.normalize()
	if body.Name == "" {
		c.JSON(400, gin.H{"error": "name is required"})
		return
	}

	channelsJSON := "[]"
	if len(body.Channels) > 0 {
		b, _ := json.Marshal(body.Channels)
		channelsJSON = string(b)
	}

	// Update token only when a non-empty value comes in the body (keeps existing when blank).
	if strings.TrimSpace(body.PanelToken) != "" {
		_, err := db.DB.Exec(`UPDATE papi_panels
			SET name = ?, base_url = ?, panel_token = ?, check_interval_sec = ?, channels = ?
			WHERE id = ? AND user_id = ?`,
			body.Name, body.BaseURL, body.PanelToken, body.CheckIntervalSec, channelsJSON, id, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	} else {
		_, err := db.DB.Exec(`UPDATE papi_panels
			SET name = ?, base_url = ?, check_interval_sec = ?, channels = ?
			WHERE id = ? AND user_id = ?`,
			body.Name, body.BaseURL, body.CheckIntervalSec, channelsJSON, id, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(200, gin.H{"ok": true})
}

// DeletePapiPanel — DELETE /api/v1/papi/panels/:id
func DeletePapiPanel(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, err := db.DB.Exec(`DELETE FROM papi_panels WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// CheckPapiPanelNow — POST /api/v1/papi/panels/:id/check (trigger check on-demand)
func CheckPapiPanelNow(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// Ownership check.
	var owner int64
	if err := db.DB.QueryRow(`SELECT user_id FROM papi_panels WHERE id = ?`, id).Scan(&owner); err != nil || owner != userID {
		c.JSON(404, gin.H{"error": "panel not found"})
		return
	}
	go papi.CheckPanel(id)
	c.JSON(200, gin.H{"ok": true, "message": "check triggered"})
}

// GetPapiPanelIncidents — GET /api/v1/papi/panels/:id/incidents (last 15)
func GetPapiPanelIncidents(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	rows, err := db.DB.Query(`SELECT id, user_id, monitor_type, monitor_id, alert_type, severity, status,
		start_time, end_time, acknowledged_by, acknowledged_at, resolved_at, ignored, comment, message
		FROM incidents WHERE user_id = ? AND monitor_type IN ('papi_panel','papi_instance') AND monitor_id = ?
		ORDER BY start_time DESC LIMIT 15`, userID, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []models.Incident{}
	for rows.Next() {
		i, err := scanIncident(rows)
		if err != nil {
			continue
		}
		list = append(list, i)
	}
	// Also include instance-level incidents for this panel by matching monitor_type='papi_instance'
	// AND papi_instances.panel_id = id.
	rowsIns, err := db.DB.Query(`SELECT i.id, i.user_id, i.monitor_type, i.monitor_id, i.alert_type, i.severity, i.status,
		i.start_time, i.end_time, i.acknowledged_by, i.acknowledged_at, i.resolved_at, i.ignored, i.comment, i.message
		FROM incidents i
		JOIN papi_instances p ON p.id = i.monitor_id
		WHERE i.user_id = ? AND i.monitor_type = 'papi_instance' AND p.panel_id = ?
		ORDER BY i.start_time DESC LIMIT 15`, userID, id)
	if err == nil {
		defer rowsIns.Close()
		for rowsIns.Next() {
			i, err := scanIncident(rowsIns)
			if err != nil {
				continue
			}
			list = append(list, i)
		}
	}
	c.JSON(200, list)
}
