package steamcommunity

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
)

type groupMemberList struct {
	GroupID64         string       `xml:"groupID64"`
	GroupDetails      groupDetails `xml:"groupDetails"`
	MemberCount       int          `xml:"memberCount"`
	MemberTotalPages  int          `xml:"totalPages"`
	MemberCurrentPage int          `xml:"currentPage"`
	Members           groupMembers `xml:"members"`
}

type groupDetails struct {
	GroupName     string `xml:"groupName"`
	GroupURL      string `xml:"groupURL"`
	Headline      string `xml:"headline"`
	Summary       string `xml:"summary"`
	AvatarIcon    string `xml:"avatarIcon"`
	AvatarMedium  string `xml:"avatarMedium"`
	AvatarFull    string `xml:"avatarFull"`
	MemberCount   int    `xml:"memberCount"`
	MembersInChat int    `xml:"membersInChat"`
	MembersInGame int    `xml:"membersInGame"`
	MembersOnline int    `xml:"membersOnline"`
}

type groupMembers struct {
	SteamID64 []string `xml:"steamID64"`
}

type Group struct {
	ID            string
	Name          string
	URL           string
	Headline      string
	Summary       string
	AvatarIcon    string
	AvatarMedium  string
	AvatarFull    string
	Members       []string
	MembersInChat int
	MembersInGame int
	MembersOnline int

	client *Client
}

// Group retrieves a Steam Group by group URL.
func (c *Client) Group(groupID string) (*Group, error) {
	resp, err := c.get(
		fmt.Sprintf("https://steamcommunity.com/groups/%s/memberslistxml/?xml=1", groupID),
	)

	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var xmlResp groupMemberList
	err = xml.Unmarshal(body, &xmlResp)

	if err != nil {
		return nil, err
	}

	group := &Group{}

	// Populate group.
	group.ID = xmlResp.GroupID64
	group.Name = xmlResp.GroupDetails.GroupName
	group.URL = xmlResp.GroupDetails.GroupURL
	group.Headline = xmlResp.GroupDetails.Headline
	group.Summary = xmlResp.GroupDetails.Summary
	group.AvatarIcon = xmlResp.GroupDetails.AvatarIcon
	group.AvatarMedium = xmlResp.GroupDetails.AvatarMedium
	group.AvatarFull = xmlResp.GroupDetails.AvatarFull
	group.Members = xmlResp.Members.SteamID64
	group.MembersInChat = xmlResp.GroupDetails.MembersInChat
	group.MembersInGame = xmlResp.GroupDetails.MembersInGame
	group.MembersOnline = xmlResp.GroupDetails.MembersOnline
	group.client = c

	return group, nil
}

// PostAnnouncement sends a request to create a new Steam Group announcement.
// headline specifies the headline of the announcement.
// content specifies the content of the announcement.
func (g *Group) PostAnnouncement(headline string, content string) error {
	resp, err := g.client.postForm(
		fmt.Sprintf("https://steamcommunity.com/gid/%s/announcements", g.ID),
		map[string]string{},
		map[string]string{
			"sessionID": g.client.SessionID,
			"action":    "post",
			"headline":  headline,
			"body":      content,
			"languages[0][headline]": headline,
			"languages[0][body]":     content,
		},
	)

	if err != nil {
		return err
	}

	if resp.StatusCode == 403 {
		return errors.New("Unauthenticated")
	}

	if resp.StatusCode != 200 {
		return errors.New("Unknown error")
	}

	return nil
}
