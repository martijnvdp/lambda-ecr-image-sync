package handlers

import "testing"

func Test_createZipFile(t *testing.T) {
	type args struct {
		file   string
		target string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createZipFile(tt.args.file, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("createZipFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
