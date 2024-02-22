package models

type ServerConfig struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
}

type DiscordConfig struct {
	Discord struct {
		WebhookUrl string `json:"webhook_url"`
		RoleID     string `json:"role_id"`
	} `json:"discord"`
}
