package shaker

import (
	"fmt"
	"os"

	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

func (s *Shaker) isSlackEnabled() bool {
	if s.config.Slack.Enabled {
		return true
	}
	return false
}

func (s *Shaker) createSlackConnection() {
	s.Log().Info("Connection to slack")
	s.connectors.slackConfig.enabled = true
	s.connectors.slackConfig.client = slack.New(s.config.Slack.Token)
	s.connectors.slackConfig.channel = s.config.Slack.Channel
}

func slackSendMessage(slackConfig slackConfig, name string, text string, color string, severity string, url string, elapsedTime float64) {
	if !slackConfig.enabled {
		return
	}

	log.WithFields(log.Fields{
		"context": "shaker",
		"type":    "slack",
	})

	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("can't get hostname: %s", err)
	}

	attachment := slack.Attachment{
		Color: color,
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "Shaker host",
				Value: hostname,
			},
			slack.AttachmentField{
				Title: "Job",
				Value: name,
			},
			slack.AttachmentField{
				Title: severity,
				Value: text,
			},
		},
	}

	if url != "" {
		attachment.Fields = append(attachment.Fields,
			slack.AttachmentField{
				Title: "URL",
				Value: url,
			},
		)
	}

	if elapsedTime > float64(0) {
		attachment.Fields = append(attachment.Fields,
			slack.AttachmentField{
				Title: "Elapsed time",
				Value: fmt.Sprintf("%.2f", elapsedTime),
			},
		)
	}

	_, _, err = slackConfig.client.PostMessage(slackConfig.channel, slack.MsgOptionText("shaker info", false), slack.MsgOptionAsUser(true), slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Errorf("can't send slack message: %s", err)
	}
}

func slackSendInfoMessage(slackConfig slackConfig, name string, text string, url string, elapsedTime float64) {
	slackSendMessage(slackConfig, name, text, "#008000", "Info", url, elapsedTime)
}

func slackSendWarningMessage(slackConfig slackConfig, name string, text string, url string, elapsedTime float64) {
	slackSendMessage(slackConfig, name, text, "#FFFF00", "Warning", url, elapsedTime)
}

func slackSendErrorMessage(slackConfig slackConfig, name string, text string, url string, elapsedTime float64) {
	slackSendMessage(slackConfig, name, text, "#ff0000", "Error", url, elapsedTime)
}
