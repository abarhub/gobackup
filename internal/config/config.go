package config

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type TypeCrypt int

const (
	CryptGpg TypeCrypt = 1 << iota
	CryptAge           = 1 << iota
)

type BackupGlobal struct {
	listeNomBackup       []string
	Rep7zip              string
	RepGpg               string
	RepCompression       string
	RepCryptage          string
	ListeBackup          []Backup
	DateHeure            string
	NbBackupIncremental  int
	Recipient            string
	ActiveVss            bool
	LettreVss            map[string]string
	LogDir               string
	TypeCryptage         TypeCrypt
	RepAge               string
	AgeRecipien          string
	RepArchivageCompress string
	RepArchivageCryptage string
	NbJourArchivage      int
	DebugCompression     bool
	DebugArchivage       bool
}

type Backup struct {
	Nom       string
	Rep       []string
	Set       map[string]bool
	Map2      map[string][]string
	FileListe string
}

func InitialisationConfig(filename string) (BackupGlobal, error) {
	var res = BackupGlobal{}

	file, err := os.Open(filename)
	if err != nil {
		return BackupGlobal{}, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Panic(err)
		}
	}(file)

	mapConfig := map[string]string{}
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// on ignore les commentaires commençant par #
		} else {
			i := strings.IndexRune(line, '=')
			if i >= 0 {
				mapConfig[line[:i]] = line[i+1:]
			}
		}
	}

	rep7zip, ok := mapConfig["global.rep_7zip"]
	if ok {
		res.Rep7zip = strings.TrimSpace(rep7zip)
	}
	repgpg, ok := mapConfig["global.rep_gpg"]
	if ok {
		res.RepGpg = strings.TrimSpace(repgpg)
	}
	repCompression, ok := mapConfig["global.rep_compression"]
	if ok {
		res.RepCompression = strings.TrimSpace(repCompression)
	}
	repCryptage, ok := mapConfig["global.rep_cryptage"]
	if ok {
		res.RepCryptage = strings.TrimSpace(repCryptage)
	}
	nbBackupIncremental, ok := mapConfig["global.nb_backup_incremental"]
	if ok {
		nbBackupIncremental = strings.TrimSpace(nbBackupIncremental)
		if len(nbBackupIncremental) > 0 {
			res.NbBackupIncremental, err = strconv.Atoi(nbBackupIncremental)
			if err != nil {
				return BackupGlobal{}, fmt.Errorf("le paramètre global.nb_backup_incremental n'est pas un nombre", err)
			}
		} else {
			res.NbBackupIncremental = 0
		}
	}
	recipient, ok := mapConfig["global.recipient"]
	if ok {
		res.Recipient = strings.TrimSpace(recipient)
	}
	activeVss, ok := mapConfig["global.activeVss"]
	if ok {
		res.ActiveVss = strings.TrimSpace(activeVss) == "true"
	}
	logdir, ok := mapConfig["global.logdir"]
	if ok {
		res.LogDir = strings.TrimSpace(logdir)
	}
	typeCryptage, ok := mapConfig["global.type_cryptage"]
	if ok {
		if typeCryptage == "gpg" {
			res.TypeCryptage = CryptGpg
		} else if typeCryptage == "age" {
			res.TypeCryptage = CryptAge
		} else {
			return BackupGlobal{}, errors.New("le paramètre typeCryptage n'est pas valide (valeurs possibles: gpg, age)")
		}
	}
	repAge, ok := mapConfig["global.rep_age"]
	if ok {
		res.RepAge = strings.TrimSpace(repAge)
	}
	ageRecipien, ok := mapConfig["global.age_recipien"]
	if ok {
		res.AgeRecipien = strings.TrimSpace(ageRecipien)
	}

	repArchivageCompress, ok := mapConfig["global.rep_archivage_compression"]
	if ok {
		res.RepArchivageCompress = strings.TrimSpace(repArchivageCompress)
	}
	RepArchivageCryptage, ok := mapConfig["global.rep_archivage_cryptage"]
	if ok {
		res.RepArchivageCryptage = strings.TrimSpace(RepArchivageCryptage)
	}
	debugCompression, ok := mapConfig["global.debug_compression"]
	if ok {
		res.DebugCompression = strings.TrimSpace(debugCompression) == "true"
	}
	debugArchivage, ok := mapConfig["global.debug_archivage"]
	if ok {
		res.DebugArchivage = strings.TrimSpace(debugArchivage) == "true"
	}
	nbJourArchive, ok := mapConfig["global.nb_jour_archive"]
	if ok {
		if len(nbJourArchive) > 0 {
			res.NbJourArchivage, err = strconv.Atoi(nbJourArchive)
			if err != nil {
				return BackupGlobal{}, fmt.Errorf("le paramètre global.nb_jour_archive n'est pas un nombre", err)
			}
		} else {
			res.NbBackupIncremental = 0
		}
	}

	res.DateHeure = strings.ReplaceAll(time.Now().Format("20060102_150405.000"), ".", "")

	listeBackup, ok := mapConfig["global.liste_backups"]
	if ok {

		liste := strings.Split(listeBackup, ",")

		for _, v := range liste {

			var res2 = Backup{}
			res2.Nom = v
			debut := "backup." + v
			key := debut + ".rep_a_sauver"
			if aSauver, ok := mapConfig[key]; ok {
				tab := strings.Split(aSauver, ",")
				res2.Rep = tab
			}
			key = debut + ".rep_nom_a_ignorer"
			if repNomAIgnorer, ok := mapConfig[key]; ok {
				tab := strings.Split(repNomAIgnorer, ",")
				set := map[string]bool{}
				for _, v := range tab {
					set[v] = true
				}
				res2.Set = set
			}
			key = debut + ".rep_a_ignorer"
			if repAIgnorer, ok := mapConfig[key]; ok {
				tab := strings.Split(repAIgnorer, ",")
				map2 := map[string][]string{}
				for _, v := range tab {
					addMap(&map2, v)
				}
				res2.Map2 = map2
			}

			fileTemp, err := createTempFile("listeFichiers_" + res2.Nom)
			if err != nil {
				return BackupGlobal{}, fmt.Errorf("erreur pour creer le fichier temporaire : %v", err)
			}
			res2.FileListe = fileTemp

			res.ListeBackup = append(res.ListeBackup, res2)
		}

	}

	if err := scanner.Err(); err != nil {
		return BackupGlobal{}, err
	}

	if len(res.ListeBackup) == 0 {
		return BackupGlobal{}, errors.New("no liste backup")
	}

	if len(res.Rep7zip) == 0 {
		return BackupGlobal{}, errors.New("no 7zip directory")
	}

	if len(res.RepCompression) == 0 {
		return BackupGlobal{}, errors.New("no compress directory")
	}

	if len(res.RepCryptage) == 0 {
		return BackupGlobal{}, errors.New("no crypt directory")
	}

	if res.NbBackupIncremental < 0 {
		return BackupGlobal{}, errors.New("nbBackupIncremental doit être superieur ou égal à 0")
	}

	if res.TypeCryptage == CryptGpg {
		if len(res.RepGpg) == 0 {
			return BackupGlobal{}, errors.New("no gpg directory")
		}

		if len(res.Recipient) == 0 {
			return BackupGlobal{}, errors.New("le paramètre recipient est vide")
		}

	} else {
		if len(res.RepAge) == 0 {
			return BackupGlobal{}, errors.New("no age directory")
		}

		if len(res.AgeRecipien) == 0 {
			return BackupGlobal{}, errors.New("le paramètre ageRecipient est vide")
		}

	}

	return res, nil
}

func addMap(map2 *map[string][]string, s string) {
	tab := strings.Split(s, "\\")
	(*map2)[tab[len(tab)-1]] = tab
}

func createTempFile(name string) (string, error) {
	f, err := os.CreateTemp("", name)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic(err)
		}
	}(f)
	if err != nil {
		return "", fmt.Errorf("erreur pour creer le fichier temporaire : %v", err)
	} else {
		return f.Name(), nil
	}
}
