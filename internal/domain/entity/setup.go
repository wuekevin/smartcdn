package entity

// CheckItem represents the result of a single environment check.
type CheckItem struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "ok" | "fail"
	Message string `json:"message,omitempty"`
}

// SetupCheckResult holds the full environment detection report.
type SetupCheckResult struct {
	MySQL  CheckItem `json:"mysql"`
	Redis  CheckItem `json:"redis"`
	GoVer  CheckItem `json:"go_version"`
	Ports  CheckItem `json:"ports"`
	Clock  CheckItem `json:"clock"`
	Perms  CheckItem `json:"permissions"`
}

// SetupCompleteInput carries the payload for the setup completion call.
type SetupCompleteInput struct {
	DBHost     string `json:"db_host" binding:"required"`
	DBPort     int    `json:"db_port" binding:"required"`
	DBName     string `json:"db_name" binding:"required"`
	DBUser     string `json:"db_user" binding:"required"`
	DBPassword string `json:"db_password"`
	RedisAddr     string `json:"redis_addr" binding:"required"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
	AdminUser  string `json:"admin_username" binding:"required"`
	AdminPass  string `json:"admin_password" binding:"required,min=8"`
	EncryptKey string `json:"encrypt_key" binding:"required"`
	SiteURL    string `json:"site_url" binding:"required"`
	Port       string `json:"port" binding:"required"`
}
