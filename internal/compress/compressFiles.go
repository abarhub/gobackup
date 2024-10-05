package compress

import (
	"errors"
	"fmt"
	"gobackup/internal/config"
	"gobackup/internal/execution"
	"gobackup/internal/hashFiles"
	"gobackup/internal/listFiles"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

type ResultatCompress struct {
	ListeFichier []string
}

func Compress(backup config.Backup, global config.BackupGlobal) (ResultatCompress, error) {

	repCompression := fmt.Sprintf("%v/%v", global.RepCompression, backup.Nom)
	err := os.MkdirAll(repCompression, os.ModePerm)
	if err != nil {
		return ResultatCompress{}, fmt.Errorf("erreur pour créer le répertoire %s : %v", repCompression, err)
	}

	complet, date, err := calculComplet(repCompression, backup, global)
	if err != nil {
		return ResultatCompress{}, err
	}

	listeFichiers, err := listFiles.ListeFiles(backup, complet, date, global)
	if err != nil {
		return ResultatCompress{}, err
	}

	if listeFichiers.NbFiles == 0 {
		log.Printf("Aucun fichier à sauvegarder")
		return ResultatCompress{}, nil
	} else {
		log.Printf("%d fichiers à sauvegarder", listeFichiers.NbFiles)
		listeFichierCompresse, err := compression(backup, global, listeFichiers.ListeFiles, repCompression, complet)
		if err != nil {
			return ResultatCompress{}, fmt.Errorf("erreur pour compresser le fichier %s (%s) : %v", backup.Nom, listeFichiers.ListeFiles, err)
		} else {
			for _, f := range listeFichierCompresse {
				err = hashFiles.ConstruitHash(f)
				if err != nil {
					return ResultatCompress{}, err
				}
			}
			err := calculHashFichiers(repCompression)
			if err != nil {
				return ResultatCompress{}, err
			}
			return ResultatCompress{listeFichierCompresse}, nil
		}
	}
}

func calculHashFichiers(repCompression string) error {
	files, err := os.ReadDir(repCompression)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := filepath.Base(file.Name())
		if !strings.HasSuffix(filename, hashFiles.GetExtension()) {
			f := repCompression + "/" + file.Name()
			fileHash := f + hashFiles.GetExtension()
			if _, err := os.Stat(fileHash); errors.Is(err, os.ErrNotExist) {
				// calcul du hash du fichier
				log.Printf("Calcul du hash du fichier %s", f)
				err = hashFiles.ConstruitHash(f)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func trimStringFromString(s string, s2 string) string {
	if idx := strings.LastIndex(s, s2); idx != -1 {
		return s[:idx]
	}
	return s
}

func calculComplet(repCompression string, backup config.Backup, global config.BackupGlobal) (bool, time.Time, error) {
	files, err := os.ReadDir(repCompression)
	if err != nil {
		return false, time.Time{}, err
	}
	var liste []string

	debutComplet := fmt.Sprintf("backupc_%v_", backup.Nom)
	debutIncrement := fmt.Sprintf("backupi_%v_", backup.Nom)

	log.Printf("Parcourt ...")
	for _, file := range files {
		log.Printf("file: %v", file.Name())
		if !file.IsDir() && (strings.HasPrefix(file.Name(), debutComplet) || strings.HasPrefix(file.Name(), debutIncrement)) {
			s := file.Name()
			log.Printf("s001: %v", s)
			for _, extCrypt := range []string{".gpg", ".age"} {
				if strings.HasSuffix(s, extCrypt) {
					s = strings.TrimSuffix(s, extCrypt)
				}
			}
			log.Printf("s002: %v", s)
			if strings.HasSuffix(s, hashFiles.GetExtension()) {
				s = strings.TrimSuffix(s, hashFiles.GetExtension())
			}
			log.Printf("s003: %v", s)
			var re = regexp.MustCompile(`\.[0-9]+$`)
			s = re.ReplaceAllString(s, ``)
			log.Printf("s004: %v", s)
			if strings.HasSuffix(s, ".7z") {
				s = strings.TrimSuffix(s, ".7z")
			}
			log.Printf("s: %v", s)
			re2 := regexp.MustCompile("^(" + debutComplet + ")|(" + debutIncrement + `)[0-9]+_[0-9]+$`)
			if re2.MatchString(s) {
				log.Printf("match: %v", s)
				//s = trimStringFromString(s, "_")
				if !slices.Contains(liste, s) {
					log.Printf("append: %v", s)
					liste = append(liste, s)
				}
			}
		}
	}

	log.Printf("trie de la liste %v ...", liste)
	slices.SortFunc(liste,
		func(a, b string) int {
			var t01, t02 time.Time
			a1 := strings.TrimPrefix(a, debutComplet)
			a1 = strings.TrimPrefix(a1, debutIncrement)
			b1 := strings.TrimPrefix(b, debutComplet)
			b1 = strings.TrimPrefix(b1, debutIncrement)
			s0 := a1[0:len(a1)-3] + "." + a1[len(a1)-3:]
			tt, err0 := time.Parse("20060102_150405.000", s0)
			if err0 != nil {
				// erreur de parsing => on ignore le fichier
				log.Printf("a=%s, a1=%s, s0=%s", a, a1, s0)
				panic(err0)
			} else {
				t01 = tt
			}
			s0 = b1[0:len(b1)-3] + "." + a1[len(b1)-3:]
			tt, err0 = time.Parse("20060102_150405.000", s0)
			if err0 != nil {
				// erreur de parsing => on ignore le fichier
				log.Printf("b=%s, b1=%s, s0=%s", b, b1, s0)
				panic(err0)
			} else {
				t02 = tt
			}
			//return cmp.Compare(t01, t02)
			if t01 == t02 {
				return 0
			} else if t01.Before(t02) {
				return -1
			} else {
				return 1
			}
		})

	//sort.Sort(sort.StringSlice(liste))

	log.Printf("liste sorted : %v", liste)

	nbBackupIncremental := 0
	var dateDebut time.Time
	var dateDebutTrouve = false
	var t1 time.Time
	var backupComplet bool
	if global.NbBackupIncremental > 0 {
		for i := len(liste) - 1; i >= 0; i-- {
			s := liste[i]
			log.Printf("boucle %d : %s", i, s)
			if !dateDebutTrouve {
				var s0 string
				if strings.HasPrefix(s, debutComplet) {
					s0 = strings.TrimPrefix(s, debutComplet)
				} else if strings.HasPrefix(s, debutIncrement) {
					s0 = strings.TrimPrefix(s, debutIncrement)
				}
				if len(s0) == 18 {
					s0 = s0[0:len(s0)-3] + "." + s0[len(s0)-3:]
					log.Printf("s0 : %s", s0)
					tt, err0 := time.Parse("20060102_150405.000", s0)
					if err0 != nil {
						// erreur de parsing => on ignore le fichier
					} else {
						dateDebutTrouve = true
						dateDebut = tt
					}
				}
			}
			log.Printf("dateDebutTrouve : %v, dateDebut : %v", dateDebutTrouve, dateDebut)
			if strings.HasPrefix(s, "backupc_") {
				break
			} else {
				nbBackupIncremental++
			}
			log.Printf("nbBackupIncremental : %s", nbBackupIncremental)
		}
	}

	log.Printf("date: %v (%v), nbBackupIncr: %d, conf nbBackupIncr: %d",
		dateDebut, dateDebutTrouve, nbBackupIncremental, global.NbBackupIncremental)

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

func trimPrefix(s string, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

func compression(backup config.Backup, global config.BackupGlobal, fileList string, repCompression string, complet bool) ([]string, error) {
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
	if err != nil {
		return []string{}, err
	}

	log.Printf("compression terminé")

	name := path.Base(res)

	var listeFichiers []string

	err = filepath.WalkDir(repCompression, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), name) {
			listeFichiers = append(listeFichiers, path)
		}
		return nil
	})

	log.Printf("liste dir: %v", listeFichiers)

	if err != nil {
		return []string{}, err
	} else {
		return listeFichiers, nil
	}
}
