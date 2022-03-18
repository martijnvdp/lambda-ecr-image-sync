package handlers

import "testing"

func Test_sendSlackNotification(t *testing.T) {
	type args struct {
		slackOAuthToken string
		channelID       string
		subject         string
		message         string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "SendTestMessageToSlack",
			args: args{
				slackOAuthToken: "",
				channelID:       "C02JDHJhS",
				subject:         "test123",
				message:         "test from golang",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sendSlackNotification(tt.args.slackOAuthToken, tt.args.channelID, tt.args.subject, tt.args.message); (err != nil) != tt.wantErr {
				t.Errorf("sendSlackNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sendResultsToSlack(t *testing.T) {
	type args struct {
		csvContent     *[]csvFormat
		token          string
		channelID      string
		messageHeader  string
		messageSubject string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestSlackMessage",
			args: args{
				token:      "",          // oauth token of the slackbot with chat:public:write and chat:write access
				channelID:  "C02Pfgsaf", // right click on the channel and click copy link, id should be in there
				csvContent: &[]csvFormat{{"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/base/infra/datadoghq/agent", "v7.32.0"}, {"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/base/infra/datadoghq/agent", "v7.31.0"}, {"gcr.io/datadoghq/agent", "123321.dkr.ecr.eu-west-2.amazonaws.com/base/infra/datadoghq/agent", "v7.28.0"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sendResultsToSlack(tt.args.messageHeader, tt.args.messageSubject, tt.args.csvContent, tt.args.token, tt.args.channelID)
		})
	}
}
