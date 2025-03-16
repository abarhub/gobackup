package vss

import (
	"bytes"
	"gobackup/internal/config"
	"os"
	"runtime"
	"testing"
)

func TestInitVss(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping vss tests (no short test)")
	}
	if runtime.GOOS != "windows" {
		t.Skip("Skipping vss tests (not windows)")
	}
	if os.Getenv("TEST_VSS") != "OK" {
		t.Skip("Skipping vss tests (not env TEST_VSS=OK)")
	}

	rep := t.TempDir()
	nom := "nom1.txt"
	repSource := rep + "/src"
	err := os.MkdirAll(repSource, os.ModePerm)
	if err != nil {
		t.Error(err)
	} else {
		nomComplet := repSource + "/" + nom
		d1 := []byte{115, 111, 109, 101, 10}
		err = os.WriteFile(nomComplet, d1, 0644)
		if err != nil {
			t.Error(err)
		} else {

			var configGlobal *config.BackupGlobal
			configGlobal = &config.BackupGlobal{
				ListeBackup: []config.Backup{
					{
						Nom: "backup1",
						Rep: []string{repSource},
					},
				},
			}

			err = InitVss(configGlobal)
			if err != nil {
				t.Errorf("InitVss() error = %v", err)
			} else {
				t.Cleanup(func() {
					err := FermeVss(*configGlobal)
					if err != nil {
						t.Errorf("InitVss() error = %v", err)
					}
				})
				if len(configGlobal.LettreVss) != 1 {
					t.Errorf("InitVss() error : la liste des repertoires vss n'est pas 1 : %v",
						configGlobal.LettreVss)
				} else {
					lettreLecteur := repSource[0:1]
					rep := configGlobal.LettreVss[lettreLecteur]
					fichier := rep + "/" + nom
					data, err := os.ReadFile(fichier)
					if err != nil {
						t.Errorf("InitVss() erreur pour lire le fichier %v = %v", fichier, err)
					} else {
						if bytes.Compare(data, d1) != 0 {
							t.Error("InitVss() erreur : le contenu des fichiers est différents")
						}
					}
				}
			}
		}
	}
}
