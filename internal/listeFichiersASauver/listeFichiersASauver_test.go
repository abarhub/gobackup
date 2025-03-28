package listeFichiersASauver

import (
	"bufio"
	"fmt"
	"gobackup/internal/config"
	"gobackup/internal/utils"
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

type args struct {
	backup  config.Backup
	complet bool
	date    time.Time
	global  config.BackupGlobal
	repTemp string
}

func TestListeFiles(t *testing.T) {
	tests := []struct {
		name    string
		args    args
		want    ListeFichiersASauver
		want2   []string
		wantErr bool
	}{
		{name: "test1", args: args{backup: config.Backup{Nom: "test1", FileListe: t.TempDir() + "/listeFichier.txt", Rep: []string{"rep"}},
			complet: true, global: config.BackupGlobal{}, repTemp: t.TempDir()}, want: ListeFichiersASauver{NbFiles: 1}, want2: []string{"rep/fichier1.txt"}},
		{name: "test2", args: args{backup: config.Backup{Nom: "test2", FileListe: t.TempDir() + "/listeFichier.txt", Rep: []string{"rep2"}},
			complet: true, global: config.BackupGlobal{}, repTemp: t.TempDir()}, want: ListeFichiersASauver{NbFiles: 6},
			want2: []string{"rep2/fichier1.txt", "rep2/fichier2.csv", "rep2/fichier3.csv", "rep2/test1/fichier01.txt",
				"rep2/test1/fichier02.txt", "rep2/test1/fichier03.doc"}},
		{name: "test3", args: args{backup: config.Backup{Nom: "test2", FileListe: t.TempDir() + "/listeFichier.txt", Rep: []string{"rep", "rep2"}},
			complet: true, global: config.BackupGlobal{}, repTemp: t.TempDir()}, want: ListeFichiersASauver{NbFiles: 7},
			want2: []string{"rep/fichier1.txt", "rep2/fichier1.txt", "rep2/fichier2.csv", "rep2/fichier3.csv", "rep2/test1/fichier01.txt",
				"rep2/test1/fichier02.txt", "rep2/test1/fichier03.doc"}},
		{name: "test4", args: args{backup: config.Backup{Nom: "test2", FileListe: t.TempDir() + "/listeFichier.txt", Rep: []string{"rep2"},
			Exclusion: config.ExclusionType{Set: map[string]bool{"test1": true}}},
			complet: true, global: config.BackupGlobal{}, repTemp: t.TempDir()}, want: ListeFichiersASauver{NbFiles: 3},
			want2: []string{"rep2/fichier1.txt", "rep2/fichier2.csv", "rep2/fichier3.csv"}},
		{name: "test5", args: args{backup: config.Backup{Nom: "test2", FileListe: t.TempDir() + "/listeFichier.txt",
			Rep: []string{"rep2"}, Exclusion: config.ExclusionType{Map2: map[string][][]string{"test1": {{"rep2", "test1"}}}}},
			complet: true, global: config.BackupGlobal{}, repTemp: t.TempDir()}, want: ListeFichiersASauver{NbFiles: 3},
			want2: []string{"rep2/fichier1.txt", "rep2/fichier2.csv", "rep2/fichier3.csv"}},
		{name: "test6", args: args{backup: config.Backup{Nom: "test2", FileListe: t.TempDir() + "/listeFichier.txt",
			Rep: []string{"rep3"}, Exclusion: config.ExclusionType{Map2: map[string][][]string{"test04": {{"rep3", "test03", "test04"}}}}},
			complet: true, global: config.BackupGlobal{}, repTemp: t.TempDir()}, want: ListeFichiersASauver{NbFiles: 4},
			want2: []string{"rep3/fichier1.txt", "rep3/test01/fichier2.txt", "rep3/test01/test02/fichier3.txt", "rep3/test03/fichier4.txt"}},
		{name: "test7", args: args{backup: config.Backup{Nom: "test2", FileListe: t.TempDir() + "/listeFichier.txt",
			Rep: []string{"rep4"}, Exclusion: config.ExclusionType{Map2: map[string][][]string{"test02": {{"rep4", "test03", "test02"}}}}},
			complet: true, global: config.BackupGlobal{}, repTemp: t.TempDir()}, want: ListeFichiersASauver{NbFiles: 4},
			want2: []string{"rep4/fichier1.txt", "rep4/test01/fichier2.txt", "rep4/test01/test02/fichier3.txt", "rep4/test03/fichier4.txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err2 := initialiseRepertoire(t, tt.args.backup, tt.args, &tt.want)
			if err2 != nil {
				t.Errorf("ParcourtListeFichiersASauver() erreur pour initialiser les fichiers = %v", err2)
				return
			}
			got, err := ParcourtListeFichiersASauver(tt.args.backup, tt.args.complet, tt.args.date, tt.args.global)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParcourtListeFichiersASauver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParcourtListeFichiersASauver() got = %v, want %v", got, tt.want)
			}
			resultat, err := lectureFichier(tt.want.ListeFiles, tt.args.repTemp)
			if err != nil {
				t.Errorf("ParcourtListeFichiersASauver() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(resultat, tt.want2) {
				t.Errorf("ParcourtListeFichiersASauver() got = %v, want %v", resultat, tt.want2)
			}
		})
	}
}

func lectureFichier(files string, tempDir string) ([]string, error) {
	file, err := os.Open(files)
	if err != nil {
		log.Fatalf("Erreur lors de l'ouverture du fichier: %s", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	var lignes []string

	// Crée un scanner pour lire le fichier ligne par ligne
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Ajoute chaque ligne lue au slice
		ligne := scanner.Text()
		ligne, found := strings.CutPrefix(ligne, tempDir)
		if !found {
			return []string{}, fmt.Errorf("la ligne ne commence pas par %s", tempDir)
		}
		ligne, _ = strings.CutPrefix(ligne, "/")
		ligne, _ = strings.CutPrefix(ligne, "\\")
		ligne = strings.Replace(ligne, "\\", "/", -1)
		lignes = append(lignes, ligne)
	}

	if err := scanner.Err(); err != nil {
		return []string{}, fmt.Errorf("Erreur lors de la lecture du fichier %s : %v", files, err)
	}

	sort.Strings(lignes)

	return lignes, nil
}

func initialiseRepertoire(t *testing.T, backup config.Backup, arguments args, wants *ListeFichiersASauver) error {
	repertoireTemporaire := arguments.repTemp
	liste := []string{
		// rep
		"rep/fichier1.txt",
		// rep2
		"rep2/fichier1.txt", "rep2/fichier2.csv", "rep2/fichier3.csv",
		"rep2/test1/fichier01.txt", "rep2/test1/fichier02.txt", "rep2/test1/fichier03.doc",
		// rep3
		"rep3/fichier1.txt", "rep3/test01/fichier2.txt", "rep3/test01/test02/fichier3.txt",
		"rep3/test03/fichier4.txt", "rep3/test03/test04/fichier5.txt",
		// rep4
		"rep4/fichier1.txt", "rep4/test01/fichier2.txt", "rep4/test01/test02/fichier3.txt",
		"rep4/test03/fichier4.txt", "rep4/test03/test02/fichier5.txt",
	}

	for _, fichier := range liste {
		f := repertoireTemporaire + "/" + fichier
		path.Dir(f)
		err := os.MkdirAll(path.Dir(f), 0777)
		if err != nil {
			return err
		}
		err = utils.CreateEmptyFile(f)
		if err != nil {
			return err
		}
	}

	for i := range backup.Rep {
		rep := backup.Rep[i]
		if rep != "" {
			backup.Rep[i] = repertoireTemporaire + "/" + rep
		}
	}
	wants.ListeFiles = arguments.backup.FileListe
	return nil
}

func Test_convertie(t *testing.T) {
	type args struct {
		root   string
		global config.BackupGlobal
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{root: "C:\\test1\\test2", global: config.BackupGlobal{LettreVss: map[string]string{"C": "d:\\link"}}}, want: "d:\\link\\test1\\test2"},
		{name: "test1", args: args{root: "c:\\test1\\test2", global: config.BackupGlobal{LettreVss: map[string]string{"C": "d:\\link"}}}, want: "d:\\link\\test1\\test2"},
		{name: "test1", args: args{root: "c:\\test1\\test2", global: config.BackupGlobal{LettreVss: map[string]string{"F": "d:\\link"}}}, want: "c:\\test1\\test2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertie(tt.args.root, tt.args.global); got != tt.want {
				t.Errorf("convertie() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parcourt(t *testing.T) {
	type args struct {
		res          config.Backup
		complet      bool
		date         time.Time
		configGlobal config.BackupGlobal
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parcourt(tt.args.res, tt.args.complet, tt.args.date, tt.args.configGlobal)
			if (err != nil) != tt.wantErr {
				t.Errorf("parcourt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parcourt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_testEqSuffixSlice(t *testing.T) {
	type args struct {
		suffix []string
		tab    []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test1", args: args{suffix: []string{"mon", "rep"}, tab: []string{"toto", "mon", "rep"}}, want: true},
		{name: "test2", args: args{suffix: []string{"mon", "rep"}, tab: []string{"mon", "rep"}}, want: true},
		{name: "test3", args: args{suffix: []string{"mon", "rep"}, tab: []string{"rep"}}, want: false},
		{name: "test4", args: args{suffix: []string{"mon", "rep"}, tab: []string{"mon", "rep2"}}, want: false},
		{name: "test5", args: args{suffix: []string{"mon", "rep"}, tab: []string{"mon2", "rep"}}, want: false},
		{name: "test6", args: args{suffix: []string{"mon", "rep"}, tab: []string{"mon3", "rep3"}}, want: false},
		{name: "test7", args: args{suffix: []string{"aaa", "mon", "rep"}, tab: []string{"mon", "rep"}}, want: false},
		{name: "test8", args: args{suffix: []string{"rep"}, tab: []string{"rep"}}, want: true},
		{name: "test8", args: args{suffix: []string{"rep"}, tab: []string{"rep2"}}, want: false},
		{name: "test8", args: args{suffix: []string{"rep", "test1", "myrep"}, tab: []string{"rep_root", "rep", "test1", "myrep"}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testEqSuffixSlice(tt.args.suffix, tt.args.tab); got != tt.want {
				t.Errorf("testEqSuffixSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExclusion(t *testing.T) {
	type args struct {
		path      string
		exclusion config.ExclusionType
		dir       bool
	}
	tests := []struct {
		name string
		args args
		want IgnoreParcourt
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{path: "rep/test1", exclusion: config.ExclusionType{Set: map[string]bool{"test1": true}}, dir: true}, want: IgnoreRepertoire},
		{name: "test2", args: args{path: "rep/test1/fichier1.txt", exclusion: config.ExclusionType{Set: map[string]bool{"fichier1.txt": true}}, dir: false}, want: IgnoreFichier},
		{name: "test3", args: args{path: "rep/test1/fichier1.txt", exclusion: config.ExclusionType{Set: map[string]bool{"fichier2.txt": true}}, dir: false}, want: Continue},
		{name: "test4", args: args{path: "rep/test1", exclusion: config.ExclusionType{Set: map[string]bool{"test2": true}}, dir: true}, want: Continue},
		{name: "test5", args: args{path: "rep/test1/rep01", exclusion: config.ExclusionType{Map2: map[string][][]string{"rep01": {{"test1", "rep01"}}}}, dir: true}, want: IgnoreRepertoire},
		{name: "test6", args: args{path: "rep/test1/rep01", exclusion: config.ExclusionType{Map2: map[string][][]string{"rep01": {{"rep", "test1", "rep01"}}}}, dir: true}, want: IgnoreRepertoire},
		{name: "test7", args: args{path: "rep/test1/rep01", exclusion: config.ExclusionType{Map2: map[string][][]string{"rep01": {{"rep01"}}}}, dir: true}, want: IgnoreRepertoire},
		{name: "test8", args: args{path: "rep/test1/rep02", exclusion: config.ExclusionType{Map2: map[string][][]string{"rep01": {{"rep", "test1", "rep01"}}}}, dir: true}, want: Continue},
		{name: "test9", args: args{path: "rep/test1/rep01", exclusion: config.ExclusionType{Map2: map[string][][]string{"rep01": {{"test1", "rep02"}}}}, dir: true}, want: Continue},
		{name: "test10", args: args{path: "rep\\test1\\rep01", exclusion: config.ExclusionType{Map2: map[string][][]string{"rep01": {{"test1", "rep01"}}}}, dir: true}, want: IgnoreRepertoire},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exclusion(tt.args.path, tt.args.exclusion, tt.args.dir); got != tt.want {
				t.Errorf("Exclusion() = %v, want %v", got, tt.want)
			}
		})
	}
}
