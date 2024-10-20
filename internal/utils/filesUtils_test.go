package utils

import (
	"errors"
	"os"
	"testing"
)

func TestCreateEmptyFile(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{filename: t.TempDir() + "test1.txt"}, wantErr: false},
		{name: "test2", args: args{filename: t.TempDir() + "test1.txt"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateEmptyFile(tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("CreateEmptyFile() error = %v, wantErr %v", err, tt.wantErr)
			} else if _, err := os.Stat(tt.args.filename); errors.Is(err, os.ErrNotExist) {
				t.Errorf("CreateEmptyFile() le fichier %s n'existe pas", tt.args.filename)
			} else {
				b, err := os.ReadFile(tt.args.filename)
				if err != nil {
					t.Errorf("error for read file %s : %v", tt.args.filename, err)
				} else {
					str := string(b)
					if str != "" {
						t.Errorf("CreateEmptyFile() le fichier %s n'est pas vide (%s)", tt.args.filename, str)
					}
				}
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	type args struct {
		filename string
		contenu  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1", args: args{filename: t.TempDir() + "test1.txt", contenu: "XXX"}, wantErr: false},
		{name: "test2", args: args{filename: t.TempDir() + "test1.txt", contenu: "abc123"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateFile(tt.args.filename, tt.args.contenu); (err != nil) != tt.wantErr {
				t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
			} else if _, err := os.Stat(tt.args.filename); errors.Is(err, os.ErrNotExist) {
				t.Errorf("CreateEmptyFile() le fichier %s n'existe pas", tt.args.filename)
			} else {
				b, err := os.ReadFile(tt.args.filename)
				if err != nil {
					t.Errorf("error for read file %s : %v", tt.args.filename, err)
				} else {
					str := string(b)
					if str != tt.args.contenu {
						t.Errorf("CreateEmptyFile() le fichier %s n'a pas le bon contenu (%s!=%s)", tt.args.filename, str, tt.args.contenu)
					}
				}
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "test1", args: args{filename: t.TempDir() + "test1.txt"}, wantErr: false, want: "XXX"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateFile(tt.args.filename, tt.want)
			if err != nil {
				t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				got, err := ReadFile(tt.args.filename)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("ReadFile() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
