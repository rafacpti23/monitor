package api

import (
	"net/http"

	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"

	"github.com/gin-gonic/gin"
)

type SettingsReq struct {
	Timezone       string `json:"timezone"`
	WhatsappNumber string `json:"whatsapp_number"`
}

func GetSettings(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req SettingsReq
	err := db.DB.QueryRow("SELECT timezone, whatsapp_number FROM users WHERE id = ?", userID).
		Scan(&req.Timezone, &req.WhatsappNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	c.JSON(http.StatusOK, req)
}

func UpdateSettings(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req SettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.DB.Exec("UPDATE users SET timezone = ?, whatsapp_number = ? WHERE id = ?", req.Timezone, req.WhatsappNumber, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}
	db.LogSystemAction(userID, "settings_updated", "Updated user settings", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}
