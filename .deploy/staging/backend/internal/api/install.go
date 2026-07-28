package api

import (
	"net/http"

	"p-mon-backend/internal/agent"
	"p-mon-backend/internal/config"

	"github.com/gin-gonic/gin"
)

func ServeInstallScript(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.String(http.StatusBadRequest, "echo 'Missing key parameter'")
		return
	}

	backendURL := config.Cfg.PublicURL
	if backendURL == "" {
		backendURL = "http://" + c.Request.Host
	}

	script, err := agent.Render(backendURL, key)
	if err != nil {
		c.String(http.StatusInternalServerError, "echo 'Failed to render install script: %s'", err.Error())
		return
	}

	c.Data(http.StatusOK, "text/x-shellscript", []byte(script))
}
