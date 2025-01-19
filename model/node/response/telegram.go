package response

type TelegramKey struct {
	AuthKey        string `json:"auth_key"`
	TelegramUserId string `json:"telegram_user_id"`
	Message        string `json:"message"`
}

type IntegrateTelegramKeyResponse struct {
	IntegrateKey []TelegramKey `json:"integrate_key"`
}
