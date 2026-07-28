package api

import (
	"encoding/json"
	"net/http"

	"p-mon-backend/internal/alerts"
	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

const channelCols = `id, user_id, type, name, config, enabled`

func scanChannel(rows interface {
	Scan(dest ...interface{}) error
}) (models.AlertChannel, error) {
	var ac models.AlertChannel
	err := rows.Scan(&ac.ID, &ac.UserID, &ac.Type, &ac.Name, &ac.Config, &ac.Enabled)
	return ac, err
}

func ListChannels(c *gin.Context) {
	userID := auth.GetUserID(c)
	rows, err := db.DB.Query(`SELECT `+channelCols+` FROM alert_channels WHERE user_id = ? ORDER BY name`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()
	chs := []models.AlertChannel{}
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err == nil {
			chs = append(chs, ch)
		}
	}
	c.JSON(http.StatusOK, chs)
}

type ChannelReq struct {
	Type    string `json:"type" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Config  string `json:"config" binding:"required"` // JSON string
	Enabled bool   `json:"enabled"`
}

func CreateChannel(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req ChannelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := db.DB.Exec(`INSERT INTO alert_channels (user_id, type, name, config, enabled) VALUES (?, ?, ?, ?, ?)`,
		userID, req.Type, req.Name, req.Config, req.Enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}
	id, _ := res.LastInsertId()
	db.LogSystemAction(userID, "channel_created", "Created channel "+req.Name, c.ClientIP())

	row := db.DB.QueryRow(`SELECT `+channelCols+` FROM alert_channels WHERE id = ?`, id)
	ch, _ := scanChannel(row)
	c.JSON(http.StatusOK, ch)
}

func UpdateChannel(c *gin.Context) {
	userID := auth.GetUserID(c)
	id := c.Param("id")
	var req ChannelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec(`UPDATE alert_channels SET type = ?, name = ?, config = ?, enabled = ? WHERE id = ? AND user_id = ?`,
		req.Type, req.Name, req.Config, req.Enabled, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
		return
	}
	db.LogSystemAction(userID, "channel_updated", "Updated channel "+id, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func DeleteChannel(c *gin.Context) {
	userID := auth.GetUserID(c)
	id := c.Param("id")
	_, err := db.DB.Exec("DELETE FROM alert_channels WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	db.LogSystemAction(userID, "channel_deleted", "Deleted channel "+id, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func TestChannel(c *gin.Context) {
	userID := auth.GetUserID(c)
	id := c.Param("id")

	var chType, configJSON string
	err := db.DB.QueryRow(`SELECT type, config FROM alert_channels WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&chType, &configJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	var cfg map[string]string
	_ = json.Unmarshal([]byte(configJSON), &cfg)
	if cfg == nil {
		cfg = map[string]string{}
	}

	subject := "[P-mon] Test notification"
	body := "This is a test message from P-mon."

	switch chType {
	case "email":
		err = alerts.SendEmail(cfg["to"], subject, body)
	case "whatsapp":
		// PAPI: config has instance, api_key, jid.
		err = alerts.SendWhatsAppPAPI(cfg["instance"], cfg["api_key"], cfg["jid"], subject+"\n"+body)
	case "webhook":
		payload, _ := json.Marshal(map[string]interface{}{"subject": subject, "body": body})
		err = alerts.SendWebhook(cfg["url"], string(payload))
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown channel type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Send failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Test message sent successfully"})
}
