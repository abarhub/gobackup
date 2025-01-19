package compress

import (
	"gobackup/internal/config"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"
	"time"
)

func Test_calculComplet(t *testing.T) {
	type args struct {
		repCompression string
		fichiers       []string
		backup         config.Backup
		global         config.BackupGlobal
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   time.Time
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{repCompression: "rep", fichiers: []string{
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001"}, backup: config.Backup{Nom: "doc1"}, global: config.BackupGlobal{DebugCompression: false, NbBackupIncremental: 5}},
			want: false, want1: date(2025, time.January, 18), wantErr: false},
		{name: "test2", args: args{repCompression: "rep", fichiers: []string{
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001"}, backup: config.Backup{Nom: "doc1"}, global: config.BackupGlobal{DebugCompression: false, NbBackupIncremental: 5}},
			want: true, want1: date(1, time.January, 1), wantErr: false},
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
			tt.args.repCompression = filepath.Join(tmpDir, tt.args.repCompression)
			got, got1, err := calculComplet(tt.args.repCompression, tt.args.backup, tt.args.global)
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

func Test_calculCompletFS(t *testing.T) {
	type args struct {
		repertoire          fs.FS
		nomBackup           string
		now                 time.Time
		debugCompression    bool
		nbBackupIncremental int
	}
	dateBackupComplet := date(1, time.January, 1)
	tests := []struct {
		name    string
		args    args
		want    bool
		want1   time.Time
		wantErr bool
	}{
		{name: "test1_incremental", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: false, want1: date(2025, time.January, 18), wantErr: false},
		{name: "test2_incremental", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
		{name: "test3_complet", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250113_083546596.7z.001",
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
		{name: "test4_incremental", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250113_083546596.7z.001",
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupc_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: false, want1: date(2025, time.January, 18), wantErr: false},
		{name: "test5_complet", args: args{repertoire: creerFichiers([]string{
			//"rep/backupc_doc1_20250112_083546596.7z.001",
			"rep/backupc_doc1_20250113_083546596.7z.001",
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
		{name: "test6_complet", args: args{repertoire: creerFichiers([]string{
			"rep/backupc_doc1_20250112_083546596.7z.001",
			"rep/backupi_doc1_20250113_083546596.7z.001",
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
		{name: "test7_complet", args: args{repertoire: creerFichiers([]string{"rep/"}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
		{name: "test8_complet", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250112_083546596.7z.001",
			"rep/backupi_doc1_20250113_083546596.7z.001",
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc2", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
		{name: "test9_complet", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc2", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
		{name: "test9_incremental", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250118_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.002",
			"rep/backupi_doc1_20250118_083546596.7z.003",
			"rep/backupi_doc1_20250118_083546596.7z.004",
			"rep/backupi_doc1_20250118_083546596.7z.005",
			"rep/backupi_doc1_20250118_083546596.7z.006",
			"rep/backupi_doc1_20250118_083546596.7z.007",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: false, want1: date(2025, time.January, 18), wantErr: false},
		{name: "test10_incremental", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.002",
			"rep/backupi_doc1_20250115_083546596.7z.003",
			"rep/backupi_doc1_20250115_083546596.7z.004",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: false, want1: date(2025, time.January, 18), wantErr: false},
		{name: "test11_complet", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.002",
			"rep/backupi_doc1_20250115_083546596.7z.003",
			"rep/backupi_doc1_20250115_083546596.7z.004",
			"rep/backupi_doc1_20250116_083546596.7z.001",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
		{name: "test12_incremental", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250115_083546596.7z.001.asc",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.002.asc",
			"rep/backupi_doc1_20250116_083546596.7z.003",
			"rep/backupi_doc1_20250117_083546596.7z.004.asc",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001.asc",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: false, want1: date(2025, time.January, 18), wantErr: false},
		{name: "test13_complet", args: args{repertoire: creerFichiers([]string{
			"rep/backupi_doc1_20250114_083546596.7z.001.asc",
			"rep/backupi_doc1_20250114_083546596.7z.001",
			"rep/backupi_doc1_20250115_083546596.7z.001.asc",
			"rep/backupi_doc1_20250115_083546596.7z.001",
			"rep/backupi_doc1_20250116_083546596.7z.002.asc",
			"rep/backupi_doc1_20250116_083546596.7z.003",
			"rep/backupi_doc1_20250117_083546596.7z.004.asc",
			"rep/backupi_doc1_20250117_083546596.7z.001",
			"rep/backupi_doc1_20250118_083546596.7z.001.asc",
			"rep/backupi_doc1_20250118_083546596.7z.001",
		}), nomBackup: "doc1", now: date(2025, time.January, 19),
			debugCompression: false, nbBackupIncremental: 5}, want: true, want1: dateBackupComplet, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := calculCompletFS(tt.args.repertoire, tt.args.nomBackup, tt.args.now, tt.args.debugCompression, tt.args.nbBackupIncremental, "rep")
			if (err != nil) != tt.wantErr {
				t.Errorf("calculCompletFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculCompletFS() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("calculCompletFS() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func creerFichiers(listeFichiers []string) fstest.MapFS {
	fichiers := fstest.MapFS{}
	for _, fichier := range listeFichiers {
		fichiers[fichier] = &fstest.MapFile{}
	}
	return fichiers
}

func date(annee int, mois time.Month, jour int) time.Time {
	return time.Date(annee, mois, jour, 0, 0, 0, 0, time.UTC)
}

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
