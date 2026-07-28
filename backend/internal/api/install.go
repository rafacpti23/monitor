package api

import (
	"net/http"
	"strconv"
	"strings"

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

	// Parse collect options from query string. Defaults: system=true,
	// docker=true, pm2=true (matching agent defaults).
	opts := agent.CollectOpts{
		System: queryBool(c, "system", true),
		Docker: queryBool(c, "docker", true),
		PM2:    queryBool(c, "pm2", true),
	}

	if svc := c.Query("services"); svc != "" {
		for _, s := range strings.Split(svc, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				opts.Services = append(opts.Services, s)
			}
		}
	}

	if iv := c.Query("interval"); iv != "" {
		if n, err := strconv.Atoi(iv); err == nil && n > 0 {
			opts.Interval = n
		}
	}

	script, err := agent.Render(backendURL, key, opts)
	if err != nil {
		c.String(http.StatusInternalServerError, "echo 'Failed to render install script: %s'", err.Error())
		return
	}

	c.Data(http.StatusOK, "text/x-shellscript", []byte(script))
}

// queryBool returns the query param as a boolean: "0", "false" → false;
// anything else (or absent) → defaultVal.
func queryBool(c *gin.Context, key string, defaultVal bool) bool {
	v := c.Query(key)
	if v == "" {
		return defaultVal
	}
	v = strings.ToLower(v)
	if v == "0" || v == "false" || v == "no" {
		return false
	}
	if v == "1" || v == "true" || v == "yes" {
		return true
	}
	return defaultVal
}
