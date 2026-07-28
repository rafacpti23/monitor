package api

import (
	"database/sql"
	"net/http"

	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

type RegisterReq struct {
	Email       string `json:"email" binding:"required,email"`
	Name        string `json:"name" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
	CompanyName string `json:"company_name"`
}

func Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create a fresh company for the new admin.
	companyName := req.CompanyName
	if companyName == "" {
		companyName = req.Name + "'s Workspace"
	}
	compRes, err := db.DB.Exec(
		`INSERT INTO companies (name, system_name, logo_url, accent_color) VALUES (?, ?, '', '#00e676')`,
		companyName, companyName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}
	companyID, _ := compRes.LastInsertId()

	res, err := db.DB.Exec(
		`INSERT INTO users (company_id, email, password_hash, name, role) VALUES (?, ?, ?, ?, 'admin')`,
		companyID, req.Email, hash, req.Name)
	if err != nil {
		// Roll back the orphan company on user insert failure.
		_, _ = db.DB.Exec("DELETE FROM companies WHERE id = ?", companyID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email already exists or DB error"})
		return
	}

	userID, _ := res.LastInsertId()
	db.LogSystemAction(userID, "user_registered", "User registered", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message":    "Registration successful",
		"user_id":    userID,
		"company_id": companyID,
	})
}

type LoginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := db.DB.QueryRow("SELECT id, company_id, email, password_hash, name, role FROM users WHERE email = ?", req.Email).
		Scan(&user.ID, &user.CompanyID, &user.Email, &user.PasswordHash, &user.Name, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		db.LogSystemAction(user.ID, "login_failed", "Failed login attempt", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	db.LogSystemAction(user.ID, "user_logged_in", "User logged in", c.ClientIP())

	// Fetch company for whitelabel branding in the login payload.
	company := loadCompany(user.CompanyID)

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"company_id": user.CompanyID,
		},
		"company": company,
	})
}

func Me(c *gin.Context) {
	userID := auth.GetUserID(c)
	var user models.User
	err := db.DB.QueryRow("SELECT id, company_id, email, name, role, whatsapp_number, timezone, created_at FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.CompanyID, &user.Email, &user.Name, &user.Role, &user.WhatsappNumber, &user.Timezone, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	company := loadCompany(user.CompanyID)
	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"company": company,
	})
}

// loadCompany fetches a company row or returns nil.
func loadCompany(id *int64) *models.Company {
	if id == nil {
		return nil
	}
	var comp models.Company
	err := db.DB.QueryRow(
		`SELECT id, name, system_name, logo_url, accent_color, created_at FROM companies WHERE id = ?`, *id).
		Scan(&comp.ID, &comp.Name, &comp.SystemName, &comp.LogoURL, &comp.AccentColor, &comp.CreatedAt)
	if err != nil {
		return nil
	}
	return &comp
}

func Logout(c *gin.Context) {
	userID := auth.GetUserID(c)
	// Invalidate any sessions for this user (best-effort; JWT is stateless but we clear stored sessions).
	_, _ = db.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	db.LogSystemAction(userID, "user_logged_out", "User logged out", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
