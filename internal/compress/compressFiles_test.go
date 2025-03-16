package compress

import (
	"errors"
	"gobackup/internal/config"
	"gobackup/internal/hashFiles"
	"os"
	"path/filepath"
	"runtime"
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

func TestCompress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping vss tests (no short test)")
	}
	if runtime.GOOS != "windows" {
		t.Skip("Skipping vss tests (not windows)")
	}
	if os.Getenv("TEST_COMPRESS") != "OK" {
		t.Skip("Skipping vss tests (not env TEST_COMPRESS=OK)")
	}
	repertoire7zip := os.Getenv("TEST_COMPRESS_7Z")
	if repertoire7zip == "" {
		t.Errorf("La variable d'environnement TEST_COMPRESS_7Z n'est pas renseigné")
		return
	}
	type args struct {
		backup  config.Backup
		global  config.BackupGlobal
		fichier bool
	}
	tests := []struct {
		name    string
		args    args
		want    ResultatCompress
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				backup: config.Backup{
					Nom: "test1",
				},
				global: config.BackupGlobal{
					RepCompression:      "",
					NbBackupIncremental: 5,
					DebugCompression:    true,
					Rep7zip:             repertoire7zip,
					DateHeure:           "20230410",
				},
				fichier: true,
			},
			want: ResultatCompress{
				ListeFichier: []string{
					"backupi_test1_20230410.7z.001",
				},
			},
			wantErr: false,
		},
		{
			name: "test2_aucun_fichier",
			args: args{
				backup: config.Backup{
					Nom: "test1",
				},
				global: config.BackupGlobal{
					RepCompression:      "",
					NbBackupIncremental: 5,
					DebugCompression:    true,
					Rep7zip:             repertoire7zip,
					DateHeure:           "20230410",
				},
				fichier: false,
			},
			want: ResultatCompress{
				ListeFichier: []string{
					"backupi_test1_20230410.7z",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := t.TempDir()
			rep := filepath.Join(res, tt.args.backup.Nom)
			err := os.MkdirAll(rep, os.ModePerm)
			if err != nil {
				t.Error(err)
			}
			rep2 := t.TempDir()
			repCompression := filepath.Join(rep2, tt.args.backup.Nom)
			err = os.MkdirAll(repCompression, os.ModePerm)
			if err != nil {
				t.Error(err)
			}
			tt.args.global.RepCompression = rep2
			tt.args.backup.Rep = []string{rep}
			// fichier contenant la liste des fichiers à compresser
			tt.args.backup.FileListe = filepath.Join(t.TempDir(), "listeFichiers_"+tt.args.backup.Nom+".txt")
			if tt.args.fichier {
				// fichier à compresser
				buf := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
				fichier := filepath.Join(rep, "fichier1.txt")
				err = os.WriteFile(fichier, buf, 0644)
				if err != nil {
					t.Errorf("Erreur pour créer le fichier %s : %v", fichier, err)
				}
			}

			got, err := Compress(tt.args.backup, tt.args.global)
			if (err != nil) != tt.wantErr {
				t.Errorf("Compress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, s := range tt.want.ListeFichier {
				nomFichier := filepath.Join(repCompression, s)
				trouve := false
				for _, s2 := range got.ListeFichier {
					if s2 == nomFichier {
						trouve = true
						break
					}
				}
				if !trouve {
					t.Errorf("impossible de trouve : %s", nomFichier)
				} else {
					t.Logf("trouve : %s", nomFichier)
					if _, err := os.Stat(nomFichier); errors.Is(err, os.ErrNotExist) {
						t.Errorf("Compress() error : le fichier %s n'existe pas", nomFichier)
					}
					nomFichierHash := nomFichier + hashFiles.GetExtension()
					if _, err := os.Stat(nomFichierHash); errors.Is(err, os.ErrNotExist) {
						t.Errorf("Compress() error : le fichier %s n'existe pas", nomFichierHash)
					}
				}
			}
		})
	}
}
