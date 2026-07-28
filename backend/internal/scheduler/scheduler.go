package scheduler

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"p-mon-backend/internal/alerts"
	"p-mon-backend/internal/checks"
	"p-mon-backend/internal/db"
	"p-mon-backend/internal/ws"
	"p-mon-backend/pkg/models"

	"github.com/gin-gonic/gin"
)

func Start() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			runWebsiteChecks()
			runGenericChecks()
			checkServerNodata()
		}
	}()

	// Data retention: runs once per hour
	retentionTicker := time.NewTicker(1 * time.Hour)
	go func() {
		// Run once on startup after a short delay
		time.Sleep(10 * time.Second)
		runRetention()
		for range retentionTicker.C {
			runRetention()
		}
	}()

	log.Println("Scheduler started (30s interval, 1h retention)")
}

// runRetention deletes old history data to prevent unbounded DB growth.
// Retention periods:
//   - server_metrics: 30 days
//   - website_history: 90 days
//   - check_history: 90 days
//   - system_logs: 90 days
//   - expired sessions: immediate
//   - resolved incidents older than 90 days
// After cleanup, VACUUM is attempted to reclaim disk space.
func runRetention() {
	queries := []struct {
		label string
		sql   string
	}{
		{"server_metrics (>7d)", `DELETE FROM server_metrics WHERE timestamp < datetime('now', '-7 days')`},
		{"website_history (>90d)", `DELETE FROM website_history WHERE timestamp < datetime('now', '-90 days')`},
		{"check_history (>90d)", `DELETE FROM check_history WHERE timestamp < datetime('now', '-90 days')`},
		{"system_logs (>90d)", `DELETE FROM system_logs WHERE timestamp < datetime('now', '-90 days')`},
		{"expired sessions", `DELETE FROM sessions WHERE expires_at < datetime('now')`},
		{"old resolved incidents (>90d)", `DELETE FROM incidents WHERE status = 'resolved' AND resolved_at < datetime('now', '-90 days')`},
		{"old notifications (>90d)", `DELETE FROM notifications_sent WHERE sent_at < datetime('now', '-90 days')`},
	}

	totalDeleted := int64(0)
	for _, q := range queries {
		res, err := db.DB.Exec(q.sql)
		if err != nil {
			log.Printf("[retention] error cleaning %s: %v", q.label, err)
			continue
		}
		n, _ := res.RowsAffected()
		if n > 0 {
			log.Printf("[retention] cleaned %d rows from %s", n, q.label)
			totalDeleted += n
		}
	}

	if totalDeleted > 0 {
		_, err := db.DB.Exec("VACUUM")
		if err != nil {
			log.Printf("[retention] VACUUM error: %v", err)
		} else {
			log.Printf("[retention] VACUUM completed, reclaimed disk space")
		}
	}
}

func runWebsiteChecks() {
	// Only fetch websites that are due for a check based on their own
	// check_interval_sec. A site last checked N seconds ago is due when
	// N >= check_interval_sec. Sites never checked (last_checked IS NULL)
	// are always due.
	rows, err := db.DB.Query(`SELECT id, user_id, url, search_string FROM websites
		WHERE last_checked IS NULL
		   OR (strftime('%s','now') - strftime('%s', last_checked)) >= check_interval_sec`)
	if err != nil {
		return
	}
	defer rows.Close()

	due := []models.Website{}
	for rows.Next() {
		var w models.Website
		if err := rows.Scan(&w.ID, &w.UserID, &w.URL, &w.SearchString); err != nil {
			continue
		}
		due = append(due, w)
	}

	for _, w := range due {
		ok, ms := checks.CheckHTTP(w.URL, w.SearchString)
		status := "up"
		statusOk := 1
		code := 200
		if !ok {
			status = "down"
			statusOk = 0
			code = 0
		}

		_, _ = db.DB.Exec(`INSERT INTO website_history (website_id, response_code, response_time_ms, status_ok)
			VALUES (?, ?, ?, ?)`, w.ID, code, ms, statusOk)

		_, _ = db.DB.Exec(`UPDATE websites SET status = ?, last_checked = datetime('now'), last_response_code = ?, last_response_time_ms = ?
			WHERE id = ?`, status, code, ms, w.ID)

		if !ok {
			alerts.CreateIncident(w.UserID, "website", w.ID, "website_down", "critical", "Website "+w.URL+" is down", "[]")
		} else {
			// Resolve incident if it was active
			_, _ = db.DB.Exec(`UPDATE incidents SET status = 'resolved', resolved_at = datetime('now'), end_time = datetime('now')
				WHERE user_id = ? AND monitor_type = 'website' AND monitor_id = ? AND alert_type = 'website_down' AND status IN ('active','acknowledged')`,
				w.UserID, w.ID)
		}
	}
}

func runGenericChecks() {
	rows, err := db.DB.Query(`SELECT id, user_id, type, target, port, expected_response FROM checks`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Check
		if err := rows.Scan(&c.ID, &c.UserID, &c.Type, &c.Target, &c.Port, &c.ExpectedResponse); err != nil {
			continue
		}
		ok, ms := false, 0
		switch c.Type {
		case "ping":
			ok, ms = checks.CheckPing(c.Target)
		case "tcp":
			ok, ms = checks.CheckTCP(c.Target, c.Port)
		case "dns":
			ok, ms = checks.CheckDNS(c.Target, c.ExpectedResponse)
		}

		status := "down"
		statusOk := 0
		if ok {
			status = "up"
			statusOk = 1
		}

		_, _ = db.DB.Exec(`INSERT INTO check_history (check_id, result_ok, response_time_ms) VALUES (?, ?, ?)`, c.ID, statusOk, ms)
		_, _ = db.DB.Exec(`UPDATE checks SET status = ?, last_checked = datetime('now'), last_result_ok = ?, last_response_time_ms = ? WHERE id = ?`,
			status, statusOk, ms, c.ID)

		alerts.EvaluateCheckResult(c.UserID, c.ID, ok, ms)
		if !ok {
			alerts.CreateIncident(c.UserID, "check", c.ID, "check_failed", "warning", "Check "+c.Name+" failed", "[]")
		}
	}
}

func checkServerNodata() {
	// Dynamic cutoff per server: 2 × agent interval, with a floor of 2 minutes
	// (to avoid false positives on transient network hiccups) and a fallback
	// of 10 minutes when the server hasn't reported its interval yet.
	now := time.Now()
	rows, err := db.DB.Query(`SELECT id, user_id, name, interval_seconds, last_seen FROM servers WHERE status = 'online'`)
	if err != nil {
		return
	}
	defer rows.Close()

	type off struct {
		id, uid    int64
		name       string
		gapSeconds int64
	}
	offs := []off{}

	for rows.Next() {
		var id, uid int64
		var name string
		var intervalSec int
		var lastSeen sql.NullTime
		if err := rows.Scan(&id, &uid, &name, &intervalSec, &lastSeen); err != nil {
			continue
		}
		if !lastSeen.Valid {
			continue
		}

		var cutoffDur time.Duration
		if intervalSec > 0 {
			cutoffDur = time.Duration(intervalSec*2) * time.Second
			if cutoffDur < 2*time.Minute {
				cutoffDur = 2 * time.Minute
			}
		} else {
			cutoffDur = 10 * time.Minute
		}

		if lastSeen.Time.Before(now.Add(-cutoffDur)) {
			gap := int64(now.Sub(lastSeen.Time).Seconds())
			offs = append(offs, off{id, uid, name, gap})
		}
	}

	for _, o := range offs {
		_, _ = db.DB.Exec("UPDATE servers SET status = 'offline' WHERE id = ?", o.id)
		gapMin := o.gapSeconds / 60
		msg := "Servidor " + o.name + " está offline (sem dados há " + strconv.FormatInt(gapMin, 10) + "min)"
		alerts.CreateIncident(o.uid, "server", o.id, "nodata", "critical", msg, "[]")

		// Notify live dashboards that this server went offline.
		if ws.DefaultHub != nil {
			ws.DefaultHub.Broadcast(o.uid, "server_offline", gin.H{
				"server_id": o.id,
				"status":    "offline",
			})
		}
	}
}


