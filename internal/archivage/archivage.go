package archivage

import (
	"errors"
	"fmt"
	"gobackup/internal/config"
	"gobackup/internal/hashFiles"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
)

type fichier struct {
	nom     string
	path    string
	date    time.Time
	complet bool
}

func Archivage(global config.BackupGlobal) error {
	if global.NbJourArchivage > 0 {
		if global.RepArchivageCompress != "" {
			log.Printf("archivage des fichiers compressés (nb jours:%d)", global.NbJourArchivage)
			err := archive(global.RepCompression, global.RepArchivageCompress, global.NbJourArchivage, global.DebugArchivage)
			if err != nil {
				return fmt.Errorf("erreur pour archiver les fichiers compressés: %w", err)
			}
		}
		if global.RepArchivageCryptage != "" {
			log.Printf("archivage des fichiers cryptés (nb jours:%d)", global.NbJourArchivage)
			err := archive(global.RepCryptage, global.RepArchivageCryptage, global.NbJourArchivage, global.DebugArchivage)
			if err != nil {
				return fmt.Errorf("erreur pour archiver les fichiers cryptés: %w", err)
			}
		}
	} else {
		log.Printf("pas d'archivage")
	}
	return nil
}

func archive(repertoireSource string, repertoireDestination string, nbJours int, debug bool) error {
	listeFichiers := []fichier{}
	err := filepath.WalkDir(repertoireSource, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			name := d.Name()
			debutComplet := fmt.Sprintf("backupc_")
			debutIncrement := fmt.Sprintf("backupi_")
			if strings.HasPrefix(name, debutComplet) || strings.HasPrefix(name, debutIncrement) {
				suffixOk := false
				for _, extention := range []string{".gpg", ".age", hashFiles.GetExtension()} {
					if strings.HasSuffix(name, extention) {
						suffixOk = true
						break
					}
				}
				if !suffixOk {
					re2 := regexp.MustCompile(`\.7z\.[0-9]{3}$`)
					if re2.MatchString(name) {
						suffixOk = true
					}
				}
				if suffixOk {
					complet := false
					s := name
					if strings.HasPrefix(s, debutComplet) {
						s = strings.TrimPrefix(s, debutComplet)
						complet = true
					} else if strings.HasPrefix(s, debutIncrement) {
						s = strings.TrimPrefix(s, debutIncrement)
					}
					pos := strings.IndexRune(s, '_')
					if pos == -1 {
						// fichier a ne pas traiter
					} else {
						s = s[pos+1:]
						pos2 := strings.IndexRune(s, '.')
						if pos2 == -1 {
							// fichier a ne pas traiter
						} else {
							s = s[:pos2]
							re2 := regexp.MustCompile(`^[0-9]+_[0-9]+$`)
							if re2.MatchString(s) {
								s0 := s[0:len(s)-3] + "." + s[len(s)-3:]
								tt, err0 := time.Parse("20060102_150405.000", s0)
								if err0 != nil {
									// format invalide => on ignore le fichier
								} else {
									f := fichier{nom: d.Name(), date: tt, complet: complet, path: path}
									listeFichiers = append(listeFichiers, f)
								}
							}
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if debug {
		log.Printf("liste des fichiers: %v", listeFichiers)
	}

	sort.Slice(listeFichiers, func(i, j int) bool {
		return listeFichiers[i].date.Before(listeFichiers[j].date)
	})

	slices.Reverse(listeFichiers)

	if debug {
		log.Printf("liste des fichiers triés: %v", listeFichiers)
	}

	now := time.Now()
	max := now.AddDate(0, 0, -nbJours)

	fichierADeplacer := []fichier{}
	etape := 1
	var dateMax time.Time
	for _, fichier := range listeFichiers {
		if etape == 1 {
			if fichier.date.Before(max) {
				etape = 2
			}
		}
		if etape == 2 {
			if fichier.complet {
				etape = 3
				dateMax = fichier.date
			}
		}
		if etape == 3 {
			if dateMax.Year() == fichier.date.Year() && dateMax.YearDay() == fichier.date.YearDay() {
				// même date => on ne fait rien
			} else {
				etape = 4
			}
		}
		if etape == 4 {
			fichierADeplacer = append(fichierADeplacer, fichier)
		}
	}

	log.Printf("nombre de fichiers à déplacer : %d", len(fichierADeplacer))
	if debug {
		log.Printf("liste des fichiers à déplacer: %v", fichierADeplacer)
	}

	for _, fichier := range fichierADeplacer {
		rep := filepath.Base(filepath.Dir(fichier.path))
		src := filepath.Clean(repertoireSource + "/" + rep + "/" + fichier.nom)
		dest := filepath.Clean(repertoireDestination + "/" + rep + "/" + fichier.nom)
		err := moveFile(src, dest)
		if err != nil {
			return fmt.Errorf("erreur pour déplacer le fichier %v (%s - > %s) : %w", fichier, src, dest, err)
		}
	}

	return nil
}

func moveFile(source, destination string) (err error) {
	log.Printf("deplacement %s vers %s", source, destination)
	destDir := filepath.Dir(destination)
	if _, err := os.Stat(destDir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("erreur pour créer le répertoire %s : %w", destDir, err)
		}
	}
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	fi, err := src.Stat()
	if err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := fi.Mode() & os.ModePerm
	dst, err := os.OpenFile(destination, flag, perm)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		dst.Close()
		os.Remove(destination)
		return err
	}
	err = dst.Close()
	if err != nil {
		return err
	}
	err = src.Close()
	if err != nil {
		return err
	}
	err = os.Remove(source)
	if err != nil {
		return err
	}
	return nil
}
