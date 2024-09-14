package vss

import (
	"fmt"
	"github.com/mxk/go-vss"
	"gobackup/internal/config"
	"log"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

func FermeVss(global config.BackupGlobal) error {

	log.Printf("VSS à supprimer : %v", global.LettreVss)

	for lettre, link := range global.LettreVss {
		log.Printf("Suppression de %v (%v) ...", lettre, link)
		err := vss.Remove(link)
		if err != nil {
			return fmt.Errorf("erreur pour supprimer %v : %v", lettre, err)
		}
		log.Printf("Suppression de %v (%v) OK", lettre, link)
	}
	return nil
}

func InitVss(configGlobal *config.BackupGlobal) error {

	var listeDisque []string

	log.Printf("initialisation de VSS ...")

	for _, b := range configGlobal.ListeBackup {
		for _, p := range b.Rep {
			p2 := strings.ToUpper(p)
			if len(p) >= 2 && p[1] == ':' && rune(p2[0]) >= 'A' && rune(p2[0]) <= 'Z' {
				if !slices.Contains(listeDisque, p2[0:1]) {
					listeDisque = append(listeDisque, p2[0:1])
				}
			}
		}
	}

	sort.Sort(sort.StringSlice(listeDisque))

	now := time.Now().UnixMilli()

	mapVss := make(map[string]string)

	for i, lettre := range listeDisque {
		lettreStr := lettre + ":"
		link := "c:\\linkgo" + strings.ToLower(lettre) + "_" + strconv.FormatInt(now, 10) + "_" + strconv.FormatInt(int64(i), 10)
		log.Printf("creation du VSS de %v", lettreStr)
		err := vss.CreateLink(link, lettreStr)
		if err != nil {
			return err
		}
		mapVss[lettre] = link
	}

	configGlobal.LettreVss = mapVss

	log.Printf("map VSS : %v", configGlobal.LettreVss)

	log.Printf("initialisation de VSS ok")

	return nil
}
