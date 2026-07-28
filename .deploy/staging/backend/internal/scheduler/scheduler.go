package scheduler

import (
	"database/sql"
	"log"
	"time"

	"p-mon-backend/internal/alerts"
	"p-mon-backend/internal/checks"
	"p-mon-backend/internal/db"
	"p-mon-backend/pkg/models"
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
	log.Println("Scheduler started (30s interval)")
}

func runWebsiteChecks() {
	rows, err := db.DB.Query(`SELECT id, user_id, url, search_string FROM websites`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var w models.Website
		if err := rows.Scan(&w.ID, &w.UserID, &w.URL, &w.SearchString); err != nil {
			continue
		}

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
	// Find servers that haven't reported in 5 minutes
	cutoff := time.Now().Add(-5 * time.Minute)
	rows, err := db.DB.Query(`SELECT id, user_id, name, last_seen FROM servers WHERE status = 'online'`)
	if err != nil {
		return
	}
	defer rows.Close()

	type off struct {
		id, uid int64
		name    string
	}
	offs := []off{}

	for rows.Next() {
		var id, uid int64
		var name string
		var lastSeen sql.NullTime
		if err := rows.Scan(&id, &uid, &name, &lastSeen); err == nil {
			if lastSeen.Valid && lastSeen.Time.Before(cutoff) {
				offs = append(offs, off{id, uid, name})
			}
		}
	}

	for _, o := range offs {
		_, _ = db.DB.Exec("UPDATE servers SET status = 'offline' WHERE id = ?", o.id)
		alerts.CreateIncident(o.uid, "server", o.id, "nodata", "critical", "Server "+o.name+" is offline (no data >5m)", "[]")
	}
}
