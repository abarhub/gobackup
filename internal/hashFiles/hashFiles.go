package hashFiles

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

func ConstruitHash(fichier string) error {
	hashFichier := fichier + ".sha256sum"
	if _, err := os.Stat(hashFichier); errors.Is(err, os.ErrNotExist) {
		hash := sha256.New()
		hashHexa, err := calculHash(fichier, hash)
		if err != nil {
			return fmt.Errorf("erreur pour calculer le hash du fichier %s: %v", fichier, err)
		}
		f, err := os.Create(hashFichier)
		if err != nil {
			return fmt.Errorf("erreur pour creer le fichier %s: %v", hashFichier, err)
		}
		defer f.Close()
		_, err = f.WriteString(hashHexa)
		if err != nil {
			return fmt.Errorf("erreur pour écrire dans le fichier %s: %v", hashFichier, err)
		}
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
