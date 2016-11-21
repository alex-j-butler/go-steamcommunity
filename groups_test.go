package steamcommunity_test

import (
	"net/http"
	"net/url"

	steamcommunity "alex-j-butler.com/steamcommunity"

	"github.com/stretchr/testify/assert"
)

func (s *ClientTestSuite) TestGroupSuccess() {
	// Setup
	s.ResponseFunc = []func(w http.ResponseWriter, r *http.Request){
		// RSA request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "publickey_mod": "B2EA82EF448EF21E5CAE3955432FB9D496307DB940EB93EEB7C8722C9F70625C4BC7A43C18CC9D0A5D40B1146406EB43384CD9B6601A871CDFE7327BD812616E0A8E0BCFA7EAC00239CA00FBF3BC408CA7E00BC62B1DBE429FBC7CABA760E2308A7C2384383BC42BE4DF6CC22A5208A747AF124CB2A0790098679450A400CE1ACC01D2BDA670FB5C17D62401B142FB0596662C5C58C7C78B3E76CBE9CD29681D96E0B3BD227088E7E308B747A2840E0E602D035860C3475D05145BB85C358D03674E1B2AD525E6AEE18BAC33B6D2D595C80BB1B09D1541924AB3958D54B28FA9CD78D823F850CED8AA74E99B55265329F8F3BCC3C493D7D89675B50A03258B3F", "publickey_exp": "010001", "timestamp": "457478400000", "token_gid": "69965557473581a"}`))
		},

		// Login request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "requires_twofactor": false, "redirect_uri": "steammobile:\/\/mobileloginsucceeded", "login_complete": true, "oauth": "{\"steamid\":\"\",\"oauth_token\":\"2NLM616F1X0D8IAJSLYRDIQQZXDIGXP4\",\"wgtoken\":\"326E6C6D36313666317830643869616A736C7972\",\"wgtoken_secure\":\"326E6C6D36313666317830643869616A736C7972\"}"}`))
		},

		// Group request.
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
				<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
				<memberList>
					<groupID64>103582791454641428</groupID64>
					<groupDetails>
						<groupName><![CDATA[shival]]></groupName>
						<groupURL><![CDATA[shival]]></groupURL>
						<headline><![CDATA[]]></headline>
						<summary><![CDATA[No information given.]]></summary>
						<avatarIcon><![CDATA[https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/fe/fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb.jpg]]></avatarIcon>
						<avatarMedium><![CDATA[https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/fe/fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb_medium.jpg]]></avatarMedium>
						<avatarFull><![CDATA[https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/fe/fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb_full.jpg]]></avatarFull>
						<memberCount>2</memberCount>
						<membersInChat>0</membersInChat>
						<membersInGame>0</membersInGame>
						<membersOnline>0</membersOnline>
					</groupDetails>
					<memberCount>2</memberCount>
					<totalPages>1</totalPages>
					<currentPage>1</currentPage>
					<startingMember>0</startingMember>
					<members>
						<steamID64>76561198063808035</steamID64>
						<steamID64>76561198333828103</steamID64>
					</members>
				</memberList>
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

	group, err := s.Client.Group("shival")

	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "103582791454641428", group.ID)
	assert.Equal(s.T(), "shival", group.Name)
	assert.Equal(s.T(), "shival", group.URL)
	assert.Equal(s.T(), "", group.Headline)
	assert.Equal(s.T(), "No information given.", group.Summary)
	assert.Equal(s.T(), "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/fe/fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb.jpg", group.AvatarIcon)
	assert.Equal(s.T(), "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/fe/fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb_medium.jpg", group.AvatarMedium)
	assert.Equal(s.T(), "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/fe/fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb_full.jpg", group.AvatarFull)
	assert.EqualValues(s.T(), []string{"76561198063808035", "76561198333828103"}, group.Members)
	assert.Equal(s.T(), 0, group.MembersInChat)
	assert.Equal(s.T(), 0, group.MembersInGame)
	assert.Equal(s.T(), 0, group.MembersOnline)
}
