package listeFichiers

import (
	"fmt"
	"gobackup/internal/config"
	"gobackup/internal/hashFiles"
	"io/fs"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Fichiers struct {
	DateStr string
	date    time.Time
	Nom     string
	liste   []string
	complet bool
}

type ListeFichiers struct {
	listeFichiers []Fichiers
	nombackup     string
}

func CreateListeFichier(nomBackup string) *ListeFichiers {

	return &ListeFichiers{
		listeFichiers: []Fichiers{},
		nombackup:     nomBackup,
	}
}

func (liste *ListeFichiers) ajouteRepertoire(repertoire string) error {
	repertoire2 := os.DirFS(repertoire)
	files, err := fs.ReadDir(repertoire2, ".")
	if err != nil {
		return err
	}

	log.Printf("Parcourt ...")
	for _, file := range files {
		if !file.IsDir() {
			err := liste.ajouteFichier(file.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (liste *ListeFichiers) ajouteFichier(nomFichier string) error {
	debutComplet := fmt.Sprintf("backupc_%v_", liste.nombackup)
	debutIncrement := fmt.Sprintf("backupi_%v_", liste.nombackup)
	if strings.HasPrefix(nomFichier, debutComplet) || strings.HasPrefix(nomFichier, debutIncrement) {
		s := nomFichier
		complet := strings.HasPrefix(nomFichier, debutComplet)
		for _, extCrypt := range []string{".gpg", ".age"} {
			if strings.HasSuffix(s, extCrypt) {
				s = strings.TrimSuffix(s, extCrypt)
			}
		}
		if strings.HasSuffix(s, hashFiles.GetExtension()) {
			s = strings.TrimSuffix(s, hashFiles.GetExtension())
		}
		var re = regexp.MustCompile(`\.[0-9]+$`)
		s = re.ReplaceAllString(s, ``)
		if strings.HasSuffix(s, ".7z") {
			s = strings.TrimSuffix(s, ".7z")
		}
		re2 := regexp.MustCompile("^(" + debutComplet + ")|(" + debutIncrement + `)[0-9]+_[0-9]+$`)
		if re2.MatchString(s) {
			s2 := s
			if strings.HasPrefix(s2, debutComplet) {
				s2 = strings.TrimPrefix(s2, debutComplet)
			}
			if strings.HasPrefix(s2, debutIncrement) {
				s2 = strings.TrimPrefix(s2, debutIncrement)
			}
			s2 = s2[0:len(s2)-3] + "." + s2[len(s2)-3:]

			dateDebutTrouve := false
			var dateDebut time.Time
			tt, err0 := time.Parse("20060102_150405.000", s2)
			if err0 != nil {
				// erreur de parsing => on ignore le fichier
			} else {
				dateDebutTrouve = true
				dateDebut = tt
			}

			if dateDebutTrouve {
				trouve := false
				for _, file2 := range liste.listeFichiers {
					if file2.date == dateDebut {
						trouve = true
						file2.liste = append(file2.liste, nomFichier)
						break
					}
				}
				if !trouve {
					file2 := Fichiers{DateStr: s2,
						liste:   []string{nomFichier},
						complet: complet,
						Nom:     s,
						date:    dateDebut,
					}
					liste.listeFichiers = append(liste.listeFichiers, file2)
				}
			} else {
				log.Printf("fichier ignore : %v (%v)", nomFichier, s2)
			}
		}
	}

	return nil
}

func (liste *ListeFichiers) Affiche() {
	log.Printf("liste des fichiers : %v", *liste)
}

func (liste *ListeFichiers) trie() {
	sort.Slice(liste.listeFichiers[:], func(i, j int) bool {
		return liste.listeFichiers[i].DateStr < liste.listeFichiers[j].DateStr
	})
}

func (liste *ListeFichiers) calculComplet(nbBackupIncremental2 int, now time.Time, debugCompression bool) (bool, time.Time, error) {
	nbBackupIncremental := 0
	var dateDebut time.Time
	var dateDebutTrouve = false
	var t1 time.Time
	var backupComplet bool
	if nbBackupIncremental2 > 0 {
		for i := len(liste.listeFichiers) - 1; i >= 0; i-- {
			s := liste.listeFichiers[i]
			if debugCompression {
				log.Printf("boucle %d : %v", i, s)
			}
			if !dateDebutTrouve {
				dateDebutTrouve = true
				dateDebut = s.date
			}
			if debugCompression {
				log.Printf("dateDebutTrouve : %v, dateDebut : %v", dateDebutTrouve, dateDebut)
			}
			if s.complet {
				break
			} else {
				nbBackupIncremental++
			}
			if debugCompression {
				log.Printf("nbBackupIncremental : %d", nbBackupIncremental)
			}
		}
	}

	log.Printf("date: %v (%v), nbBackupIncr: %d, conf nbBackupIncr: %d",
		dateDebut, dateDebutTrouve, nbBackupIncremental, nbBackupIncremental2)

	if nbBackupIncremental2 == 0 || nbBackupIncremental >= nbBackupIncremental2 {
		backupComplet = true
	} else {
		backupComplet = false
		if dateDebutTrouve {
			t1 = time.Date(dateDebut.Year(), dateDebut.Month(), dateDebut.Day(), 0, 0, 0, 0, dateDebut.Location())
		} else {
			t1 = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		}
	}

	if debugCompression {
		log.Printf("liste %v", liste)
	}
	log.Printf("backup complet: %v date: %v", backupComplet, t1)

	return backupComplet, t1, nil
}

func Calcul(repertoire string, backup config.Backup, global config.BackupGlobal) (bool, time.Time, error) {
	listeFichiers := CreateListeFichier(backup.Nom)
	log.Printf("ajout des fichiers du répertoire %s", repertoire)
	err := listeFichiers.ajouteRepertoire(repertoire)
	if err != nil {
		return false, time.Now(), err
	}
	listeFichiers.Affiche()
	log.Printf("trie de la liste ...")
	listeFichiers.trie()
	listeFichiers.Affiche()
	log.Printf("calcul complet ...")
	return listeFichiers.calculComplet(global.NbBackupIncremental, time.Now(), global.DebugCompression)
}
