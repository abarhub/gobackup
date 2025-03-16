package crypt

import (
	"bytes"
	"errors"
	"gobackup/internal/compress"
	"gobackup/internal/config"
	"gobackup/internal/hashFiles"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCrypt(t *testing.T) {
	type args struct {
		fileCompressed compress.ResultatCompress
		b              config.Backup
		global         config.BackupGlobal
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
			if err := Crypt(tt.args.fileCompressed, tt.args.b, tt.args.global); (err != nil) != tt.wantErr {
				t.Errorf("Crypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCrypt2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping vss tests (no short test)")
	}
	if runtime.GOOS != "windows" {
		t.Skip("Skipping vss tests (not windows)")
	}
	//if os.Getenv("TEST_CRYPT") != "OK" {
	//	t.Skip("Skipping vss tests (not env TEST_CRYPT=OK)")
	//}
	repertoireAge := os.Getenv("TEST_CRYPT_AGE_REPERTOIRE")
	recipientAge := os.Getenv("TEST_CRYPT_AGE_RECIPIENT")
	if repertoireAge == "" {
		t.Errorf("La variable d'environnement TEST_CRYPT_AGE_REPERTOIRE n'est pas renseigné")
		return
	}
	if recipientAge == "" {
		t.Errorf("La variable d'environnement TEST_CRYPT_AGE_RECIPIENT n'est pas renseigné")
		return
	}
	type args struct {
		fileCompressed compress.ResultatCompress
		configBackup   config.Backup
		global         config.BackupGlobal
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "TestCrypt2",
			args: args{
				fileCompressed: compress.ResultatCompress{
					ListeFichier: []string{"backupc_document_20250312_075953349.7z.001",
						"backupc_document_20250312_075953349.7z.002"},
				},
				configBackup: config.Backup{
					Nom: "test1",
				},
				global: config.BackupGlobal{
					RepCryptage:  "",
					TypeCryptage: config.CryptAge,
					RepAge:       repertoireAge,
					AgeRecipien:  recipientAge,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repCompresse := t.TempDir()
			repCryptage := t.TempDir()
			tt.args.global.RepCryptage = repCryptage
			tt.args.global.RepCompression = repCompresse
			rep := filepath.Join(repCompresse, tt.args.configBackup.Nom)
			err := os.MkdirAll(rep, os.ModePerm)
			if err != nil {
				t.Error(err)
			}
			if !t.Failed() {
				buf := []byte{115, 111, 109, 101, 10}

				// création des fichiers
				for i := range tt.args.fileCompressed.ListeFichier {
					nom := tt.args.fileCompressed.ListeFichier[i]
					nomComplet := filepath.Join(rep, nom)
					d1 := buf
					err := os.WriteFile(nomComplet, d1, 0644)
					if err != nil {
						t.Errorf("Erreur pour créer le fichier %s : %v", nomComplet, err)
						break
					}
					tt.args.fileCompressed.ListeFichier[i] = nomComplet
				}

				if !t.Failed() {
					if err := Crypt(tt.args.fileCompressed, tt.args.configBackup, tt.args.global); (err != nil) != tt.wantErr {
						t.Errorf("Crypt() error = %v, wantErr %v", err, tt.wantErr)
					} else {
						for _, fichier := range tt.args.fileCompressed.ListeFichier {
							f := filepath.Join(repCryptage, tt.args.configBackup.Nom,
								filepath.Base(fichier)+extension2+hashFiles.GetExtension())
							if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
								t.Errorf("Crypt() error : le fichier %s n'existe pas", f)
							} else {
								data, err := os.ReadFile(f)
								if err != nil {
									t.Errorf("Crypt() error : erreur pour lire le fichier %s : %v", f, err)
								} else if len(data) == 0 {
									t.Errorf("Crypt() error : le fichier %s est vide", f)
								} else if bytes.Compare(data, buf) == 0 {
									t.Errorf("Crypt() error : le fichier %s est la même que la source", f)
								}
							}
						}
					}
				}
			}
		})
	}
}
