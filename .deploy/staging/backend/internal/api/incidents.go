package api

import (
	"net/http"
	"strconv"

	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

const incidentCols = `id, user_id, monitor_type, monitor_id, alert_type, severity, status, 
	start_time, end_time, acknowledged_by, acknowledged_at, resolved_at, ignored, comment, message`

func scanIncident(rows interface {
	Scan(dest ...interface{}) error
}) (models.Incident, error) {
	var i models.Incident
	err := rows.Scan(&i.ID, &i.UserID, &i.MonitorType, &i.MonitorID, &i.AlertType, &i.Severity, &i.Status,
		&i.StartTime, &i.EndTime, &i.AcknowledgedBy, &i.AcknowledgedAt, &i.ResolvedAt, &i.Ignored, &i.Comment, &i.Message)
	return i, err
}

func ListIncidents(c *gin.Context) {
	userID := auth.GetUserID(c)
	status := c.Query("status") // active, acknowledged, resolved
	monitorType := c.Query("monitor_type")

	query := `SELECT ` + incidentCols + ` FROM incidents WHERE user_id = ?`
	args := []interface{}{userID}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	if monitorType != "" {
		query += ` AND monitor_type = ?`
		args = append(args, monitorType)
	}
	query += ` ORDER BY start_time DESC LIMIT 100`

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	incidents := []models.Incident{}
	for rows.Next() {
		i, err := scanIncident(rows)
		if err == nil {
			incidents = append(incidents, i)
		}
	}
	c.JSON(http.StatusOK, incidents)
}

func AcknowledgeIncident(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	res, err := db.DB.Exec(`UPDATE incidents SET status = 'acknowledged', acknowledged_by = ?, acknowledged_at = datetime('now')
		WHERE id = ? AND user_id = ? AND status = 'active'`, userID, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found or not active"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Acknowledged"})
}

func ResolveIncident(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	res, err := db.DB.Exec(`UPDATE incidents SET status = 'resolved', resolved_at = datetime('now'), end_time = datetime('now')
		WHERE id = ? AND user_id = ? AND status IN ('active', 'acknowledged')`, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found or already resolved"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resolved"})
}

type IgnoreReq struct {
	Comment string `json:"comment"`
}

func IgnoreIncident(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req IgnoreReq
	_ = c.ShouldBindJSON(&req)

	res, err := db.DB.Exec(`UPDATE incidents SET ignored = 1, comment = ? WHERE id = ? AND user_id = ?`, req.Comment, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ignore"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Ignored"})
}
