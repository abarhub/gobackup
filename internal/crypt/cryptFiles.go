package crypt

import (
	"errors"
	"fmt"
	"gobackup/internal/compress"
	"gobackup/internal/config"
	"gobackup/internal/execution"
	"gobackup/internal/hashFiles"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Crypt(fileCompressed compress.ResultatCompress, b config.Backup, global config.BackupGlobal) error {

	repCrypt := fmt.Sprintf("%v/%v", global.RepCryptage, b.Nom)
	err := os.MkdirAll(repCrypt, os.ModePerm)
	if err != nil {
		return err
	}

	err = initialisationCryptage(global)
	if err != nil {
		return err
	}

	for _, file := range fileCompressed.ListeFichier {
		filename := filepath.Base(file)
		if !strings.HasSuffix(filename, ".gpg") && !strings.HasSuffix(filename, ".sha256sum") {
			f := file
			f2 := repCrypt + "/" + filename + ".gpg"
			if _, err := os.Stat(f2); errors.Is(err, os.ErrNotExist) {
				// cryptate du fichier f2
				_, err := cryptFile(f, f2, global)
				if err != nil {
					return err
				}
				// calcul du hash du fichier f2
				err = hashFiles.ConstruitHash(f2)
				if err != nil {
					return err
				}
			} else if err != nil {
				log.Printf("Erreur pour tester l'existance du fichier %s : %v", f2, err.Error())
			} else {
				log.Printf("Fichier %s déjà crypté (%s)\n", file, f2)
			}
		}
	}

	return nil
}

func initialisationCryptage(global config.BackupGlobal) error {
	program := filepath.Dir(global.RepGpg) + "/gpg-connect-agent"
	args := []string{"-v"}

	log.Printf("initialisation agent gpg ...")

	err := execution.Execution(program, args)

	log.Printf("initialisation agent gpg terminé")

	if err != nil {
		return err
	}

	return nil
}

func cryptFile(fileCompressed string, fileCrypted string, global config.BackupGlobal) (string, error) {
	var program string
	var args []string

	program = global.RepGpg
	args = []string{"-v", "--encrypt", "--recipient=" + global.Recipient, "--output=" + fileCrypted,
		fileCompressed}

	log.Printf("cryptage de %v ...", path.Base(fileCompressed))

	err := execution.Execution(program, args)

	log.Printf("cryptage terminé")

	if err != nil {
		return "", err
	} else {
		return fileCrypted, nil
	}
}
