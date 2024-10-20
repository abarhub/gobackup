package archivage

import (
	"fmt"
	"gobackup/internal/config"
	"gobackup/internal/utils"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestArchivage(t *testing.T) {
	type args struct {
		global config.BackupGlobal
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
			if err := Archivage(tt.args.global); (err != nil) != tt.wantErr {
				t.Errorf("Archivage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_archive(t *testing.T) {
	type args struct {
		nbJours             int
		listeFichierSrc     []string
		listeFichierReste   []string
		listeFichierArchive []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{nbJours: 3,
			listeFichierSrc: []string{"backupc_document_" + date(1) + "_214424065.7z.001",
				"backupc_document_" + date(2) + "_214424065.7z.001",
				"backupc_document_" + date(3) + "_214424065.7z.001",
				"backupc_document_" + date(4) + "_214424065.7z.001",
				"backupc_document_" + date(5) + "_214424065.7z.001"},
			listeFichierReste: []string{"backupc_document_" + date(1) + "_214424065.7z.001",
				"backupc_document_" + date(2) + "_214424065.7z.001",
				"backupc_document_" + date(3) + "_214424065.7z.001",
				"backupc_document_" + date(4) + "_214424065.7z.001"},
			listeFichierArchive: []string{"backupc_document_" + date(5) + "_214424065.7z.001"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep := t.TempDir()
			nom := "nom1"
			repSource := rep + "/src"
			err := os.MkdirAll(repSource+"/"+nom, os.ModePerm)
			if err != nil {
				t.Error(err)
			}
			repDestinaire := rep + "/dest"
			err = os.MkdirAll(repDestinaire+"/"+nom, os.ModePerm)
			if err != nil {
				t.Error(err)
			}
			for _, f := range tt.args.listeFichierSrc {
				err := utils.CreateEmptyFile(repSource + "/" + nom + "/" + f)
				if err != nil {
					t.Errorf("erreur pour créer le fichier %s : %v", f, err)
					break
				}
			}
			if err := archive(repSource, repDestinaire, tt.args.nbJours, false); (err != nil) != tt.wantErr {
				t.Errorf("archive() error = %v, wantErr %v", err, tt.wantErr)
			} else if err := compareRepertoire(repSource+"/"+nom, tt.args.listeFichierReste); err != nil {
				t.Errorf("archive() error = %v, wantErr %v", err, tt.wantErr)
			} else if err := compareRepertoire(repDestinaire+"/"+nom, tt.args.listeFichierArchive); err != nil {
				t.Errorf("archive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func compareRepertoire(repertoire string, listeFichiers []string) error {
	files, err := os.ReadDir(repertoire)
	if err != nil {
		return fmt.Errorf("erreur pour récupérer la liste des fichiers de %s : %v", repertoire, err)
	} else {
		liste := []string{}
		for _, f := range files {
			liste = append(liste, f.Name())
		}
		slices.Sort(liste)
		liste2 := []string{}
		for _, f := range listeFichiers {
			liste2 = append(liste2, f)
		}
		slices.Sort(liste2)
		if !reflect.DeepEqual(liste, liste2) {
			return fmt.Errorf("les listes sont différentes (%s) : %v != %v", filepath.Base(repertoire), liste, liste2)
		}
	}
	return nil
}

func date(decalage int) string {
	now := time.Now()
	date := now.AddDate(0, 0, -decalage)
	return date.Format("20060102")
}

func Test_moveFile(t *testing.T) {
	type args struct {
		source      string
		destination string
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
			if err := moveFile(tt.args.source, tt.args.destination); (err != nil) != tt.wantErr {
				t.Errorf("moveFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
