package compress

import (
	"embed"
	"errors"
	"fmt"
	"gobackup/internal/config"
	"gobackup/internal/execution"
	"gobackup/internal/hashFiles"
	"gobackup/internal/listFiles"
	"gobackup/internal/listeFichiers"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type ResultatCompress struct {
	ListeFichier []string
}

//go:embed data/archive_vide.7z
var f embed.FS

func Compress(backup config.Backup, global config.BackupGlobal) (ResultatCompress, error) {

	repCompression := fmt.Sprintf("%v/%v", global.RepCompression, backup.Nom)
	err := os.MkdirAll(repCompression, os.ModePerm)
	if err != nil {
		return ResultatCompress{}, fmt.Errorf("erreur pour créer le répertoire %s : %v", repCompression, err)
	}

	var complet bool
	var date time.Time

	log.Printf("calcul complet2")
	complet, date, err = calculComplet2(repCompression, backup, global)
	if err != nil {
		return ResultatCompress{}, err
	}

	var listeFichierCompresse []string
	listeFichiers, err := listFiles.ListeFiles(backup, complet, date, global)
	if err != nil {
		return ResultatCompress{}, err
	}

	if listeFichiers.NbFiles == 0 {
		log.Printf("Aucun fichier à sauvegarder => création d'un zip vide")
		listeFichierCompresse, err = compression(backup, global, "", repCompression, complet)
		if err != nil {
			return ResultatCompress{}, fmt.Errorf("erreur pour compresser le fichier %s (%s) : %v", backup.Nom, listeFichiers.ListeFiles, err)
		}
	} else {
		log.Printf("%d fichiers à sauvegarder", listeFichiers.NbFiles)
		listeFichierCompresse, err = compression(backup, global, listeFichiers.ListeFiles, repCompression, complet)
		if err != nil {
			return ResultatCompress{}, fmt.Errorf("erreur pour compresser le fichier %s (%s) : %v", backup.Nom, listeFichiers.ListeFiles, err)
		}
	}
	for _, f := range listeFichierCompresse {
		err = hashFiles.ConstruitHash(f)
		if err != nil {
			return ResultatCompress{}, err
		}
	}
	err = calculHashFichiers(repCompression)
	if err != nil {
		return ResultatCompress{}, err
	}
	return ResultatCompress{listeFichierCompresse}, nil
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

func calculComplet2(repCompression string, backup config.Backup, global config.BackupGlobal) (bool, time.Time, error) {
	return listeFichiers.Calcul(repCompression, backup, global)
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
	if len(fileList) > 0 {
		args = []string{"a", "-t7z", "-spf", "-bt", "-v1g", res, "@" + fileList}

		program = global.Rep7zip

		log.Printf("compression ...")

		err := execution.Execution(program, args)
		if err != nil {
			return []string{}, err
		}

		log.Printf("compression terminé")
	} else {
		data, err := f.ReadFile("data/archive_vide.7z")
		if err != nil {
			return []string{}, err
		}
		err = os.WriteFile(res, data, 0644)
		if err != nil {
			return []string{}, err
		}
	}

	return listeFichiersCompresse(res, repCompression)
}

func listeFichiersCompresse(res string, repCompression string) ([]string, error) {
	name := path.Base(res)

	var listeFichiers2 []string

	err := filepath.WalkDir(repCompression, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasPrefix(d.Name(), name) {
			listeFichiers2 = append(listeFichiers2, path)
		}
		return nil
	})

	log.Printf("liste dir: %v", listeFichiers2)

	if err != nil {
		return []string{}, err
	} else {
		return listeFichiers2, nil
	}
}
