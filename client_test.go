package steamcommunity_test

import (
	"net/http"
	"net/url"
	"testing"

	steamcommunity "alex-j-butler.com/steamcommunity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) TestSuccess() {
	// Setup
	s.ResponseFunc = []func(w http.ResponseWriter, r *http.Request){
		// RSA request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
			{
				"success": true,
				"publickey_mod": "B2EA82EF448EF21E5CAE3955432FB9D496307DB940EB93EEB7C8722C9F70625C4BC7A43C18CC9D0A5D40B1146406EB43384CD9B6601A871CDFE7327BD812616E0A8E0BCFA7EAC00239CA00FBF3BC408CA7E00BC62B1DBE429FBC7CABA760E2308A7C2384383BC42BE4DF6CC22A5208A747AF124CB2A0790098679450A400CE1ACC01D2BDA670FB5C17D62401B142FB0596662C5C58C7C78B3E76CBE9CD29681D96E0B3BD227088E7E308B747A2840E0E602D035860C3475D05145BB85C358D03674E1B2AD525E6AEE18BAC33B6D2D595C80BB1B09D1541924AB3958D54B28FA9CD78D823F850CED8AA74E99B55265329F8F3BCC3C493D7D89675B50A03258B3F",
				"publickey_exp": "010001",
				"timestamp": "457478400000",
				"token_gid": "69965557473581a"
			}`))
		},

		// Login request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
			{
				"success": true,
				"requires_twofactor": false,
				"redirect_uri": "steammobile:\/\/mobileloginsucceeded",
				"login_complete": true,
				"oauth": "{\"steamid\":\"\",\"oauth_token\":\"2NLM616F1X0D8IAJSLYRDIQQZXDIGXP4\",\"wgtoken\":\"326E6C6D36313666317830643869616A736C7972\",\"wgtoken_secure\":\"326E6C6D36313666317830643869616A736C7972\"}"
			}
			`))
		},
	}

	var err error
	s.Client, err = steamcommunity.New(&steamcommunity.LoginDetails{
		AccountName: "example",
		Password:    "example",
		Transport: RewriteTransport{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(s.Server.URL)
				},
			},
		},
	})

	assert.NoError(s.T(), err)
}

func (s *ClientTestSuite) TestIncorrectLogin() {
	// Setup
	s.ResponseFunc = []func(w http.ResponseWriter, r *http.Request){
		// RSA request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
			{
				"success": true,
				"publickey_mod": "B2EA82EF448EF21E5CAE3955432FB9D496307DB940EB93EEB7C8722C9F70625C4BC7A43C18CC9D0A5D40B1146406EB43384CD9B6601A871CDFE7327BD812616E0A8E0BCFA7EAC00239CA00FBF3BC408CA7E00BC62B1DBE429FBC7CABA760E2308A7C2384383BC42BE4DF6CC22A5208A747AF124CB2A0790098679450A400CE1ACC01D2BDA670FB5C17D62401B142FB0596662C5C58C7C78B3E76CBE9CD29681D96E0B3BD227088E7E308B747A2840E0E602D035860C3475D05145BB85C358D03674E1B2AD525E6AEE18BAC33B6D2D595C80BB1B09D1541924AB3958D54B28FA9CD78D823F850CED8AA74E99B55265329F8F3BCC3C493D7D89675B50A03258B3F",
				"publickey_exp": "010001",
				"timestamp": "457478400000",
				"token_gid": "69965557473581a"
			}`))
		},

		// Login request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
			{
				"success": false,
				"requires_twofactor": false,
				"clear_password_field": true,
				"captcha_needed": false,
				"captcha_gid": -1,
				"message": "Incorrect login."
			}
			`))
		},
	}

	var err error
	s.Client, err = steamcommunity.New(&steamcommunity.LoginDetails{
		AccountName: "example",
		Password:    "incorrect",
		Transport: RewriteTransport{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(s.Server.URL)
				},
			},
		},
	})

	assert.EqualError(s.T(), err, "steamcommunity: Incorrect login.")
}

func (s *ClientTestSuite) TestEmailAuthFailure() {
	// Setup
	s.ResponseFunc = []func(w http.ResponseWriter, r *http.Request){
		// RSA request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
			{
				"success": true,
				"publickey_mod": "B2EA82EF448EF21E5CAE3955432FB9D496307DB940EB93EEB7C8722C9F70625C4BC7A43C18CC9D0A5D40B1146406EB43384CD9B6601A871CDFE7327BD812616E0A8E0BCFA7EAC00239CA00FBF3BC408CA7E00BC62B1DBE429FBC7CABA760E2308A7C2384383BC42BE4DF6CC22A5208A747AF124CB2A0790098679450A400CE1ACC01D2BDA670FB5C17D62401B142FB0596662C5C58C7C78B3E76CBE9CD29681D96E0B3BD227088E7E308B747A2840E0E602D035860C3475D05145BB85C358D03674E1B2AD525E6AEE18BAC33B6D2D595C80BB1B09D1541924AB3958D54B28FA9CD78D823F850CED8AA74E99B55265329F8F3BCC3C493D7D89675B50A03258B3F",
				"publickey_exp": "010001",
				"timestamp": "457478400000",
				"token_gid": "69965557473581a"
			}`))
		},

		// Login request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
			{
				"success": false,
				"requires_twofactor": false,
				"message": "",
				"emailauth_needed": true,
				"emaildomain": "gmail.com",
				"emailsteamid": ""
			}
			`))
		},
	}

	var err error
	s.Client, err = steamcommunity.New(&steamcommunity.LoginDetails{
		AccountName: "example",
		Password:    "example",
		Transport: RewriteTransport{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(s.Server.URL)
				},
			},
		},
	})

	assert.EqualError(s.T(), err, steamcommunity.ErrorEmailAuth.Error())
}

func (s *ClientTestSuite) TestMobileAuthFailure() {
	// Setup
	s.ResponseFunc = []func(w http.ResponseWriter, r *http.Request){
		// RSA request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
			{
				"success": true,
				"publickey_mod": "B2EA82EF448EF21E5CAE3955432FB9D496307DB940EB93EEB7C8722C9F70625C4BC7A43C18CC9D0A5D40B1146406EB43384CD9B6601A871CDFE7327BD812616E0A8E0BCFA7EAC00239CA00FBF3BC408CA7E00BC62B1DBE429FBC7CABA760E2308A7C2384383BC42BE4DF6CC22A5208A747AF124CB2A0790098679450A400CE1ACC01D2BDA670FB5C17D62401B142FB0596662C5C58C7C78B3E76CBE9CD29681D96E0B3BD227088E7E308B747A2840E0E602D035860C3475D05145BB85C358D03674E1B2AD525E6AEE18BAC33B6D2D595C80BB1B09D1541924AB3958D54B28FA9CD78D823F850CED8AA74E99B55265329F8F3BCC3C493D7D89675B50A03258B3F",
				"publickey_exp": "010001",
				"timestamp": "457478400000",
				"token_gid": "69965557473581a"
			}`))
		},

		// Login request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
			{
				"success": false,
				"requires_twofactor": true,
				"message": ""
			}
			`))
		},
	}

	var err error
	s.Client, err = steamcommunity.New(&steamcommunity.LoginDetails{
		AccountName: "example",
		Password:    "example",
		Transport: RewriteTransport{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(s.Server.URL)
				},
			},
		},
	})

	assert.EqualError(s.T(), err, steamcommunity.ErrorMobileAuth.Error())
}
