package steamcommunity

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var previousCaptchaGID string

var (
	ErrorRSARequest    = errors.New("steamcommunity: RSA key request failed")
	ErrorRSAResponse   = errors.New("steamcommunity: Malformed RSA key response")
	ErrorRSAEncrypt    = errors.New("steamcommunity: RSA key failed to encrypt")
	ErrorLoginFailed   = errors.New("steamcommunity: Login request failed to send")
	ErrorLoginResponse = errors.New("steamcommunity: Malformed login response")
	ErrorEmailAuth     = errors.New("steamcommunity: SteamGuard email auth required")
	ErrorMobileAuth    = errors.New("steamcommunity: SteamGuard mobile auth required")
	ErrorCaptcha       = errors.New("steamcommunity: CAPTCHA input required")
	ErrorUnknown       = errors.New("steamcommunity: Unknown error")
)

type Client struct {
	SteamID      string
	SessionID    string
	SteamGuardID string
	OAuthToken   string
	Cookies      []*http.Cookie

	client     *http.Client
	captchaGID string
}

type loginResponse struct {
	Success           bool   `json:"success"`
	RequiresEmailAuth bool   `json:"emailauth_needed"`
	RequiresTwoFactor bool   `json:"requires_twofactor"`
	RequiresCaptcha   bool   `json:"captcha_needed"`
	Message           string `json:"message"`
	OAuth             string `json:"oauth"`
}

type loginCaptchaResponse struct {
	CaptchaGID string `json:"captcha_gid"`
}

type oauthResponse struct {
	SteamID    string `json:"steamid"`
	OAuthToken string `json:"oauth_token"`
}

type rsaResponse struct {
	Success           bool   `json:"success"`
	PublicKeyMod      string `json:"publickey_mod"`
	PublicKeyExponent string `json:"publickey_exp"`
	Timestamp         string `json:"timestamp"`
	Token             string `json:"token_gid"`
}

type pinResponse struct {
	Success bool `json:"success"`
}

func (r rsaResponse) GetModulus() *big.Int {
	by, _ := hex.DecodeString(r.PublicKeyMod)
	i := new(big.Int)
	in := i.SetBytes(by)
	return in
}

func (r rsaResponse) GetExponent() int {
	i, _ := strconv.ParseInt(r.PublicKeyExponent, 16, 32)
	return int(i)
}

func New(details *LoginDetails) (*Client, error) {
	jar, _ := cookiejar.New(nil)

	client := &Client{
		client: &http.Client{Jar: jar, Transport: details.Transport},
	}

	client.setCookie(&http.Cookie{Name: "mobileClientVersion", Value: "0 (2.1.3)"}, true)
	client.setCookie(&http.Cookie{Name: "mobileClient", Value: "android"}, true)

	resp, err := client.postForm(
		"https://steamcommunity.com/login/getrsakey",
		map[string]string{
			"X-Requested-With": "com.valvesoftware.android.steam.community",
			"Referer":          "https://steamcommunity.com/mobilelogin?oauth_client_id=DE45CD61&oauth_scope=read_profile%20write_profile%20read_client%20write_client",
			"User-Agent":       "Mozilla/5.0 (Linux; U; Android 4.1.1; en-us; Google Nexus 4 - 4.1.1 - API 16 - 768x1280 Build/JRO03S) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30",
			"Accept":           "text/javascript, text/html, application/xml, text/xml, */*",
		},
		map[string]string{
			"username": details.AccountName,
		},
	)

	if err != nil {
		return nil, ErrorRSARequest
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	var response rsaResponse
	err = json.Unmarshal(respBody, &response)

	if err != nil {
		return nil, ErrorRSAResponse
	}

	// Encrypt the password with the RSA key.
	publicKey := rsa.PublicKey{N: response.GetModulus(), E: response.GetExponent()}
	pass, err := rsa.EncryptPKCS1v15(rand.Reader, &publicKey, []byte(details.Password))
	base64Pass := base64.StdEncoding.EncodeToString(pass)

	if err != nil {
		return nil, ErrorRSAEncrypt
	}

	resp, err = client.postForm(
		"https://steamcommunity.com/login/dologin",
		map[string]string{
			"X-Requested-With": "com.valvesoftware.android.steam.community",
			"Referer":          "https://steamcommunity.com/mobilelogin?oauth_client_id=DE45CD61&oauth_scope=read_profile%20write_profile%20read_client%20write_client",
			"User-Agent":       "Mozilla/5.0 (Linux; U; Android 4.1.1; en-us; Google Nexus 4 - 4.1.1 - API 16 - 768x1280 Build/JRO03S) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30",
			"Accept":           "text/javascript, text/html, application/xml, text/xml, */*",
		},
		map[string]string{
			"captcha_text":      details.Captcha,
			"captchagid":        previousCaptchaGID,
			"emailauth":         details.AuthCode,
			"emailsteamid":      "",
			"password":          base64Pass,
			"remember_login":    "true",
			"rsatimestamp":      response.Timestamp,
			"twofactorcode":     details.TwoFactorCode,
			"username":          details.AccountName,
			"oauth_client_id":   "DE45CD61",
			"oauth_scope":       "read_profile write_profile read_client write_client",
			"loginfriendlyname": "#login_emailauth_friendlyname_mobile",
			"donotcache":        strconv.FormatInt(time.Now().Unix(), 10),
		},
	)

	if err != nil {
		return nil, ErrorLoginFailed
	}

	respBody, _ = ioutil.ReadAll(resp.Body)
	var logResponse loginResponse
	err = json.Unmarshal(respBody, &logResponse)

	var logCaptchaResponse loginCaptchaResponse
	json.Unmarshal(respBody, &logCaptchaResponse)

	if err != nil {
		log.Println(err)
		return nil, ErrorLoginResponse
	}

	if !logResponse.Success && logResponse.RequiresEmailAuth {
		// Requires SteamGuard auth from email.
		return nil, ErrorEmailAuth
	}

	if !logResponse.Success && logResponse.RequiresTwoFactor {
		// Requires SteamGuard auth from mobile app.
		return nil, ErrorMobileAuth
	}

	if !logResponse.Success && logResponse.RequiresCaptcha {
		// Requires CAPTCHA.
		client.captchaGID = logCaptchaResponse.CaptchaGID
		previousCaptchaGID = logCaptchaResponse.CaptchaGID
		return client, ErrorCaptcha
	}

	if !logResponse.Success {
		if logResponse.Message != "" {
			return nil, errors.New(fmt.Sprintf("steamcommunity: %s", logResponse.Message))
		}

		return nil, ErrorUnknown
	}

	if logResponse.OAuth == "" {
		return nil, ErrorLoginResponse
	}

	// Generate a session ID.
	sessionID, err := generateSessionID()

	if err != nil {
		return nil, err
	}

	var oauthResp oauthResponse
	err = json.Unmarshal([]byte(logResponse.OAuth), &oauthResp)

	// Set SessionID cookie.
	client.setCookie(&http.Cookie{Name: "sessionid", Value: sessionID}, true)

	cookies := client.client.Jar.Cookies(&url.URL{Scheme: "https", Host: "steamcommunity.com"})
	var steamguard string
	for _, cookie := range cookies {
		if cookie.Name == fmt.Sprintf("steamMachineAuth%s", oauthResp.SteamID) {
			steamguard = fmt.Sprintf("%s||%s", oauthResp.SteamID, cookie.Value)
		}
	}

	// Set all cookies on the global client.
	client.setCookies(cookies, true)

	// Populate the client.
	client.SessionID = sessionID
	client.Cookies = cookies
	client.SteamGuardID = steamguard
	client.SteamID = oauthResp.SteamID
	client.OAuthToken = oauthResp.OAuthToken

	return client, nil
}

// GetCaptchaURL returns the URL of the CAPTCHA image. This will only be populated if a login was attempted, but returned a CAPTCHA error.
func (c *Client) GetCaptchaURL() (string, error) {
	if c.captchaGID != "" {
		return fmt.Sprintf("https://steamcommunity.com/login/rendercaptcha/?gid=%s", c.captchaGID), nil
	}

	return "", errors.New("No CAPTCHA available")
}

// ParentalUnlock is used to unlock a Steam account from the parental controls.
// Error is nil on success, otherwise it will contain the error message.
func (c *Client) ParentalUnlock(pin string) error {
	resp, err := c.postForm(
		"https://steamcommunity.com/parental/ajaxunlock",
		map[string]string{},
		map[string]string{
			"pin": pin,
		},
	)

	if err != nil {
		return errors.New("Failed to send PIN request")
	}

	var pinResp pinResponse
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &pinResp)

	if err != nil {
		return errors.New("Failed to unmarshal PIN response")
	}

	if !pinResp.Success {
		return errors.New("Incorrect PIN")
	}

	return nil
}

func (c *Client) setCookie(cookie *http.Cookie, secure bool) {
	var protocol string
	if secure {
		protocol = "https"
	} else {
		protocol = "http"
	}

	hosts := []string{"steamcommunity.com", "store.steampowered.com", "help.steampowered.com"}

	for _, host := range hosts {
		c.client.Jar.SetCookies(
			&url.URL{Scheme: protocol, Host: host},
			[]*http.Cookie{cookie},
		)
	}
}

func (c *Client) setCookies(cookies []*http.Cookie, secure bool) {
	for _, cookie := range cookies {
		c.setCookie(cookie, secure)
	}
}

func (c *Client) get(uri string) (*http.Response, error) {
	return c.client.Get(uri)
}

func (c *Client) postForm(uri string, headers map[string]string, form map[string]string) (*http.Response, error) {
	values := url.Values{}
	for k, v := range form {
		values.Add(k, v)
	}

	req, err := http.NewRequest("POST", uri, strings.NewReader(values.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.client.Do(req)
}
