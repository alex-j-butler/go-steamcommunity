package steamcommunity

import (
	"net/http"
)

type LoginDetails struct {
	AccountName   string
	Password      string
	SteamGuard    string
	AuthCode      string
	TwoFactorCode string
	Captcha       string
	Transport     http.RoundTripper
}
