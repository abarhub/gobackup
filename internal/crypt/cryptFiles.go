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

const extension = ".gpg"
const extension2 = ".age"

func Crypt(fileCompressed compress.ResultatCompress, b config.Backup, global config.BackupGlobal) error {

	repCrypt := fmt.Sprintf("%v/%v", global.RepCryptage, b.Nom)
	err := os.MkdirAll(repCrypt, os.ModePerm)
	if err != nil {
		return fmt.Errorf("erreur pour creer le répertoire %s : %w", repCrypt, err)
	}

	err = initialisationCryptage(global)
	if err != nil {
		return fmt.Errorf("erreur pour initialiser le cryptage : %w", err)
	}

	for _, file := range fileCompressed.ListeFichier {
		filename := filepath.Base(file)
		if !strings.HasSuffix(filename, extension) && !strings.HasSuffix(filename, extension2) && !strings.HasSuffix(filename, hashFiles.GetExtension()) {
			f := file
			var f2 = ""
			if global.TypeCryptage == config.CryptGpg {
				f2 = repCrypt + "/" + filename + extension
			} else {
				f2 = repCrypt + "/" + filename + extension2
			}

			if _, err := os.Stat(f2); errors.Is(err, os.ErrNotExist) {
				// cryptate du fichier f2
				_, err := cryptFile(f, f2, global)
				if err != nil {
					return fmt.Errorf("erreur pour crypter le fichier %s : %w", f, err)
				}
				// calcul du hash du fichier f2
				err = hashFiles.ConstruitHash(f2)
				if err != nil {
					return fmt.Errorf("erreur pour calculer le hash du fichier %s : %w", f2, err)
				}
			} else if err != nil {
				log.Printf("Erreur pour tester l'existance du fichier %s : %v", f2, err.Error())
			} else {
				log.Printf("Fichier %s déjà crypté (%s)\n", file, f2)
			}
		}
	}

	err = cryptFichiersNonCryptes(global, repCrypt, b)
	if err != nil {
		return fmt.Errorf("erreur pour crypter les fichiers non crypte : %w", err)
	}

	err = hashFichiersNonHashes(repCrypt)
	if err != nil {
		return fmt.Errorf("erreur pour calculer le hash des fichiers non hashés : %w", err)
	}

	return nil
}

func hashFichiersNonHashes(cryptRepertoire string) error {
	files, err := os.ReadDir(cryptRepertoire)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := filepath.Base(file.Name())
		if !strings.HasSuffix(filename, hashFiles.GetExtension()) &&
			(strings.Contains(filename, extension) || strings.Contains(filename, extension2)) {
			fichierCrypte := cryptRepertoire + "/" + filename
			fichierHash := fichierCrypte + hashFiles.GetExtension()
			if _, err := os.Stat(fichierHash); errors.Is(err, os.ErrNotExist) {
				log.Printf("Calcul du hash de %s", fichierCrypte)
				// calcul du hash du fichier fichierCrypte
				err = hashFiles.ConstruitHash(fichierCrypte)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func cryptFichiersNonCryptes(global config.BackupGlobal, cryptRepertoire string, b config.Backup) error {
	repCompression := global.RepCompression + "/" + b.Nom
	files, err := os.ReadDir(repCompression)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := filepath.Base(file.Name())
		if !strings.HasSuffix(filename, hashFiles.GetExtension()) && strings.Contains(filename, ".7z") {
			fichierCompresse := repCompression + "/" + file.Name()
			fichierCrypte := cryptRepertoire + "/" + filename + extension
			if _, err := os.Stat(fichierCrypte); errors.Is(err, os.ErrNotExist) {
				fichierCrypte2 := cryptRepertoire + "/" + filename + extension2
				if _, err := os.Stat(fichierCrypte2); errors.Is(err, os.ErrNotExist) {
					fichierCrypte3 := ""
					if global.TypeCryptage == config.CryptGpg {
						fichierCrypte3 = fichierCrypte
					} else {
						fichierCrypte3 = fichierCrypte2
					}
					log.Printf("fichier %s a crypter", fichierCompresse)
					_, err := cryptFile(fichierCompresse, fichierCrypte3, global)
					if err != nil {
						return err
					}
					// calcul du hash du fichier fichierCrypte
					err = hashFiles.ConstruitHash(fichierCrypte3)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func initialisationCryptage(global config.BackupGlobal) error {

	if global.TypeCryptage == config.CryptGpg {
		program := filepath.Dir(global.RepGpg) + "/gpg-connect-agent"
		args := []string{"-v"}

		log.Printf("initialisation agent gpg ...")

		err := execution.Execution(program, args)

		log.Printf("initialisation agent gpg terminé")

		if err != nil {
			return err
		}
	}

	return nil
}

func cryptFile(fileCompressed string, fileCrypted string, global config.BackupGlobal) (string, error) {
	var program string
	var args []string

	if global.TypeCryptage == config.CryptGpg {

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
	} else {
		program = global.RepAge
		args = []string{"-r", global.AgeRecipien, "-o", fileCrypted,
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
}

func GetExtensions() []string {
	return []string{extension, extension2}
}
