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
		return 0, err
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
			fileName := filepath.Base(path)

			_, ok := res.Set[fileName]
			if ok {
				fmt.Printf("Répertoire ignoré: %q\n", path)
				return filepath.SkipDir
			}

			_, ok2 := res.Map2[fileName]
			if ok2 {
				tab := strings.Split(path, "\\")
				if testEqSuffixSlice(res.Map2[fileName], tab) {
					fmt.Printf("Répertoire ignoré: %q\n", path)
					return filepath.SkipDir
				}
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

func convertie(root string, global config.BackupGlobal) string {
	if len(root) >= 2 && root[1] == ':' && len(global.LettreVss) > 0 {
		lettre := root[0]
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
	for i := len(suffix) - 1; i >= 0; i-- {
		if suffix[i] != tab[i] {
			return false
		}
	}

	return true
}
