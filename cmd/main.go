package main

import (
	"fmt"
	"gobackup/internal/compress"
	"gobackup/internal/config"
	"gobackup/internal/crypt"
	"gobackup/internal/hashFiles"
	"gobackup/internal/noSleep"
	"gobackup/internal/vss"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	var configGlobal config.BackupGlobal
	var configFile string
	var err error

	// Capture le temps de début
	startTime := time.Now()

	args := os.Args

	if len(args) > 1 {
		configFile = args[1]
	} else {
		log.Panic("Le fichier de config n'est pas indiqué")
	}

	go noSleep.PasSleep()

	configGlobal, err = config.InitialisationConfig(configFile)
	if err != nil {
		log.Panic(err)
	}

	if len(configGlobal.LogDir) > 0 {
		file := configGlobal.LogDir + fmt.Sprintf("/app-%s.log", time.Now().Format("20060102"))
		logFile, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Panic(fmt.Errorf("erreur pour ouvrir le fichier %s : %w", file, err))
		}
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
		defer func(logFile *os.File) {
			err := logFile.Close()
			if err != nil {
				log.Panic(fmt.Errorf("erreur pour fermer le fichier %s : %w", file, err))
			}
		}(logFile)
	}

	log.Printf("current pid: %v", os.Getpid())

	if configGlobal.ActiveVss {
		err = vss.InitVss(&configGlobal)
		if err != nil {
			log.Panic(fmt.Errorf("erreur pour initialiser VSS : %w", err))
		}
		log.Printf("mapVss apres init: %v", configGlobal.LettreVss)

		configGlobalCopy := configGlobal
		defer func(global config.BackupGlobal) {
			log.Printf("mapVss avant fermeture: %v", configGlobalCopy.LettreVss)
			err := vss.FermeVss(global)
			if err != nil {
				log.Panic(fmt.Errorf("erreur pour fermer VSS: %w", err))
			}
		}(configGlobalCopy)
	}

	// backup
	err = backup(configGlobal)
	if err != nil {
		log.Panic(fmt.Errorf("erreur pour le backup : %w", err))
	}

	// Capture le temps de fin
	endTime := time.Now()

	// Calcul de la durée écoulée
	duration := endTime.Sub(startTime)

	// Affichage de la durée écoulée
	log.Printf("Duree totale = %v\n", duration)
}

func backup(configGlobal config.BackupGlobal) error {
	for _, backup := range configGlobal.ListeBackup {

		log.Printf("traitement de %v", backup.Nom)

		fileCompressed, err := compress.Compress(backup, configGlobal)
		if err != nil {
			return fmt.Errorf("erreur pour compresser les fichiers : %w", err)
		}

		if len(fileCompressed.ListeFichier) > 0 {
			err = crypt.Crypt(fileCompressed, backup, configGlobal)
			if err != nil {
				return fmt.Errorf("erreur pour crypter les fichiers : %w", err)
			}
		}

		err = hashFiles.VerifieHash(backup.Nom, configGlobal)
		if err != nil {
			return fmt.Errorf("erreur pour vérifier les hash : %w", err)
		}
	}

	return nil
}
