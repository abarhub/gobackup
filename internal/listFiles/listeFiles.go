package listFiles

import (
	"fmt"
	"gobackup/internal/config"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ListeFichiers struct {
	ListeFiles string
	NbFiles    int
}

type IgnoreParcourt int

const (
	Continue IgnoreParcourt = iota + 1
	IgnoreRepertoire
	IgnoreFichier
)

func ListeFiles(backup config.Backup, complet bool, date time.Time, global config.BackupGlobal) (ListeFichiers, error) {

	log.Printf("ecriture de la liste des fichiers dans  %s (complet=%v) ...\n", backup.FileListe, complet)

	start := time.Now()

	nbFichiers, err := parcourt(backup, complet, date, global)
	if err != nil {
		return ListeFichiers{}, err
	}

	elapsed := time.Since(start)

	log.Printf("parcourt %s", elapsed)

	return ListeFichiers{backup.FileListe, nbFichiers}, nil
}

func parcourt(res config.Backup, complet bool, date time.Time, configGlobal config.BackupGlobal) (int, error) {
	nbFichier := 0
	f, err := os.OpenFile(res.FileListe, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return 0, fmt.Errorf("erreur pour creer le fichier '%s' : %v", res.FileListe, err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panic(err)
		}
	}(f)

	for i := range res.Rep {
		root0 := res.Rep[i]
		root := convertie(root0, configGlobal)
		log.Printf("Parcourt de %q (%v)\n", root0, root)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Erreur d'accès à %q: %v\n", path, err)
				return err
			}

			ignore := Exclusion(path, res.Exclusion, info.IsDir())
			if ignore == IgnoreFichier {
				fmt.Printf("Fichier ignoré: %q\n", path)
				return nil
			} else if ignore == IgnoreRepertoire {
				fmt.Printf("Répertoire ignoré: %q\n", path)
				return filepath.SkipDir
			}

			if !info.IsDir() {

				traitement := false

				if complet {
					traitement = true
				} else {
					if info.ModTime().After(date) {
						traitement = true
					}
				}

				if traitement {
					nbFichier++
					if _, err = f.WriteString(fmt.Sprintf("%s\n", path)); err != nil {
						return err
					}
				}
			}

			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("Erreur lors du parcours : %v\n", err)
		}
	}

	return nbFichier, nil
}

func Exclusion(path string, exclusion config.ExclusionType, dir bool) IgnoreParcourt {
	fileName := filepath.Base(path)

	if dir {
		_, ok := exclusion.Set[fileName]
		if ok {
			return IgnoreRepertoire
		}

		_, ok2 := exclusion.Map2[fileName]
		if ok2 {
			s := strings.Replace(path, "\\", "/", -1)
			tab := strings.Split(s, "/")
			for _, list := range exclusion.Map2[fileName] {
				if testEqSuffixSlice(list, tab) {
					return IgnoreRepertoire
				}
			}
		}
	} else {
		_, ok := exclusion.Set[fileName]
		if ok {
			return IgnoreFichier
		}

		_, ok2 := exclusion.Map2[fileName]
		if ok2 {
			s := strings.Replace(path, "\\", "/", -1)
			tab := strings.Split(s, "/")
			for _, list := range exclusion.Map2[fileName] {
				if testEqSuffixSlice(list, tab) {
					return IgnoreFichier
				}
			}
		}
	}
	return Continue
}

func convertie(root string, global config.BackupGlobal) string {
	if len(root) >= 2 && root[1] == ':' && len(global.LettreVss) > 0 {
		lettre := strings.ToUpper(root)[0]
		if link, ok := global.LettreVss[string(rune(lettre))]; ok {
			root2 := link + root[2:]
			return root2
		}
	}
	return root
}

func testEqSuffixSlice(suffix, tab []string) bool {
	if len(suffix) > len(tab) {
		return false
	}
	for i := 0; i < len(suffix); i++ {
		if suffix[len(suffix)-i-1] != tab[len(tab)-i-1] {
			return false
		}
	}

	return true
}
