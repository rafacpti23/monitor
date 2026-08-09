package models

import (
	"encoding/json"
	"time"
)

type Company struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	SystemName  string    `json:"system_name"`
	LogoURL     string    `json:"logo_url"`
	AccentColor string    `json:"accent_color"`
	CreatedAt   time.Time `json:"created_at"`
}

type User struct {
	ID             int64     `json:"id"`
	CompanyID      *int64    `json:"company_id"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	WhatsappNumber string    `json:"whatsapp_number"`
	Timezone       string    `json:"timezone"`
	CreatedAt      time.Time `json:"created_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type ServerGroup struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
}

type Server struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	GroupID      *int64     `json:"group_id"`
	Name         string     `json:"name"`
	ServerKey    string     `json:"server_key,omitempty"`
	Type         string     `json:"type"`
	OS           string     `json:"os"`
	Status       string     `json:"status"`
	OnMap        bool       `json:"on_map"`
	Lat          float64    `json:"lat"`
	Lng          float64    `json:"lng"`
	Hostname     string     `json:"hostname"`
	AgentVersion string     `json:"agent_version"`
	IntervalSeconds int     `json:"interval_seconds"`
	LastSeen     *time.Time `json:"last_seen"`
	CreatedAt    time.Time  `json:"created_at"`

	// Inline latest metrics (populated on list/detail)
	LatestMetrics *ServerMetrics  `json:"latest_metrics,omitempty"`
	Services      []ServerService `json:"services,omitempty"`
}

type ServerMetrics struct {
	ID            int64     `json:"id"`
	ServerID      int64     `json:"server_id"`
	Timestamp     time.Time `json:"timestamp"`
	CPUPercent    float64   `json:"cpu_percent"`
	RAMTotal      int64     `json:"ram_total"`
	RAMUsed       int64     `json:"ram_used"`
	DiskTotal     int64     `json:"disk_total"`
	DiskUsed      int64     `json:"disk_used"`
	Load1         float64   `json:"load1"`
	Load5         float64   `json:"load5"`
	Load15        float64   `json:"load15"`
	NetRX         int64     `json:"net_rx"`
	NetTX         int64     `json:"net_tx"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	RawData       string    `json:"raw_data,omitempty"`
}

type ServerService struct {
	ID        int64     `json:"id"`
	ServerID  int64     `json:"server_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WebsiteGroup struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
}

type Website struct {
	ID                 int64      `json:"id"`
	UserID             int64      `json:"user_id"`
	GroupID            *int64     `json:"group_id"`
	Name               string     `json:"name"`
	URL                string     `json:"url"`
	Method             string     `json:"method"`
	CheckIntervalSec   int        `json:"check_interval_sec"`
	SearchString       string     `json:"search_string"`
	Status             string     `json:"status"`
	LastChecked        *time.Time `json:"last_checked"`
	LastResponseCode   int        `json:"last_response_code"`
	LastResponseTimeMs int        `json:"last_response_time_ms"`
	SSLExpiresAt       *time.Time `json:"ssl_expires_at"`
	CreatedAt          time.Time  `json:"created_at"`
}

type WebsiteHistory struct {
	ID             int64     `json:"id"`
	WebsiteID      int64     `json:"website_id"`
	Timestamp      time.Time `json:"timestamp"`
	ResponseCode   int       `json:"response_code"`
	ResponseTimeMs int       `json:"response_time_ms"`
	StatusOK       bool      `json:"status_ok"`
}

type Check struct {
	ID                 int64      `json:"id"`
	UserID             int64      `json:"user_id"`
	Name               string     `json:"name"`
	Type               string     `json:"type"`
	Target             string     `json:"target"`
	Port               int        `json:"port"`
	ExpectedResponse   string     `json:"expected_response"`
	IntervalSec        int        `json:"interval_sec"`
	Status             string     `json:"status"`
	LastChecked        *time.Time `json:"last_checked"`
	LastResultOK       bool       `json:"last_result_ok"`
	LastResponseTimeMs int        `json:"last_response_time_ms"`
	CreatedAt          time.Time  `json:"created_at"`
}

type CheckHistory struct {
	ID             int64     `json:"id"`
	CheckID        int64     `json:"check_id"`
	Timestamp      time.Time `json:"timestamp"`
	ResultOK       bool      `json:"result_ok"`
	ResponseTimeMs int       `json:"response_time_ms"`
	Details        string    `json:"details"`
}

// PapiPanel represents a WhatsApp monitoring panel (PAPI or Stevo provider).
type PapiPanel struct {
	ID                 int64      `json:"id"`
	UserID             int64      `json:"user_id"`
	Name               string     `json:"name"`
	Provider           string     `json:"provider"`
	BaseURL            string     `json:"base_url"`
	PanelToken         string     `json:"panel_token,omitempty"`
	CheckIntervalSec   int        `json:"check_interval_sec"`
	Status             string     `json:"status"`
	LastChecked        *time.Time `json:"last_checked"`
	LastError          string     `json:"last_error"`
	TotalInstances     int        `json:"total_instances"`
	ConnectedInstances int        `json:"connected_instances"`
	Channels           string     `json:"channels"`
	CreatedAt          time.Time  `json:"created_at"`

	// Inline instances (populated on detail).
	Instances []PapiInstance `json:"instances,omitempty"`
}

// PapiInstance is a single WhatsApp instance discovered under a panel.
type PapiInstance struct {
	ID          int64      `json:"id"`
	PanelID     int64      `json:"panel_id"`
	UserID      int64      `json:"user_id"`
	InstanceID  string     `json:"instance_id"`
	Name        string     `json:"name"`
	PhoneNumber string     `json:"phone_number"`
	Status      string     `json:"status"`
	LastSeen    *time.Time `json:"last_seen"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PapiAPIResponse is the envelope returned by PAPI /api/v1/instances.
type PapiAPIResponse struct {
	Success   bool             `json:"success"`
	Count     int              `json:"count"`
	Instances []PapiAPIInstance `json:"instances"`
}

// PapiAPIInstance mirrors one instance in the PAPI response.
type PapiAPIInstance struct {
	ID             string `json:"id"`
	UpstreamID     string `json:"upstream_id"`
	Status         string `json:"status"`
	Name           string `json:"name"`
	PhoneConnected string `json:"phone_connected"`
	CreatedAt      string `json:"created_at"`
}

// StevoMCPRequest is the JSON-RPC request body for the Stevo MCP endpoint.
type StevoMCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// StevoMCPResponse is the JSON-RPC response from the Stevo MCP endpoint.
type StevoMCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *StevoMCPError  `json:"error,omitempty"`
}

// StevoMCPError represents an error in a Stevo MCP response.
type StevoMCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// StevoInstance represents a single instance from Stevo's list_instances result.
type StevoInstance struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Phone  string `json:"phone"`
}

type AlertRule struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	MonitorType string `json:"monitor_type"`
	MonitorID   int64  `json:"monitor_id"`
	AlertType   string `json:"alert_type"`
	Comparison  string `json:"comparison"`
	Threshold   string `json:"threshold"`
	Occurrences int    `json:"occurrences"`
	CooldownMin int    `json:"cooldown_min"`
	Status      string `json:"status"`
	Channels    string `json:"channels"`
}

type AlertChannel struct {
	ID      int64  `json:"id"`
	UserID  int64  `json:"user_id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Config  string `json:"config"`
	Enabled bool   `json:"enabled"`
}

type Incident struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	MonitorType    string     `json:"monitor_type"`
	MonitorID      int64      `json:"monitor_id"`
	AlertType      string     `json:"alert_type"`
	Severity       string     `json:"severity"`
	Status         string     `json:"status"`
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	AcknowledgedBy *int64     `json:"acknowledged_by"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	Ignored        bool       `json:"ignored"`
	Comment        string     `json:"comment"`
	Message        string     `json:"message"`
}

type NotificationSent struct {
	ID         int64     `json:"id"`
	IncidentID int64     `json:"incident_id"`
	ChannelID  int64     `json:"channel_id"`
	SentAt     time.Time `json:"sent_at"`
	Success    bool      `json:"success"`
}

type SystemLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

// AgentPayload is the JSON sent by the VPS agent.
type AgentPayload struct {
	Hostname         string            `json:"hostname"`
	OS               string            `json:"os"`
	Arch             string            `json:"arch"`
	UptimeSeconds    int64             `json:"uptime_seconds"`
	CPUPercent       float64           `json:"cpu_percent"`
	CPUCores         int               `json:"cpu_cores"`
	CPUModel         string            `json:"cpu_model"`
	LoadAvg          [3]float64        `json:"load_avg"`
	RAMTotalBytes    int64             `json:"ram_total_bytes"`
	RAMUsedBytes     int64             `json:"ram_used_bytes"`
	Disks            []AgentDisk       `json:"disks"`
	NetRXBytes       int64             `json:"net_rx_bytes"`
	NetTXBytes       int64             `json:"net_tx_bytes"`
	DockerContainers []AgentContainer  `json:"docker_containers"`
	PM2Processes     []AgentPM2Process `json:"pm2_processes"`
	Services         []AgentService    `json:"services"`
	AgentVersion     string            `json:"agent_version"`
	IntervalSeconds  int               `json:"interval_seconds"`
}

type AlertTemplate struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	AlertType string    `json:"alert_type"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AgentDisk struct {
	// Agent emits "path"; keep "mount" as a fallback for older payloads.
	Path       string `json:"path"`
	Mount      string `json:"mount"`
	TotalBytes int64  `json:"total_bytes"`
	UsedBytes  int64  `json:"used_bytes"`
}

type AgentContainer struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	State  string `json:"state"`
	Image  string `json:"image"`
}

type AgentPM2Process struct {
	Name   string  `json:"name"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	Memory int64   `json:"memory"`
}

type AgentService struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
	PID     int    `json:"pid,omitempty"`
}
