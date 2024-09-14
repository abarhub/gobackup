package crypt

import (
	"errors"
	"fmt"
	"gobackup/internal/compress"
	"gobackup/internal/config"
	"gobackup/internal/execution"
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

	for _, file := range fileCompressed.ListeFichier {
		filename := filepath.Base(file)
		if !strings.HasSuffix(filename, ".gpg") {
			f := file
			f2 := repCrypt + "/" + filename + ".gpg"
			if _, err := os.Stat(f2); errors.Is(err, os.ErrNotExist) {
				_, err := cryptFile(f, f2, global)
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
