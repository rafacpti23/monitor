package api

import (
	"net/http"
	"strconv"

	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

const ruleCols = `id, user_id, monitor_type, monitor_id, alert_type, comparison, threshold, occurrences, cooldown_min, status, channels`

func scanRule(rows interface {
	Scan(dest ...interface{}) error
}) (models.AlertRule, error) {
	var r models.AlertRule
	err := rows.Scan(&r.ID, &r.UserID, &r.MonitorType, &r.MonitorID, &r.AlertType, &r.Comparison, &r.Threshold, &r.Occurrences, &r.CooldownMin, &r.Status, &r.Channels)
	return r, err
}

func ListRules(c *gin.Context) {
	userID := auth.GetUserID(c)
	monitorType := c.Query("monitor_type")
	monitorID := c.Query("monitor_id")

	query := `SELECT ` + ruleCols + ` FROM alert_rules WHERE user_id = ?`
	args := []interface{}{userID}

	if monitorType != "" && monitorID != "" {
		query += ` AND monitor_type = ? AND monitor_id = ?`
		args = append(args, monitorType, monitorID)
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	rules := []models.AlertRule{}
	for rows.Next() {
		r, err := scanRule(rows)
		if err == nil {
			rules = append(rules, r)
		}
	}
	c.JSON(http.StatusOK, rules)
}

type RuleReq struct {
	MonitorType string `json:"monitor_type" binding:"required"`
	MonitorID   int64  `json:"monitor_id" binding:"required"`
	AlertType   string `json:"alert_type" binding:"required"`
	Comparison  string `json:"comparison" binding:"required"`
	Threshold   string `json:"threshold" binding:"required"`
	Occurrences int    `json:"occurrences"`
	CooldownMin int    `json:"cooldown_min"`
	Channels    string `json:"channels"` // JSON array string
}

func CreateRule(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req RuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Occurrences == 0 {
		req.Occurrences = 1
	}
	if req.CooldownMin == 0 {
		req.CooldownMin = 30
	}
	if req.Channels == "" {
		req.Channels = "[]"
	}

	res, err := db.DB.Exec(`INSERT INTO alert_rules (user_id, monitor_type, monitor_id, alert_type, comparison, threshold, occurrences, cooldown_min, status, channels) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'enabled', ?)`,
		userID, req.MonitorType, req.MonitorID, req.AlertType, req.Comparison, req.Threshold, req.Occurrences, req.CooldownMin, req.Channels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule"})
		return
	}
	id, _ := res.LastInsertId()
	db.LogSystemAction(userID, "rule_created", "Created rule for "+req.MonitorType, c.ClientIP())

	row := db.DB.QueryRow(`SELECT `+ruleCols+` FROM alert_rules WHERE id = ?`, id)
	r, _ := scanRule(row)
	c.JSON(http.StatusOK, r)
}

func UpdateRule(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req RuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`UPDATE alert_rules SET monitor_type = ?, monitor_id = ?, alert_type = ?, comparison = ?, threshold = ?, occurrences = ?, cooldown_min = ?, channels = ? 
		WHERE id = ? AND user_id = ?`,
		req.MonitorType, req.MonitorID, req.AlertType, req.Comparison, req.Threshold, req.Occurrences, req.CooldownMin, req.Channels, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule"})
		return
	}
	db.LogSystemAction(userID, "rule_updated", "Updated rule "+strconv.FormatInt(id, 10), c.ClientIP())

	row := db.DB.QueryRow(`SELECT `+ruleCols+` FROM alert_rules WHERE id = ?`, id)
	r, _ := scanRule(row)
	c.JSON(http.StatusOK, r)
}

func DeleteRule(c *gin.Context) {
	userID := auth.GetUserID(c)
	id := c.Param("id")
	_, err := db.DB.Exec("DELETE FROM alert_rules WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	db.LogSystemAction(userID, "rule_deleted", "Deleted rule "+id, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}
