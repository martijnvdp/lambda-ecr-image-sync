package lambda

import (
	"context"
	"time"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/slack"
)

func sendSlackNotification(slackToken, channelID string, subject, message string) (err error) {
	if slackToken == "" {
		return err
	}
	notifier := notify.New()
	slackService := slack.New(slackToken)
	slackService.AddReceivers(channelID)
	notifier.UseServices(slackService)

	err = notifier.Send(
		context.Background(),
		subject,
		message,
	)
	if err != nil {
		return err
	}
	return err
}

func sendResultsToSlack(messageHeader, messageSubject string, token, channelID string) {
	if token == "" {
		return
	}
	subject := tryString(messageSubject, "Lambda ECR-IMAGE-SYNC has run.") + "\n"
	message := tryString(messageHeader, "The following ecr images are being Synced to ECR:") + "\n"

	currentTime := time.Now()
	message = message + "\n" + currentTime.Format("2006-01-02 15:04:05")
	sendSlackNotification(token, channelID, subject, message)
}
