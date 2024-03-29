package utl

import (
	"errors"
	"github.com/ashwanthkumar/slack-go-webhook"
	"log"
	"time"
)

type SlackNotif struct {
	conf *SlackConfig
}

func NewSlackNotif() *SlackNotif {
	return &SlackNotif{
		conf: GetSlackConfig(),
	}
}

func (sn *SlackNotif) SendSlackNotification(backupStatus BackupStatus, backupName, errorMsg string) error {
	payload := slack.Payload{}
	attachment1 := slack.Attachment{}
	attachment1.AddField(slack.Field{Title: "Date", Value: time.Now().String()}).AddField(slack.Field{Title: "DCName",
		Value: "🏢 " + sn.conf.DC}).AddField(slack.Field{Title: "Type", Value: sn.conf.BackupType})

	if backupStatus == FAILED {
		attachment1.AddField(slack.Field{Title: "Status", Value: "❌ Failed"}).AddField(slack.Field{Title: "Debug", Value: errorMsg})
		payload = slack.Payload{
			Text:        string(FAILEDMSG),
			Username:    "backup-robot",
			Channel:     sn.conf.SlackChannelName,
			IconEmoji:   ":warning:",
			Attachments: []slack.Attachment{attachment1},
		}
	} else if backupStatus == SUCCEEDED {
		attachment1.AddField(slack.Field{Title: "Status", Value: "✅ succeeded"}).AddField(slack.Field{Title: "Name",
			Value: backupName})
		payload = slack.Payload{
			Text:        string(SUCCEEDEDMSG),
			Username:    "backup-robot",
			Channel:     sn.conf.SlackChannelName,
			IconEmoji:   ":ok:",
			Attachments: []slack.Attachment{attachment1},
		}
	} else {
		log.Println("Error utl  BackupStatus is not supported")
		return errors.New("BackupStatus is not supported")
	}

	err := slack.Send(sn.conf.SlackWebhookUrl, "", payload)
	if err != nil {
		log.Println("Error utl  Sending notification is failed")
		return errors.New("Sending notification is failed")
	}
	log.Println("Info utl  Notification Sent Successfully!")
	return nil
}
