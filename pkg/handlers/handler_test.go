package handlers

import (
	"context"
	"reflect"
	"testing"
)

func Test_getEnvironmentVars(t *testing.T) {
	tests := []struct {
		name     string
		wantVars environmentVars
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVars, err := getEnvironmentVars()
			if (err != nil) != tt.wantErr {
				t.Errorf("getEnvironmentVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotVars, tt.wantVars) {
				t.Errorf("getEnvironmentVars() = %v, want %v", gotVars, tt.wantVars)
			}
		})
	}
}

func Test_getEcrImageName(t *testing.T) {
	type args struct {
		imageName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestGetECRImageName-FullName",
			args: args{
				imageName: "docker.io/martijnvdp/ecr-image-sync",
			},
			want: "martijnvdp/ecr-image-sync",
		},
		{
			name: "TestGetECRImageName-NameOnly",
			args: args{
				imageName: "martijnvdp/ecr-image-sync",
			},
			want: "martijnvdp/ecr-image-sync",
		},
		{
			name: "TestGetECRImageName-RootImage",
			args: args{
				imageName: "docker.io/nginx",
			},
			want: "nginx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEcrImageName(tt.args.imageName); got != tt.want {
				t.Errorf("getEcrImageName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_maxInt(t *testing.T) {
	type args struct {
		x int
		y int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maxInt(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("maxInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_returnErr(t *testing.T) {
	type args struct {
		err              error
		slackOAuthTokenm string
		slackChannelID   string
		errSubject       string
		errText          string
	}
	tests := []struct {
		name    string
		args    args
		want    response
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := returnErr(tt.args.err, tt.args.slackOAuthTokenm, tt.args.slackChannelID, tt.args.errSubject, tt.args.errText)
			if (err != nil) != tt.wantErr {
				t.Errorf("returnErr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("returnErr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLambda(t *testing.T) {
	type args struct {
		ctx         context.Context
		lambdaEvent LambdaEvent
	}
	tests := []struct {
		name    string
		args    args
		want    response
		wantErr bool
	}{
		{
			name: "TestNoEnvironmetVarsSet",
			args: args{
				lambdaEvent: LambdaEvent{
					SlackChannelID: "C0123455",
				},
			},
			want: response{
				Message: "Error reading environment variables , or not set: &{%!f(string=error no environment variables set)}",
				Ok:      false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Lambda(tt.args.ctx, tt.args.lambdaEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lambda() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lambda() = %v, want %v", got, tt.want)
			}
		})
	}
}
