package api

import (
	"net/http"
	"strconv"

	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func ListTemplates(c *gin.Context) {
	userID := auth.GetUserID(c)

	rows, err := db.DB.Query(`SELECT id, user_id, alert_type, subject, body, created_at, updated_at FROM alert_templates WHERE user_id = ? ORDER BY alert_type`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	templates := []models.AlertTemplate{}
	for rows.Next() {
		var t models.AlertTemplate
		if err := rows.Scan(&t.ID, &t.UserID, &t.AlertType, &t.Subject, &t.Body, &t.CreatedAt, &t.UpdatedAt); err == nil {
			templates = append(templates, t)
		}
	}
	c.JSON(http.StatusOK, templates)
}

type CreateTemplateReq struct {
	AlertType string `json:"alert_type" binding:"required"`
	Subject   string `json:"subject" binding:"required"`
	Body      string `json:"body" binding:"required"`
}

func CreateTemplate(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req CreateTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	res, err := db.DB.Exec(`INSERT INTO alert_templates (user_id, alert_type, subject, body) VALUES (?, ?, ?, ?)`,
		userID, req.AlertType, req.Subject, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	id, _ := res.LastInsertId()
	template := models.AlertTemplate{
		ID:        id,
		UserID:    userID,
		AlertType: req.AlertType,
		Subject:   req.Subject,
		Body:      req.Body,
	}
	c.JSON(http.StatusOK, template)
}

type UpdateTemplateReq struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func UpdateTemplate(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req UpdateTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	res, err := db.DB.Exec(`UPDATE alert_templates SET subject = ?, body = ?, updated_at = datetime('now')
		WHERE id = ? AND user_id = ?`, req.Subject, req.Body, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func DeleteTemplate(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	res, err := db.DB.Exec(`DELETE FROM alert_templates WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func ResetTemplate(c *gin.Context) {
	userID := auth.GetUserID(c)
	alertType := c.Param("alert_type")

	_, _ = db.DB.Exec(`DELETE FROM alert_templates WHERE user_id = ? AND alert_type = ?`, userID, alertType)
	c.JSON(http.StatusOK, gin.H{"message": "Reset to default"})
}
