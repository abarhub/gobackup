package hashFiles

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"gobackup/internal/config"
	"hash"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const extension = ".sha256sum"

var hashAlgo = sha256.New()

func ConstruitHash(fichier string) error {
	hashFichier := fichier + extension
	if _, err := os.Stat(hashFichier); errors.Is(err, os.ErrNotExist) {
		hashHexa, err := calculHash(fichier, hashAlgo)
		if err != nil {
			return fmt.Errorf("erreur pour calculer le hash du fichier %s: %v", fichier, err)
		}
		f, err := os.Create(hashFichier)
		if err != nil {
			return fmt.Errorf("erreur pour creer le fichier %s: %v", hashFichier, err)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Panicf("%v", err)
			}
		}(f)
		_, err = f.WriteString(hashHexa)
		if err != nil {
			return fmt.Errorf("erreur pour écrire dans le fichier %s: %v", hashFichier, err)
		}
		log.Printf("Ecriture du fichier %s", hashFichier)
		return nil
	} else if err != nil {
		return fmt.Errorf("erreur pour le fichier %s: %v", hashFichier, err)
	} else {
		return nil
	}
}

func calculHash(fichier string, hasher hash.Hash) (string, error) {
	hashHex := ""
	inFile, err := os.Open(fichier)
	if err != nil {
		return hashHex, fmt.Errorf("can't hash %s: %v", fichier, err)
	}
	var ret error
	defer func() {
		err = inFile.Close()
		if err != nil {
			ret = fmt.Errorf("error closing %s: %v", fichier, err)
		}
	}()
	hasher.Reset()
	_, err = io.Copy(hasher, inFile)
	if err != nil {
		return hashHex, fmt.Errorf("can't hash %s: %v", fichier, err)
	}
	sum := hasher.Sum(nil)
	hashHex = fmt.Sprintf("%x", sum)
	return hashHex, ret
}

func VerifieHash(nom string, config config.BackupGlobal) error {

	err := verifieFichiers(config.RepCompression + "/" + nom)
	if err != nil {
		return err
	}

	err = verifieFichiers(config.RepCryptage + "/" + nom)
	if err != nil {
		return err
	}
	return nil
}

func verifieFichiers(repertoire string) error {

	extensionHash := extension
	nbErreurs := 0
	nbFichiers := 0
	nbFichiersIgnores := 0

	log.Printf("vérification du hash %s\n", filepath.Base(repertoire))

	err := filepath.WalkDir(repertoire, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			rep := repertoire
			if d != nil {
				rep = repertoire + "/" + d.Name()
			}
			return fmt.Errorf("erreur pour parcourit les fiuchiers pour le répertoire %s (%s,%s) : %w", rep, repertoire, path, err)

		}
		if !d.IsDir() && !strings.HasSuffix(d.Name(), extensionHash) {
			fileHash := repertoire + "/" + d.Name() + extensionHash
			if _, err := os.Stat(fileHash); err == nil {
				nbFichiers++
				file := repertoire + "/" + d.Name()
				hashHexa, err := calculHash(file, hashAlgo)
				if err != nil {
					return fmt.Errorf("erreur pour calculer le hash du fichier %s : %w", file, err)
				}
				fileHashHandler, err := os.ReadFile(fileHash)
				if err != nil {
					return fmt.Errorf("erreur pour lire le fichier %s : %w", fileHash, err)
				}
				fileHashHandlerStr := string(fileHashHandler)
				if hashHexa == fileHashHandlerStr {
					log.Printf("%s : OK", d.Name())
				} else {
					nbErreurs++
					log.Printf("%s : ERROR", d.Name())
				}
			} else if err != nil && errors.Is(err, os.ErrNotExist) {
				nbFichiersIgnores++
				log.Printf("le fichier %s n'existe pas", fileHash)
			} else {
				nbFichiersIgnores++
				log.Printf("erreur pour accéder au fichier %s : %v", fileHash, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if nbErreurs > 0 {
		return fmt.Errorf("verifieFichiers found %d erreurs", nbErreurs)
	} else {
		log.Printf("aucune erreur dans les %d fichiers testés (%d ignorés)", nbFichiers, nbFichiersIgnores)
		return nil
	}
}

func GetExtension() string {
	return extension
}
