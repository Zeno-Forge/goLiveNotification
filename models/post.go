package models

import "time"

type Post struct {
	ID         string         `json:"id"`
	Template   bool           `json:"template"`
	ScheduleAt time.Time      `json:"schedule_at"`
	Message    DiscordMessage `json:"message"`
}

type DiscordMessage struct {
	Content string  `json:"content"`
	Embed   []Embed `json:"embeds"`
}

type Embed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Color       int    `json:"color"`
	Thumbnail   URL    `json:"thumbnail"`
	Image       URL    `json:"image"`
	Footer      Footer `json:"footer"`
}

type URL struct {
	URL string `json:"url"`
}

type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}
