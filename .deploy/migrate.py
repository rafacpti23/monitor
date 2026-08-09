import sqlite3

db = sqlite3.connect('/var/lib/p-mon/p-mon.db')
db.execute("""
CREATE TABLE IF NOT EXISTS alert_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL CHECK(alert_type IN ('website_down','server_down','check_failed','nodata','resolved')),
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE(user_id, alert_type)
)
""")
db.commit()

# Verify
tables = db.execute("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name").fetchall()
print("Tables:", [t[0] for t in tables])
db.close()
print("Migration OK")
