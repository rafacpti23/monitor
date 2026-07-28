package api

import (
	"net/http"
	"strconv"

	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

const websiteCols = `id, user_id, group_id, name, url, method, check_interval_sec, search_string, status, last_checked, last_response_code, last_response_time_ms, ssl_expires_at, created_at`

func scanWebsite(rows interface {
	Scan(dest ...interface{}) error
}) (models.Website, error) {
	var w models.Website
	err := rows.Scan(&w.ID, &w.UserID, &w.GroupID, &w.Name, &w.URL, &w.Method, &w.CheckIntervalSec,
		&w.SearchString, &w.Status, &w.LastChecked, &w.LastResponseCode, &w.LastResponseTimeMs, &w.SSLExpiresAt, &w.CreatedAt)
	return w, err
}

func ListWebsites(c *gin.Context) {
	userID := auth.GetUserID(c)
	rows, err := db.DB.Query(`SELECT `+websiteCols+` FROM websites WHERE user_id = ? ORDER BY name`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	websites := []models.Website{}
	for rows.Next() {
		w, err := scanWebsite(rows)
		if err == nil {
			websites = append(websites, w)
		}
	}
	c.JSON(http.StatusOK, websites)
}

type WebsiteReq struct {
	Name             string `json:"name" binding:"required"`
	URL              string `json:"url" binding:"required"`
	Method           string `json:"method"`
	CheckIntervalSec int    `json:"check_interval_sec"`
	SearchString     string `json:"search_string"`
	GroupID          *int64 `json:"group_id"`
}

func CreateWebsite(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req WebsiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.CheckIntervalSec == 0 {
		req.CheckIntervalSec = 60
	}

	res, err := db.DB.Exec(`INSERT INTO websites (user_id, group_id, name, url, method, check_interval_sec, search_string, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')`,
		userID, req.GroupID, req.Name, req.URL, req.Method, req.CheckIntervalSec, req.SearchString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create website"})
		return
	}
	id, _ := res.LastInsertId()
	db.LogSystemAction(userID, "website_created", "Created website "+req.Name, c.ClientIP())

	row := db.DB.QueryRow(`SELECT `+websiteCols+` FROM websites WHERE id = ?`, id)
	w, err := scanWebsite(row)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Created but failed to load"})
		return
	}
	c.JSON(http.StatusOK, w)
}

func GetWebsite(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	row := db.DB.QueryRow(`SELECT `+websiteCols+` FROM websites WHERE id = ? AND user_id = ?`, id, userID)
	w, err := scanWebsite(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}
	c.JSON(http.StatusOK, w)
}

func UpdateWebsite(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req WebsiteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := db.DB.Exec(`UPDATE websites SET name = ?, url = ?, method = ?, check_interval_sec = ?, search_string = ?, group_id = ?
		WHERE id = ? AND user_id = ?`,
		req.Name, req.URL, req.Method, req.CheckIntervalSec, req.SearchString, req.GroupID, id, userID)
	if err != nil || res == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update website"})
		return
	}
	db.LogSystemAction(userID, "website_updated", "Updated website "+strconv.FormatInt(id, 10), c.ClientIP())
	GetWebsite(c)
}

func DeleteWebsite(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, _ = db.DB.Exec("DELETE FROM incidents WHERE user_id = ? AND monitor_type = 'website' AND monitor_id = ?", userID, id)
	res, err := db.DB.Exec("DELETE FROM websites WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}
	db.LogSystemAction(userID, "website_deleted", "Deleted website "+strconv.FormatInt(id, 10), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func GetWebsiteIncidents(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	rows, err := db.DB.Query(`SELECT id, user_id, monitor_type, monitor_id, alert_type, severity, status,
		start_time, end_time, acknowledged_by, acknowledged_at, resolved_at, ignored, comment, message
		FROM incidents WHERE user_id = ? AND monitor_type = 'website' AND monitor_id = ?
		ORDER BY start_time DESC LIMIT 15`, userID, id)
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

func GetWebsiteHistory(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var dummy int
	if err := db.DB.QueryRow("SELECT id FROM websites WHERE id = ? AND user_id = ?", id, userID).Scan(&dummy); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}

	rows, err := db.DB.Query(`SELECT id, website_id, timestamp, response_code, response_time_ms, status_ok 
		FROM website_history WHERE website_id = ? ORDER BY timestamp DESC LIMIT 50`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()
	hist := []models.WebsiteHistory{}
	for rows.Next() {
		var h models.WebsiteHistory
		if err := rows.Scan(&h.ID, &h.WebsiteID, &h.Timestamp, &h.ResponseCode, &h.ResponseTimeMs, &h.StatusOK); err == nil {
			hist = append(hist, h)
		}
	}
	c.JSON(http.StatusOK, hist)
}
