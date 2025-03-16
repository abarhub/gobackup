package compress

import (
	"testing"
)

func Test_calculHashFichiers(t *testing.T) {
	type args struct {
		repCompression string
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
			if err := calculHashFichiers(tt.args.repCompression); (err != nil) != tt.wantErr {
				t.Errorf("calculHashFichiers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
