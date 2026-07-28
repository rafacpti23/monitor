package api

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"p-mon-backend/internal/alerts"
	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/internal/ws"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func generateServerKey() string {
	b := make([]byte, 24) // 24 bytes -> 48 hex chars
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// scanLatestMetrics loads the most recent metrics row for a server (or nil).
func scanLatestMetrics(serverID int64) *models.ServerMetrics {
	var m models.ServerMetrics
	err := db.DB.QueryRow(`SELECT id, server_id, timestamp, cpu_percent, ram_total, ram_used,
		disk_total, disk_used, load1, load5, load15, net_rx, net_tx, uptime_seconds
		FROM server_metrics WHERE server_id = ? ORDER BY timestamp DESC LIMIT 1`, serverID).
		Scan(&m.ID, &m.ServerID, &m.Timestamp, &m.CPUPercent, &m.RAMTotal, &m.RAMUsed,
			&m.DiskTotal, &m.DiskUsed, &m.Load1, &m.Load5, &m.Load15, &m.NetRX, &m.NetTX, &m.UptimeSeconds)
	if err != nil {
		return nil
	}
	return &m
}

func scanServer(rows interface {
	Scan(dest ...interface{}) error
}) (models.Server, error) {
	var s models.Server
	err := rows.Scan(&s.ID, &s.UserID, &s.GroupID, &s.Name, &s.ServerKey, &s.Type, &s.OS,
		&s.Status, &s.OnMap, &s.Lat, &s.Lng, &s.Hostname, &s.AgentVersion, &s.LastSeen, &s.CreatedAt)
	return s, err
}

const serverCols = `id, user_id, group_id, name, server_key, type, os, status, on_map, lat, lng, hostname, agent_version, last_seen, created_at`

func ListServers(c *gin.Context) {
	userID := auth.GetUserID(c)
	rows, err := db.DB.Query(`SELECT `+serverCols+` FROM servers WHERE user_id = ? ORDER BY name`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	servers := []models.Server{}
	for rows.Next() {
		s, err := scanServer(rows)
		if err != nil {
			continue
		}
		s.LatestMetrics = scanLatestMetrics(s.ID)
		servers = append(servers, s)
	}
	c.JSON(http.StatusOK, servers)
}

type ServerReq struct {
	Name    string  `json:"name" binding:"required"`
	GroupID *int64  `json:"group_id"`
	Type    string  `json:"type"`
	OnMap   bool    `json:"on_map"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

func CreateServer(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req ServerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Type == "" {
		req.Type = "linux"
	}
	key := generateServerKey()
	res, err := db.DB.Exec(`INSERT INTO servers (user_id, group_id, name, server_key, type, on_map, lat, lng, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')`,
		userID, req.GroupID, req.Name, key, req.Type, req.OnMap, req.Lat, req.Lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}
	id, _ := res.LastInsertId()
	db.LogSystemAction(userID, "server_created", "Created server "+req.Name, c.ClientIP())

	var s models.Server
	err = scanServerByID(id, userID, &s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Created but failed to load"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func scanServerByID(id, userID int64, s *models.Server) error {
	row := db.DB.QueryRow(`SELECT `+serverCols+` FROM servers WHERE id = ? AND user_id = ?`, id, userID)
	res, err := scanServer(row)
	if err != nil {
		return err
	}
	*s = res
	return nil
}

func GetServer(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var s models.Server
	if err := scanServerByID(id, userID, &s); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}
	s.LatestMetrics = scanLatestMetrics(s.ID)

	// Load services
	rows, err := db.DB.Query("SELECT id, server_id, name, status, updated_at FROM server_services WHERE server_id = ? ORDER BY name", s.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var svc models.ServerService
			if err := rows.Scan(&svc.ID, &svc.ServerID, &svc.Name, &svc.Status, &svc.UpdatedAt); err == nil {
				s.Services = append(s.Services, svc)
			}
		}
	}
	c.JSON(http.StatusOK, s)
}

func UpdateServer(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req ServerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := db.DB.Exec(`UPDATE servers SET name = ?, group_id = ?, type = ?, on_map = ?, lat = ?, lng = ?
		WHERE id = ? AND user_id = ?`,
		req.Name, req.GroupID, req.Type, req.OnMap, req.Lat, req.Lng, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}
	db.LogSystemAction(userID, "server_updated", "Updated server "+strconv.FormatInt(id, 10), c.ClientIP())

	var s models.Server
	_ = scanServerByID(id, userID, &s)
	c.JSON(http.StatusOK, s)
}

func DeleteServer(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// Cascade delete incidents tied to this server (metrics/services cascade via FK).
	_, _ = db.DB.Exec("DELETE FROM incidents WHERE user_id = ? AND monitor_type = 'server' AND monitor_id = ?", userID, id)

	res, err := db.DB.Exec("DELETE FROM servers WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}
	db.LogSystemAction(userID, "server_deleted", "Deleted server "+strconv.FormatInt(id, 10), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func rangeToDuration(r string) time.Duration {
	switch r {
	case "1h":
		return time.Hour
	case "6h":
		return 6 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "24h":
		return 24 * time.Hour
	case "7d":
		return 7 * 24 * time.Hour
	case "30d":
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

func GetServerHistory(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// Ownership check
	var owner int64
	if err := db.DB.QueryRow("SELECT user_id FROM servers WHERE id = ?", id).Scan(&owner); err != nil || owner != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	dur := rangeToDuration(c.DefaultQuery("range", "24h"))
	since := time.Now().Add(-dur)

	rows, err := db.DB.Query(`SELECT id, server_id, timestamp, cpu_percent, ram_total, ram_used,
		disk_total, disk_used, load1, load5, load15, net_rx, net_tx, uptime_seconds
		FROM server_metrics WHERE server_id = ? AND timestamp >= ? ORDER BY timestamp ASC`, id, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	metrics := []models.ServerMetrics{}
	for rows.Next() {
		var m models.ServerMetrics
		if err := rows.Scan(&m.ID, &m.ServerID, &m.Timestamp, &m.CPUPercent, &m.RAMTotal, &m.RAMUsed,
			&m.DiskTotal, &m.DiskUsed, &m.Load1, &m.Load5, &m.Load15, &m.NetRX, &m.NetTX, &m.UptimeSeconds); err == nil {
			metrics = append(metrics, m)
		}
	}
	c.JSON(http.StatusOK, metrics)
}

func GetServerIncidents(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	rows, err := db.DB.Query(`SELECT id, user_id, monitor_type, monitor_id, alert_type, severity, status,
		start_time, end_time, acknowledged_by, acknowledged_at, resolved_at, ignored, comment, message
		FROM incidents WHERE user_id = ? AND monitor_type = 'server' AND monitor_id = ? ORDER BY start_time DESC`, userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	incidents := []models.Incident{}
	for rows.Next() {
		inc, err := scanIncident(rows)
		if err == nil {
			incidents = append(incidents, inc)
		}
	}
	c.JSON(http.StatusOK, incidents)
}

// ---- Agent Receiver ----

// AgentReceive handles POST /api/v1/agent/:key
func AgentReceive(c *gin.Context) {
	key := c.Param("key")

	// Find server by key
	var serverID, userID int64
	err := db.DB.QueryRow("SELECT id, user_id FROM servers WHERE server_key = ?", key).Scan(&serverID, &userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid server key"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Buffer the body first so we can retry gzip vs raw safely.
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	body := raw
	// Try gzip if declared or magic bytes match.
	if c.GetHeader("Content-Encoding") == "gzip" || (len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b) {
		gz, gzErr := gzip.NewReader(bytes.NewReader(raw))
		if gzErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gzip payload"})
			return
		}
		defer gz.Close()
		body, err = io.ReadAll(gz)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decompress body"})
			return
		}
	}

	var payload models.AgentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Aggregate disk totals
	var diskTotal, diskUsed int64
	for _, d := range payload.Disks {
		diskTotal += d.TotalBytes
		diskUsed += d.UsedBytes
	}

	rawData, _ := json.Marshal(payload)

	_, err = db.DB.Exec(`INSERT INTO server_metrics
		(server_id, cpu_percent, ram_total, ram_used, disk_total, disk_used, load1, load5, load15, net_rx, net_tx, uptime_seconds, raw_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		serverID, payload.CPUPercent, payload.RAMTotalBytes, payload.RAMUsedBytes,
		diskTotal, diskUsed, payload.LoadAvg[0], payload.LoadAvg[1], payload.LoadAvg[2],
		payload.NetRXBytes, payload.NetTXBytes, payload.UptimeSeconds, string(rawData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store metrics"})
		return
	}

	// Update server metadata
	_, _ = db.DB.Exec(`UPDATE servers SET last_seen = datetime('now'), status = 'online',
		hostname = ?, os = ?, agent_version = ? WHERE id = ?`,
		payload.Hostname, payload.OS, payload.AgentVersion, serverID)

	// Upsert services (docker + pm2 + system services all together)
	upsertServices(serverID, payload)

	// Resolve any active nodata incidents for this server.
	resolveNodataIncidents(userID, serverID)

	// Broadcast to WebSocket clients for this user.
	if ws.DefaultHub != nil {
		ws.DefaultHub.Broadcast(userID, "server_update", gin.H{
			"server_id":   serverID,
			"cpu_percent": payload.CPUPercent,
			"ram_used":    payload.RAMUsedBytes,
			"ram_total":   payload.RAMTotalBytes,
			"disk_used":   diskUsed,
			"disk_total":  diskTotal,
			"status":      "online",
			"hostname":    payload.Hostname,
		})
	}

	// Evaluate alert rules against this metric.
	alerts.EvaluateServerMetrics(userID, serverID, &payload, diskTotal, diskUsed)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func upsertServices(serverID int64, payload models.AgentPayload) {
	seen := map[string]string{}
	for _, dc := range payload.DockerContainers {
		status := dc.State
		if status == "" {
			status = dc.Status
		}
		seen["docker:"+dc.Name] = status
	}
	for _, p := range payload.PM2Processes {
		seen["pm2:"+p.Name] = p.Status
	}
	for _, s := range payload.Services {
		status := "stopped"
		if s.Running {
			status = "running"
		}
		seen["service:"+s.Name] = status
	}
	for name, status := range seen {
		_, _ = db.DB.Exec(`INSERT INTO server_services (server_id, name, status, updated_at)
			VALUES (?, ?, ?, datetime('now'))
			ON CONFLICT(server_id, name) DO UPDATE SET status = excluded.status, updated_at = datetime('now')`,
			serverID, name, status)
	}
}

func resolveNodataIncidents(userID, serverID int64) {
	_, _ = db.DB.Exec(`UPDATE incidents SET status = 'resolved', resolved_at = datetime('now'), end_time = datetime('now')
		WHERE user_id = ? AND monitor_type = 'server' AND monitor_id = ? AND alert_type = 'nodata' AND status IN ('active','acknowledged')`,
		userID, serverID)
}
