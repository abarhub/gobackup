package crypt

import (
	"errors"
	"fmt"
	"gobackup/internal/config"
	"gobackup/internal/execution"
	"log"
	"os"
	"path"
	"strings"
)

func Crypt(fileCompressed string, b config.Backup, global config.BackupGlobal) error {

	rep := path.Dir(fileCompressed)
	filename := path.Base(fileCompressed)
	repCrypt := fmt.Sprintf("%v/%v", global.RepCryptage, b.Nom)
	err := os.MkdirAll(repCrypt, os.ModePerm)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(rep)
	if err != nil {
		return err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), filename) && !strings.HasSuffix(file.Name(), ".gpg") {
			f := rep + "\\" + file.Name()
			f2 := repCrypt + "/" + file.Name() + ".gpg"
			if _, err := os.Stat(f2); errors.Is(err, os.ErrNotExist) {
				_, err := cryptFile(f, f2, global)
				if err != nil {
					return err
				}
			} else {
				log.Printf("File %s is already crypted\n", file.Name())
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
