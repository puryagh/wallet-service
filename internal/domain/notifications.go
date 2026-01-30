package domain

type SMSNotification struct {
	PhoneNumber string   `json:"phone_number" mapstructure:"phone_number"`
	Pattern     string   `json:"pattern" mapstructure:"pattern"`
	PatternID   string   `json:"pattern_id" mapstructure:"pattern_id"`
	Arguments   []string `json:"arguments" mapstructure:"arguments"`
}
