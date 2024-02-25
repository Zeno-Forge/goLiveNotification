package models

type AppConfig struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Settings Settings `json:"settings"`
}

type Settings struct {
	Theme          string       `json:"theme"`
	DiscordWebhook string       `json:"discord_webhook"`
	PostTemplate   PostTemplate `json:"post_template"`
}

type PostTemplate struct {
	Message DiscordMessage `json:"message"`
}
