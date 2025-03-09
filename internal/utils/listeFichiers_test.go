package utils

import (
	"gobackup/internal/config"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var listeFichiers = []string{
	"rep/backupi_doc1_20250115_083546596.7z.001",
	"rep/backupi_doc1_20250116_083546596.7z.001",
	"rep/backupi_doc1_20250117_083546596.7z.001",
	"rep/backupi_doc1_20250118_083546596.7z.001"}

var listeFichiers2 = []string{
	"rep/backupc_doc1_20250115_083546596.7z.001"}

var listeFichiers3 = []string{
	"rep/backupc_doc1_20250115_083546596.7z.001",
	"rep/backupc_doc1_20250116_083546596.7z.001",
	"rep/backupc_doc1_20250117_083546596.7z.001",
	"rep/backupc_doc1_20250118_083546596.7z.001"}

var listeFichiers4 = []string{
	"rep/backupc_doc1_20250115_083546596.7z.001",
	"rep/backupi_doc1_20250116_083546596.7z.001",
	"rep/backupi_doc1_20250117_083546596.7z.001",
	"rep/backupi_doc1_20250118_083546596.7z.001"}

func TestCalcul(t *testing.T) {
	type args struct {
		fichiers   []string
		nbIncement int
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 time.Time
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{fichiers: listeFichiers, nbIncement: 2}, want: true},
		{name: "test2", args: args{fichiers: listeFichiers, nbIncement: 6}, want: false, want1: time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC)},
		{name: "test3", args: args{fichiers: listeFichiers, nbIncement: 5}, want: false, want1: time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC)},
		{name: "test4", args: args{fichiers: listeFichiers, nbIncement: 4}, want: true},
		{name: "test5", args: args{fichiers: listeFichiers, nbIncement: 3}, want: true},
		{name: "test6", args: args{fichiers: listeFichiers, nbIncement: 0}, want: true},
		{name: "test7", args: args{fichiers: listeFichiers2, nbIncement: 3}, want: false, want1: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{name: "test8", args: args{fichiers: listeFichiers3, nbIncement: 3}, want: false, want1: time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC)},
		{name: "test9", args: args{fichiers: listeFichiers3, nbIncement: 0}, want: true},
		{name: "test10", args: args{fichiers: listeFichiers4, nbIncement: 2}, want: true},
		{name: "test11", args: args{fichiers: listeFichiers4, nbIncement: 3}, want: true},
		{name: "test12", args: args{fichiers: listeFichiers4, nbIncement: 4}, want: false, want1: time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for _, file := range tt.args.fichiers {
				path := filepath.Dir(filepath.Join(tmpDir, file))
				if path != tmpDir {
					t.Logf("tmpDir:%s, path:%s", tmpDir, path)
					err := os.MkdirAll(path, 0777)
					if err != nil {
						t.Fatal(err)
					}
				}
				err := os.WriteFile(filepath.Join(tmpDir, file), []byte(""), os.FileMode(0755))
				if err != nil {
					t.Fatal(err)
				}
			}
			backup := config.Backup{
				Nom: "doc1",
			}
			global := config.BackupGlobal{
				DebugCompression:    false,
				NbBackupIncremental: tt.args.nbIncement,
			}
			rep := filepath.Join(tmpDir, "rep")
			got, got1, err := Calcul(rep, backup, global)
			if err != nil {
				t.Errorf("Calcul() error = %v, wantErr %v", err, false)
				return
			}
			if got != tt.want {
				t.Errorf("Calcul() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Calcul() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestCreateListeFichier(t *testing.T) {
	type args struct {
		nomBackup string
	}
	tests := []struct {
		name string
		args args
		want *ListeFichiers
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{nomBackup: "nom1"}, want: &ListeFichiers{listeFichiers: []Fichiers{}, nombackup: "nom1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateListeFichier(tt.args.nomBackup); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateListeFichier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListeFichiers_ajouteFichier(t *testing.T) {
	type fields struct {
		listeFichiers []Fichiers
		nombackup     string
	}
	type args struct {
		nomFichier string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Fichiers
		wantErr bool
	}{
		// TODO: Add test cases.
		{"test1_complet", fields{listeFichiers: []Fichiers{}, nombackup: "doc1"},
			args{nomFichier: "backupc_doc1_20250115_083546596.7z.001"},
			[]Fichiers{Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001"},
				true}}, false},
		{"test2_extention_gpg", fields{listeFichiers: []Fichiers{}, nombackup: "doc1"},
			args{nomFichier: "backupc_doc1_20250115_083546596.7z.001.gpg"},
			[]Fichiers{Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001.gpg"},
				true}}, false},
		{"test3_extention_age", fields{listeFichiers: []Fichiers{}, nombackup: "doc1"},
			args{nomFichier: "backupc_doc1_20250115_083546596.7z.001.age"},
			[]Fichiers{Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001.age"},
				true}}, false},
		{"test4_autre_date", fields{listeFichiers: []Fichiers{}, nombackup: "doc1"},
			args{nomFichier: "backupc_doc1_20210307_164105496.7z.001"},
			[]Fichiers{Fichiers{"20210307_164105.496",
				time.Date(2021, 3, 07, 16, 41, 05, 496000000, time.UTC),
				"backupc_doc1_20210307_164105496", []string{"backupc_doc1_20210307_164105496.7z.001"},
				true}}, false},
		{"test5_nom_different", fields{listeFichiers: []Fichiers{}, nombackup: "doc1"},
			args{nomFichier: "backupc_doc2_20250115_083546596.7z.001"},
			[]Fichiers{}, false},
		{"test6_nom_sans_rapport", fields{listeFichiers: []Fichiers{}, nombackup: "doc1"},
			args{nomFichier: "aaa.txt"},
			[]Fichiers{}, false},
		{"test7_fichier_incomplet", fields{listeFichiers: []Fichiers{}, nombackup: "doc1"},
			args{nomFichier: "backupi_doc1_20250115_083546596.7z.001"},
			[]Fichiers{Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupi_doc1_20250115_083546596", []string{"backupi_doc1_20250115_083546596.7z.001"},
				false}}, false},
		{"test8_date_invalide", fields{listeFichiers: []Fichiers{}, nombackup: "doc1"},
			args{nomFichier: "backupc_doc1_20252015_083546596.7z.001"},
			[]Fichiers{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			liste := &ListeFichiers{
				listeFichiers: tt.fields.listeFichiers,
				nombackup:     tt.fields.nombackup,
			}
			if err := liste.ajouteFichier(tt.args.nomFichier); (err != nil) != tt.wantErr {
				t.Errorf("ajouteFichier() error = %v, wantErr %v", err, tt.wantErr)
			} else if !reflect.DeepEqual(liste.listeFichiers, tt.want) {
				t.Errorf("ajouteFichier() = %v, want %v", liste.listeFichiers, tt.want)
			}
		})
	}
}

func TestListeFichiers_ajouteRepertoire(t *testing.T) {
	type fields struct {
		listeFichiers []Fichiers
		nombackup     string
	}
	type args struct {
		repertoire string
		fichiers   []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Fichiers
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", fields: fields{listeFichiers: []Fichiers{}}, args: args{repertoire: "doc1", fichiers: listeFichiers}, want: []Fichiers{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for _, file := range tt.args.fichiers {
				path := filepath.Dir(filepath.Join(tmpDir, file))
				if path != tmpDir {
					t.Logf("tmpDir:%s, path:%s", tmpDir, path)
					err := os.MkdirAll(path, 0777)
					if err != nil {
						t.Fatal(err)
					}
				}
				err := os.WriteFile(filepath.Join(tmpDir, file), []byte(""), os.FileMode(0755))
				if err != nil {
					t.Fatal(err)
				}
			}
			liste := &ListeFichiers{
				listeFichiers: tt.fields.listeFichiers,
				nombackup:     tt.fields.nombackup,
			}
			rep := filepath.Join(tmpDir, "rep")
			if err := liste.ajouteRepertoire(rep); (err != nil) != tt.wantErr {
				t.Errorf("ajouteRepertoire() error = %v, wantErr %v", err, tt.wantErr)
			} else if !reflect.DeepEqual(liste.listeFichiers, tt.want) {
				t.Errorf("ajouteRepertoire() = %v, want %v", liste.listeFichiers, tt.want)
			}
		})
	}
}

func TestListeFichiers_calculComplet(t *testing.T) {
	type fields struct {
		listeFichiers []Fichiers
		nombackup     string
	}
	type args struct {
		nbBackupIncremental2 int
		now                  time.Time
		debugCompression     bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		want1   time.Time
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1_incremental", fields: fields{listeFichiers: []Fichiers{
			Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupi_doc1_20250115_083546596", []string{"backupi_doc1_20250115_083546596.7z.001"},
				false},
			Fichiers{"20250116_083546.596",
				time.Date(2025, 1, 16, 8, 35, 46, 596000000, time.UTC),
				"backupi_doc1_20250116_083546596", []string{"backupi_doc1_20250116_083546596.7z.001"},
				false},
			Fichiers{"20250117_083546.596",
				time.Date(2025, 1, 17, 8, 35, 46, 596000000, time.UTC),
				"backupi_doc1_20250117_083546596", []string{"backupi_doc1_20250117_083546596.7z.001"},
				false},
		}}, args: args{nbBackupIncremental2: 5, now: time.Date(2025, 1, 17, 8, 35, 46, 596000000, time.UTC), debugCompression: true},
			want: false, want1: time.Date(2025, 1, 17, 0, 0, 0, 0, time.UTC), wantErr: false},
		{name: "test2_complet", fields: fields{listeFichiers: []Fichiers{
			Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupi_doc1_20250115_083546596", []string{"backupi_doc1_20250115_083546596.7z.001"},
				false},
			Fichiers{"20250116_083546.596",
				time.Date(2025, 1, 16, 8, 35, 46, 596000000, time.UTC),
				"backupi_doc1_20250116_083546596", []string{"backupi_doc1_20250116_083546596.7z.001"},
				false},
			Fichiers{"20250117_083546.596",
				time.Date(2025, 1, 17, 8, 35, 46, 596000000, time.UTC),
				"backupi_doc1_20250117_083546596", []string{"backupi_doc1_20250117_083546596.7z.001"},
				false},
		}}, args: args{nbBackupIncremental2: 2, now: time.Date(2025, 1, 17, 8, 35, 46, 596000000, time.UTC), debugCompression: true},
			want: true, want1: time.Time{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			liste := &ListeFichiers{
				listeFichiers: tt.fields.listeFichiers,
				nombackup:     tt.fields.nombackup,
			}
			got, got1, err := liste.calculComplet(tt.args.nbBackupIncremental2, tt.args.now, tt.args.debugCompression)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculComplet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculComplet() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("calculComplet() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestListeFichiers_trie(t *testing.T) {
	type fields struct {
		listeFichiers []Fichiers
	}
	tests := []struct {
		name   string
		fields fields
		want   []Fichiers
	}{
		// TODO: Add test cases.
		{name: "test1", fields: fields{listeFichiers: []Fichiers{}}, want: []Fichiers{}},
		{name: "test2", fields: fields{listeFichiers: []Fichiers{
			Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001.gpg"},
				true},
			Fichiers{"20250116_083546.596",
				time.Date(2025, 1, 16, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250116_083546596", []string{"backupc_doc1_20250116_083546596.7z.001.gpg"},
				true},
		}}, want: []Fichiers{
			Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001.gpg"},
				true},
			Fichiers{"20250116_083546.596",
				time.Date(2025, 1, 16, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250116_083546596", []string{"backupc_doc1_20250116_083546596.7z.001.gpg"},
				true}}},
		{name: "test3", fields: fields{listeFichiers: []Fichiers{
			Fichiers{"20250116_083546.596",
				time.Date(2025, 1, 16, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250116_083546596", []string{"backupc_doc1_20250116_083546596.7z.001.gpg"},
				true},
			Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001.gpg"},
				true},
		}}, want: []Fichiers{
			Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001.gpg"},
				true},
			Fichiers{"20250116_083546.596",
				time.Date(2025, 1, 16, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250116_083546596", []string{"backupc_doc1_20250116_083546596.7z.001.gpg"},
				true}}},
		{name: "test4", fields: fields{listeFichiers: []Fichiers{
			Fichiers{"20250116_083546.596",
				time.Date(2025, 1, 16, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250116_083546596", []string{"backupc_doc1_20250116_083546596.7z.001.gpg"},
				true},
			Fichiers{"20230110_083546.596",
				time.Date(2023, 1, 10, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20230110_083546596", []string{"backupc_doc1_20230110_083546596.7z.001.gpg"},
				true},
			Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001.gpg"},
				true},
		}}, want: []Fichiers{
			Fichiers{"20230110_083546.596",
				time.Date(2023, 1, 10, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20230110_083546596", []string{"backupc_doc1_20230110_083546596.7z.001.gpg"},
				true},
			Fichiers{"20250115_083546.596",
				time.Date(2025, 1, 15, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250115_083546596", []string{"backupc_doc1_20250115_083546596.7z.001.gpg"},
				true},
			Fichiers{"20250116_083546.596",
				time.Date(2025, 1, 16, 8, 35, 46, 596000000, time.UTC),
				"backupc_doc1_20250116_083546596", []string{"backupc_doc1_20250116_083546596.7z.001.gpg"},
				true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			liste := &ListeFichiers{
				listeFichiers: tt.fields.listeFichiers,
				nombackup:     "nom1",
			}
			liste.trie()
			if !reflect.DeepEqual(liste.listeFichiers, tt.want) {
				t.Errorf("trie() listeFichiers = %v, want %v", liste.listeFichiers, tt.want)
			}
		})
	}
}
