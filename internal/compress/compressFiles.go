package compress

import (
	"fmt"
	"gobackup/internal/config"
	"gobackup/internal/execution"
	"gobackup/internal/listFiles"
	"log"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
)

func Compress(backup config.Backup, global config.BackupGlobal) (string, error) {

	var res string

	repCompression := fmt.Sprintf("%v/%v", global.RepCompression, backup.Nom)
	err := os.MkdirAll(repCompression, os.ModePerm)
	if err != nil {
		return "", err
	}

	complet, date, err := calculComplet(repCompression, backup, global)
	if err != nil {
		return "", err
	}

	listeFichiers, err := listFiles.ListeFiles(backup, complet, date, global)
	if err != nil {
		return "", err
	}

	if listeFichiers.NbFiles == 0 {
		log.Printf("Aucun fichier à sauvegarder")
		return "", nil
	} else {
		log.Printf("%d fichiers à sauvegarder", listeFichiers.NbFiles)
		res, err = compression(backup, global, listeFichiers.ListeFiles, repCompression, complet, date)
		if err != nil {
			return "", fmt.Errorf("erreur pour compresser le fichier %s (%s) : %v", backup.Nom, listeFichiers.ListeFiles, err)
		} else {
			return res, nil
		}
	}
}

func calculComplet(repCompression string, backup config.Backup, global config.BackupGlobal) (bool, time.Time, error) {
	files, err := os.ReadDir(repCompression)
	if err != nil {
		return false, time.Time{}, err
	}
	var liste []string

	debutComplet := fmt.Sprintf("backupc_%v_", backup.Nom)
	debutIncrement := fmt.Sprintf("backupi_%v_", backup.Nom)

	for _, file := range files {
		if !file.IsDir() && (strings.HasPrefix(file.Name(), debutComplet) || strings.HasPrefix(file.Name(), debutIncrement)) {
			s := file.Name()
			if strings.HasSuffix(s, ".gpg") {
				s = strings.TrimSuffix(s, ".gpg")
			}
			var re = regexp.MustCompile(`\.[0-9]+$`)
			s = re.ReplaceAllString(s, ``)
			if strings.HasSuffix(s, ".7z") {
				s = strings.TrimSuffix(s, ".7z")
			}
			re2 := regexp.MustCompile("^(" + debutComplet + ")|(" + debutIncrement + `)[0-9]+_[0-9]+$`)
			if re2.MatchString(s) {
				if !slices.Contains(liste, s) {
					liste = append(liste, s)
				}
			}
		}
	}

	sort.Sort(sort.StringSlice(liste))

	nbBackupIncremental := 0
	var dateDebut time.Time
	var dateDebutTrouve = false
	var t1 time.Time
	var backupComplet bool
	if global.NbBackupIncremental > 0 {
		for i := len(liste) - 1; i >= 0; i-- {
			s := liste[i]
			if !dateDebutTrouve {
				var s0 string
				if strings.HasPrefix(s, debutComplet) {
					s0 = strings.TrimPrefix(s, debutComplet)
				} else if strings.HasPrefix(s, debutIncrement) {
					s0 = strings.TrimPrefix(s, debutIncrement)
				}
				if len(s0) == 18 {
					s0 = s0[0:len(s0)-3] + "." + s0[len(s0)-3:]
					tt, err0 := time.Parse("20060102_150405.000", s0)
					if err0 != nil {
						// erreur de parsing => on ignore le fichier
					} else {
						dateDebutTrouve = true
						dateDebut = tt
					}
				}
			}
			if strings.HasPrefix(s, "backupc_") {
				break
			} else {
				nbBackupIncremental++
			}
		}
	}

	log.Printf("date: %v (%v), nbBackupIncr: %d", dateDebut, dateDebutTrouve, nbBackupIncremental)

	if global.NbBackupIncremental == 0 || nbBackupIncremental > global.NbBackupIncremental {
		backupComplet = true
	} else {
		backupComplet = false
		if dateDebutTrouve {
			t1 = time.Date(dateDebut.Year(), dateDebut.Month(), dateDebut.Day(), 0, 0, 0, 0, dateDebut.Location())
		} else {
			now := time.Now()
			t1 = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		}
	}

	log.Printf("liste %v", liste)
	log.Printf("backup complet: %v date: %v", backupComplet, t1)

	return backupComplet, t1, nil
}

func compression(backup config.Backup, global config.BackupGlobal, fileList string, repCompression string, complet bool, date time.Time) (string, error) {
	var program string
	var args []string
	var res string

	log.Printf("Préparation de la compression")

	var c string
	if complet {
		c = "c"
	} else {
		c = "i"
	}
	res = fmt.Sprintf("%v/backup%s_%v_%s.7z", repCompression, c, backup.Nom, global.DateHeure)

	program = global.Rep7zip
	args = []string{"a", "-t7z", "-spf", "-bt", "-v1g", res, "@" + fileList}

	log.Printf("compression ...")

	err := execution.Execution(program, args)

	log.Printf("compression terminé")

	if err != nil {
		return "", err
	} else {
		return res, nil
	}
}
