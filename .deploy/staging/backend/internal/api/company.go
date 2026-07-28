package api

import (
	"database/sql"
	"net/http"

	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

// resolveCompanyID reads the caller's company_id or aborts with 400.
func resolveCompanyID(c *gin.Context) (int64, bool) {
	userID := auth.GetUserID(c)
	var companyID sql.NullInt64
	err := db.DB.QueryRow("SELECT company_id FROM users WHERE id = ?", userID).Scan(&companyID)
	if err != nil || !companyID.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has no company"})
		return 0, false
	}
	return companyID.Int64, true
}

// GetCompany returns the caller's company (whitelabel branding).
func GetCompany(c *gin.Context) {
	companyID, ok := resolveCompanyID(c)
	if !ok {
		return
	}
	var comp models.Company
	err := db.DB.QueryRow(
		`SELECT id, name, system_name, logo_url, accent_color, created_at FROM companies WHERE id = ?`, companyID).
		Scan(&comp.ID, &comp.Name, &comp.SystemName, &comp.LogoURL, &comp.AccentColor, &comp.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}
	c.JSON(http.StatusOK, comp)
}

type CompanyUpdateReq struct {
	Name        string `json:"name"`
	SystemName  string `json:"system_name"`
	LogoURL     string `json:"logo_url"`
	AccentColor string `json:"accent_color"`
}

// UpdateCompany updates whitelabel settings — admin only.
func UpdateCompany(c *gin.Context) {
	if auth.GetUserRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin role required"})
		return
	}
	companyID, ok := resolveCompanyID(c)
	if !ok {
		return
	}
	var req CompanyUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.AccentColor == "" {
		req.AccentColor = "#00e676"
	}
	_, err := db.DB.Exec(
		`UPDATE companies SET name = ?, system_name = ?, logo_url = ?, accent_color = ? WHERE id = ?`,
		req.Name, req.SystemName, req.LogoURL, req.AccentColor, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}
	db.LogSystemAction(auth.GetUserID(c), "company_updated", "Whitelabel updated", c.ClientIP())

	var comp models.Company
	_ = db.DB.QueryRow(
		`SELECT id, name, system_name, logo_url, accent_color, created_at FROM companies WHERE id = ?`, companyID).
		Scan(&comp.ID, &comp.Name, &comp.SystemName, &comp.LogoURL, &comp.AccentColor, &comp.CreatedAt)
	c.JSON(http.StatusOK, comp)
}

// ---- Internal Users (within a company) ----

// ListCompanyUsers returns all users in the caller's company.
func ListCompanyUsers(c *gin.Context) {
	companyID, ok := resolveCompanyID(c)
	if !ok {
		return
	}
	rows, err := db.DB.Query(
		`SELECT id, company_id, email, name, role, whatsapp_number, timezone, created_at
		 FROM users WHERE company_id = ? ORDER BY created_at ASC`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()
	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.CompanyID, &u.Email, &u.Name, &u.Role,
			&u.WhatsappNumber, &u.Timezone, &u.CreatedAt); err == nil {
			users = append(users, u)
		}
	}
	c.JSON(http.StatusOK, users)
}

type CompanyUserCreateReq struct {
	Email          string `json:"email" binding:"required,email"`
	Name           string `json:"name" binding:"required"`
	Password       string `json:"password" binding:"required,min=8"`
	Role           string `json:"role"`
	WhatsappNumber string `json:"whatsapp_number"`
}

// CreateCompanyUser adds an internal user — admin only.
func CreateCompanyUser(c *gin.Context) {
	if auth.GetUserRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin role required"})
		return
	}
	companyID, ok := resolveCompanyID(c)
	if !ok {
		return
	}
	var req CompanyUserCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "admin" && req.Role != "member" && req.Role != "viewer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	res, err := db.DB.Exec(
		`INSERT INTO users (company_id, email, password_hash, name, role, whatsapp_number)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		companyID, req.Email, hash, req.Name, req.Role, req.WhatsappNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
		return
	}
	id, _ := res.LastInsertId()
	db.LogSystemAction(auth.GetUserID(c), "user_created", "Created internal user "+req.Email, c.ClientIP())

	var u models.User
	_ = db.DB.QueryRow(
		`SELECT id, company_id, email, name, role, whatsapp_number, timezone, created_at
		 FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.CompanyID, &u.Email, &u.Name, &u.Role, &u.WhatsappNumber, &u.Timezone, &u.CreatedAt)
	c.JSON(http.StatusOK, u)
}

type CompanyUserUpdateReq struct {
	Name           string `json:"name"`
	Role           string `json:"role"`
	WhatsappNumber string `json:"whatsapp_number"`
	Password       string `json:"password"` // optional; only updates when non-empty
}

// UpdateCompanyUser edits an internal user — admin only.
func UpdateCompanyUser(c *gin.Context) {
	if auth.GetUserRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin role required"})
		return
	}
	companyID, ok := resolveCompanyID(c)
	if !ok {
		return
	}
	targetID := c.Param("id")
	var req CompanyUserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role != "" && req.Role != "admin" && req.Role != "member" && req.Role != "viewer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		_, err = db.DB.Exec(
			`UPDATE users SET name = ?, role = COALESCE(NULLIF(?, ''), role), whatsapp_number = ?, password_hash = ?
			 WHERE id = ? AND company_id = ?`,
			req.Name, req.Role, req.WhatsappNumber, hash, targetID, companyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}
	} else {
		_, err := db.DB.Exec(
			`UPDATE users SET name = ?, role = COALESCE(NULLIF(?, ''), role), whatsapp_number = ?
			 WHERE id = ? AND company_id = ?`,
			req.Name, req.Role, req.WhatsappNumber, targetID, companyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}
	}
	db.LogSystemAction(auth.GetUserID(c), "user_updated", "Updated user "+targetID, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

// DeleteCompanyUser removes an internal user — admin only, can't delete self.
func DeleteCompanyUser(c *gin.Context) {
	if auth.GetUserRole(c) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin role required"})
		return
	}
	companyID, ok := resolveCompanyID(c)
	if !ok {
		return
	}
	targetID := c.Param("id")
	callerID := auth.GetUserID(c)

	// Prevent self-deletion.
	var candidate int64
	if err := db.DB.QueryRow("SELECT id FROM users WHERE id = ? AND company_id = ?", targetID, companyID).Scan(&candidate); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in company"})
		return
	}
	if candidate == callerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete yourself"})
		return
	}
	_, err := db.DB.Exec("DELETE FROM users WHERE id = ? AND company_id = ?", targetID, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	db.LogSystemAction(callerID, "user_deleted", "Deleted user "+targetID, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}
