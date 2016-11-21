package steamcommunity

import (
	"errors"
	"io/ioutil"
	"log"
)

func (c *Client) GetNotifications() error {
	resp, err := c.get("https://steamcommunity.com/actions/GetNotificationCounts")

	if err != nil {
		return err
	}

	if resp.StatusCode == 403 {
		return errors.New("Unauthenticated")
	}

	log.Println("Response:", resp)
	log.Println("Error:", err)

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("Body:", string(body))

	return nil
}
