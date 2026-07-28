package api

import (
	"net/http"
	"strconv"

	"p-mon-backend/internal/auth"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

const checkCols = `id, user_id, name, type, target, port, expected_response, interval_sec, status, last_checked, last_result_ok, last_response_time_ms, created_at`

func scanCheck(rows interface {
	Scan(dest ...interface{}) error
}) (models.Check, error) {
	var ch models.Check
	err := rows.Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Type, &ch.Target, &ch.Port, &ch.ExpectedResponse,
		&ch.IntervalSec, &ch.Status, &ch.LastChecked, &ch.LastResultOK, &ch.LastResponseTimeMs, &ch.CreatedAt)
	return ch, err
}

func ListChecks(c *gin.Context) {
	userID := auth.GetUserID(c)
	rows, err := db.DB.Query(`SELECT `+checkCols+` FROM checks WHERE user_id = ? ORDER BY name`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	checks := []models.Check{}
	for rows.Next() {
		ch, err := scanCheck(rows)
		if err == nil {
			checks = append(checks, ch)
		}
	}
	c.JSON(http.StatusOK, checks)
}

type CheckReq struct {
	Name             string `json:"name" binding:"required"`
	Type             string `json:"type" binding:"required"`
	Target           string `json:"target" binding:"required"`
	Port             int    `json:"port"`
	ExpectedResponse string `json:"expected_response"`
	IntervalSec      int    `json:"interval_sec"`
}

func CreateCheck(c *gin.Context) {
	userID := auth.GetUserID(c)
	var req CheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IntervalSec == 0 {
		req.IntervalSec = 60
	}

	res, err := db.DB.Exec(`INSERT INTO checks (user_id, name, type, target, port, expected_response, interval_sec, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')`,
		userID, req.Name, req.Type, req.Target, req.Port, req.ExpectedResponse, req.IntervalSec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create check"})
		return
	}
	id, _ := res.LastInsertId()
	db.LogSystemAction(userID, "check_created", "Created check "+req.Name, c.ClientIP())

	row := db.DB.QueryRow(`SELECT `+checkCols+` FROM checks WHERE id = ?`, id)
	ch, _ := scanCheck(row)
	c.JSON(http.StatusOK, ch)
}

func GetCheck(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	row := db.DB.QueryRow(`SELECT `+checkCols+` FROM checks WHERE id = ? AND user_id = ?`, id, userID)
	ch, err := scanCheck(row)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Check not found"})
		return
	}
	c.JSON(http.StatusOK, ch)
}

func UpdateCheck(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req CheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := db.DB.Exec(`UPDATE checks SET name = ?, type = ?, target = ?, port = ?, expected_response = ?, interval_sec = ?
		WHERE id = ? AND user_id = ?`,
		req.Name, req.Type, req.Target, req.Port, req.ExpectedResponse, req.IntervalSec, id, userID)
	if err != nil || res == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update check"})
		return
	}
	db.LogSystemAction(userID, "check_updated", "Updated check "+strconv.FormatInt(id, 10), c.ClientIP())
	GetCheck(c)
}

func DeleteCheck(c *gin.Context) {
	userID := auth.GetUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, _ = db.DB.Exec("DELETE FROM incidents WHERE user_id = ? AND monitor_type = 'check' AND monitor_id = ?", userID, id)
	res, err := db.DB.Exec("DELETE FROM checks WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Check not found"})
		return
	}
	db.LogSystemAction(userID, "check_deleted", "Deleted check "+strconv.FormatInt(id, 10), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}
