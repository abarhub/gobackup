package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestInitialisationConfig(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    BackupGlobal
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitialisationConfig(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitialisationConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitialisationConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitialisationConfig_test1(t *testing.T) {
	s := `
[global]
Rep_7zip="c:\\test\\toto.exe"
Rep_gpg="c:/test/titi.exe"
Rep_compression="c:\\titi"
Rep_cryptage="c:\\aaa001"
Nb_backup_incremental=10
Recipient="ddd"
ActiveVss=true
Logdir="aaa"
Type_cryptage="gpg"
Rep_age="c:\\aaa004"
Age_recipien="ccc"
Rep_archivage_compression="c:\\aaa005"
Rep_archivage_cryptage="c:\\aaa006"
Nb_jour_archivage=15
Debug_compression=true
Debug_archivage=true
[backup.nom1]
Rep_a_sauver=["c:\\test_backup"]
`
	filename := t.TempDir() + "/test1.config"
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic(err)
		}
	}(f)
	_, err = f.WriteString(s)
	if err != nil {
		panic(err)
	}
	err = f.Sync()
	if err != nil {
		panic(err)
	}

	res, err := InitialisationConfig(filename)
	if err != nil {
		t.Errorf("InitialisationConfig() error = %v", err)
		return
	}
	compare(t, res.Rep7zip, "c:\\test\\toto.exe", "Rep7zip")
	compare(t, res.RepGpg, "c:/test/titi.exe", "RepGpg")
	compare(t, res.RepCompression, "c:\\titi", "RepCompression")
	compare(t, res.RepCryptage, "c:\\aaa001", "RepCryptage")
	compare(t, res.NbBackupIncremental, 10, "NbBackupIncremental")
	compare(t, res.Recipient, "ddd", "Recipient")
	compare(t, res.ActiveVss, true, "ActiveVss")
	compare(t, res.LogDir, "aaa", "LogDir")
	compare(t, res.TypeCryptage, CryptGpg, "TypeCryptage")
	compare(t, res.RepAge, "c:\\aaa004", "RepAge")
	compare(t, res.AgeRecipien, "ccc", "AgeRecipien")
	compare(t, res.RepArchivageCompress, "c:\\aaa005", "RepArchivageCompress")
	compare(t, res.RepArchivageCryptage, "c:\\aaa006", "RepArchivageCryptage")
	compare(t, res.NbJourArchivage, 15, "NbJourArchivage")
	compare(t, res.DebugCompression, true, "DebugCompression")
	compare(t, res.DebugArchivage, true, "DebugArchivage")
}

func TestInitialisationConfig_test2(t *testing.T) {
	s := `
[global]
Rep_7zip="c:\\test\\toto2.exe"
Rep_gpg="c:/test/titi2.exe"
Rep_compression="c:\\titi2"
Rep_cryptage="c:\\aaa001_2"
Nb_backup_incremental=20
Recipient="ddd_2"
ActiveVss=false
Logdir="aaa_2"
Type_cryptage="age"
Rep_age="c:\\aaa004_2"
Age_recipien="ccc_2"
Rep_archivage_compression="c:\\aaa005_2"
Rep_archivage_cryptage="c:\\aaa006_2"
Nb_jour_archivage=25
Debug_compression=false
Debug_archivage=false
[backup.nom1]
Rep_a_sauver=["c:\\test_backup"]
`
	filename := t.TempDir() + "/test1.config"
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic(err)
		}
	}(f)
	_, err = f.WriteString(s)
	if err != nil {
		panic(err)
	}
	err = f.Sync()
	if err != nil {
		panic(err)
	}

	res, err := InitialisationConfig(filename)
	if err != nil {
		t.Errorf("InitialisationConfig() error = %v", err)
		return
	}
	compare(t, res.Rep7zip, "c:\\test\\toto2.exe", "Rep7zip")
	compare(t, res.RepGpg, "c:/test/titi2.exe", "RepGpg")
	compare(t, res.RepCompression, "c:\\titi2", "RepCompression")
	compare(t, res.RepCryptage, "c:\\aaa001_2", "RepCryptage")
	compare(t, res.NbBackupIncremental, 20, "NbBackupIncremental")
	compare(t, res.Recipient, "ddd_2", "Recipient")
	compare(t, res.ActiveVss, false, "ActiveVss")
	compare(t, res.LogDir, "aaa_2", "LogDir")
	compare(t, res.TypeCryptage, CryptAge, "TypeCryptage")
	compare(t, res.RepAge, "c:\\aaa004_2", "RepAge")
	compare(t, res.AgeRecipien, "ccc_2", "AgeRecipien")
	compare(t, res.RepArchivageCompress, "c:\\aaa005_2", "RepArchivageCompress")
	compare(t, res.RepArchivageCryptage, "c:\\aaa006_2", "RepArchivageCryptage")
	compare(t, res.NbJourArchivage, 25, "NbJourArchivage")
	compare(t, res.DebugCompression, false, "DebugCompression")
	compare(t, res.DebugArchivage, false, "DebugArchivage")
}

func TestInitialisationConfig_test3(t *testing.T) {
	s := `
[global]
Rep_7zip="c:\\test\\toto2.exe"
Rep_gpg="c:/test/titi2.exe"
Rep_compression="c:\\titi2"
Rep_cryptage="c:\\aaa001_2"
Nb_backup_incremental=20
Recipient="ddd_2"
ActiveVss=true
Logdir="aaa_2"
Type_cryptage="age"
Rep_age="c:\\aaa004_2"
Age_recipien="ccc_2"
Rep_archivage_compression="c:\\aaa005_2"
Rep_archivage_cryptage="c:\\aaa006_2"
Nb_jour_archivage=25
Debug_compression=false
Debug_archivage=true
[backup.nom1]
Rep_a_sauver=["c:\\test_backup"]
Rep_nom_a_ignorer=[]
Rep_a_ignorer=[]
`
	filename := t.TempDir() + "/test1.config"
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic(err)
		}
	}(f)
	_, err = f.WriteString(s)
	if err != nil {
		panic(err)
	}
	err = f.Sync()
	if err != nil {
		panic(err)
	}

	res, err := InitialisationConfig(filename)
	if err != nil {
		t.Errorf("InitialisationConfig() error = %v", err)
		return
	}
	compare(t, res.Rep7zip, "c:\\test\\toto2.exe", "Rep7zip")
	compare(t, res.RepGpg, "c:/test/titi2.exe", "RepGpg")
	compare(t, res.RepCompression, "c:\\titi2", "RepCompression")
	compare(t, res.RepCryptage, "c:\\aaa001_2", "RepCryptage")
	compare(t, res.NbBackupIncremental, 20, "NbBackupIncremental")
	compare(t, res.Recipient, "ddd_2", "Recipient")
	compare(t, res.ActiveVss, true, "ActiveVss")
	compare(t, res.LogDir, "aaa_2", "LogDir")
	compare(t, res.TypeCryptage, CryptAge, "TypeCryptage")
	compare(t, res.RepAge, "c:\\aaa004_2", "RepAge")
	compare(t, res.AgeRecipien, "ccc_2", "AgeRecipien")
	compare(t, res.RepArchivageCompress, "c:\\aaa005_2", "RepArchivageCompress")
	compare(t, res.RepArchivageCryptage, "c:\\aaa006_2", "RepArchivageCryptage")
	compare(t, res.NbJourArchivage, 25, "NbJourArchivage")
	compare(t, res.DebugCompression, false, "DebugCompression")
	compare(t, res.DebugArchivage, true, "DebugArchivage")
	if len(res.ListeBackup) != 1 {
		t.Errorf("InitialisationConfig() aucun backup: %d", len(res.ListeBackup))
	} else {
		compare(t, res.ListeBackup[0].Nom, "nom1", "ListeBackup.nom")
		compareDeep(t, res.ListeBackup[0].Rep, []string{"c:\\test_backup"}, "ListeBackup.Rep")
		compareDeepMapBool(t, res.ListeBackup[0].Exclusion.Set, map[string]bool{}, "ListeBackup.Set")
		compareDeepMapListString(t, res.ListeBackup[0].Exclusion.Map2, map[string][][]string{}, "ListeBackup.Map2")
	}
}

func TestInitialisationConfig_test4(t *testing.T) {
	s := `
[global]
Rep_7zip="c:\\test\\toto2.exe"
Rep_gpg="c:/test/titi2.exe"
Rep_compression="c:\\titi2"
Rep_cryptage="c:\\aaa001_2"
Nb_backup_incremental=20
Recipient="ddd_2"
ActiveVss=false
Logdir="aaa_2"
Type_cryptage="age"
Rep_age="c:\\aaa004_2"
Age_recipien="ccc_2"
Rep_archivage_compression="c:\\aaa005_2"
Rep_archivage_cryptage="c:\\aaa006_2"
Nb_jour_archivage=25
Debug_compression=true
Debug_archivage=true
[backup.nom1]
Rep_a_sauver=["c:\\test_backup"]
Rep_nom_a_ignorer=["aaa","bbb"]
Rep_a_ignorer=["rep1\\rep2","rep3\\rep4\\rep5"]
[backup.nom2]
Rep_a_sauver=["c:\\test_backup2"]
Rep_nom_a_ignorer=["XXX"]
Rep_a_ignorer=["rep01\\rep04"]
`
	filename := t.TempDir() + "/test1.config"
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic(err)
		}
	}(f)
	_, err = f.WriteString(s)
	if err != nil {
		panic(err)
	}
	err = f.Sync()
	if err != nil {
		panic(err)
	}

	res, err := InitialisationConfig(filename)
	if err != nil {
		t.Errorf("InitialisationConfig() error = %v", err)
		return
	}
	compare(t, res.Rep7zip, "c:\\test\\toto2.exe", "Rep7zip")
	compare(t, res.RepGpg, "c:/test/titi2.exe", "RepGpg")
	compare(t, res.RepCompression, "c:\\titi2", "RepCompression")
	compare(t, res.RepCryptage, "c:\\aaa001_2", "RepCryptage")
	compare(t, res.NbBackupIncremental, 20, "NbBackupIncremental")
	compare(t, res.Recipient, "ddd_2", "Recipient")
	compare(t, res.ActiveVss, false, "ActiveVss")
	compare(t, res.LogDir, "aaa_2", "LogDir")
	compare(t, res.TypeCryptage, CryptAge, "TypeCryptage")
	compare(t, res.RepAge, "c:\\aaa004_2", "RepAge")
	compare(t, res.AgeRecipien, "ccc_2", "AgeRecipien")
	compare(t, res.RepArchivageCompress, "c:\\aaa005_2", "RepArchivageCompress")
	compare(t, res.RepArchivageCryptage, "c:\\aaa006_2", "RepArchivageCryptage")
	compare(t, res.NbJourArchivage, 25, "NbJourArchivage")
	compare(t, res.DebugCompression, true, "DebugCompression")
	compare(t, res.DebugArchivage, true, "DebugArchivage")
	if len(res.ListeBackup) != 2 {
		t.Errorf("InitialisationConfig() aucun backup: %d", len(res.ListeBackup))
	} else {
		var backup Backup
		backup, err = getBackup(res.ListeBackup, "nom1")
		if err != nil {
			t.Errorf("GetBackup() error = %v, nom=%s", err, "nom1")
		} else {
			compare(t, backup.Nom, "nom1", "ListeBackup.nom")
			compareDeep(t, backup.Rep, []string{"c:\\test_backup"}, "ListeBackup.Rep")
			compareDeepMapBool(t, backup.Exclusion.Set, map[string]bool{"aaa": true, "bbb": true}, "ListeBackup.Set")
			compareDeepMapListString(t, backup.Exclusion.Map2, map[string][][]string{"rep2": [][]string{{"rep1", "rep2"}}, "rep5": [][]string{{"rep3", "rep4", "rep5"}}}, "ListeBackup.Map2")
		}
		backup, err = getBackup(res.ListeBackup, "nom2")
		if err != nil {
			t.Errorf("GetBackup() error = %v, nom=%s", err, "nom2")
		} else {
			compare(t, backup.Nom, "nom2", "ListeBackup.nom")
			compareDeep(t, backup.Rep, []string{"c:\\test_backup2"}, "ListeBackup.Rep")
			compareDeepMapBool(t, backup.Exclusion.Set, map[string]bool{"XXX": true}, "ListeBackup.Set")
			compareDeepMapListString(t, backup.Exclusion.Map2, map[string][][]string{"rep04": [][]string{{"rep01", "rep04"}}}, "ListeBackup.Map2")
		}
	}
}

func TestAjoutExclusion(t *testing.T) {
	s := `
[global]
[backup.nom1]
Rep_nom_a_ignorer=["aaa","bbb"]
Rep_a_ignorer=["rep1\\rep2","rep3\\rep4\\rep5"]
`
	var config configToml
	_, err := toml.Decode(s, &config)
	if err != nil {
		t.Errorf("InitialisationConfig() error = %v", err)
		return
	}

	if len(config.Backup) != 1 {
		t.Errorf("InitialisationConfig() aucun backup: %d", len(config.Backup))
	} else {
		exclusion := AjoutExclusion(config.Backup["nom1"])
		compareDeepMapBool(t, exclusion.Set, map[string]bool{"aaa": true, "bbb": true}, "ListeBackup.Set")
		compareDeepMapListString(t, exclusion.Map2, map[string][][]string{"rep2": {{"rep1", "rep2"}}, "rep5": {{"rep3", "rep4", "rep5"}}}, "ListeBackup.Map2")
	}
}

func getBackup(listeBackup []Backup, nom string) (Backup, error) {
	for _, backup := range listeBackup {
		if backup.Nom == nom {
			return backup, nil
		}
	}
	return Backup{}, errors.New("Nom " + nom + " not found")
}

func compare(t *testing.T, got, want interface{}, nomChamps string) {
	if got != want {
		t.Errorf("InitialisationConfig() got = %v, want %v for %s", got, want, nomChamps)
	}
}

func compareDeep(t *testing.T, got, want interface{}, nomChamps string) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("InitialisationConfig() got = %v, want %v for %s", got, want, nomChamps)
	}
}

func compareDeepMapBool(t *testing.T, got, want map[string]bool, nomChamps string) {
	if (got == nil || len(got) == 0) && (want == nil || len(want) == 0) {
		return
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("InitialisationConfig() got = %v, want %v for %s", got, want, nomChamps)
	}
}

func compareDeepMapListString(t *testing.T, got, want map[string][][]string, nomChamps string) {
	if (got == nil || len(got) == 0) && (want == nil || len(want) == 0) {
		return
	} else if !reflect.DeepEqual(got, want) {
		t.Errorf("InitialisationConfig() got = %v, want %v for %s", got, want, nomChamps)
	}
}

func Test_addMap(t *testing.T) {
	type args struct {
		map2 *map[string][][]string
		s    string
	}
	tests := []struct {
		name string
		args args
		want *map[string][][]string
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{map2: &map[string][][]string{}, s: "rep/test1"}, want: &map[string][][]string{"test1": {{"rep", "test1"}}}},
		{name: "test2", args: args{map2: &map[string][][]string{}, s: "rep/test1/test2"}, want: &map[string][][]string{"test2": {{"rep", "test1", "test2"}}}},
		{name: "test3", args: args{map2: &map[string][][]string{"test1": {{"rep0", "test1"}}}, s: "rep/test1"}, want: &map[string][][]string{"test1": {{"rep0", "test1"}, {"rep", "test1"}}}},
		{name: "test4", args: args{map2: &map[string][][]string{}, s: "rep\\test1\\test2"}, want: &map[string][][]string{"test2": {{"rep", "test1", "test2"}}}},
		{name: "test5", args: args{map2: &map[string][][]string{}, s: "rep/test1\\test2"}, want: &map[string][][]string{"test2": {{"rep", "test1", "test2"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addMap(tt.args.map2, tt.args.s)
			if !reflect.DeepEqual(tt.args.map2, tt.want) {
				t.Errorf("addMap() got = %v, want %v", tt.args.map2, tt.want)
			}
		})
	}
}

func Test_createTempFile(t *testing.T) {
	filename := "toto.txt"
	got, err := createTempFile(filename)
	if (err != nil) != false {
		t.Errorf("createTempFile() error = %v", err)
		return
	}
	if !strings.Contains(got, filename) {
		t.Errorf("createTempFile() got = %v, want %v", got, filename)
	}
	if _, err := os.Stat(got); errors.Is(err, os.ErrNotExist) {
		t.Errorf("createTempFile() le fichier %s n'existe pas", got)
	} else {
		err = os.Remove(got)
		if err != nil {
			t.Errorf("createTempFile() delete %s error = %v", got, err)
		}
	}
}
