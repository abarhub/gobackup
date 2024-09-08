package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mxk/go-vss"
)

type backupGlobal struct {
	listeNomBackup      []string
	rep7zip             string
	repGpg              string
	repCompression      string
	repCryptage         string
	listeBackup         []backup
	dateHeure           string
	nbBackupIncremental int
	recipient           string
	activeVss           bool
	lettreVss           map[string]string
	logDir              string
}

type backup struct {
	nom       string
	rep       []string
	set       map[string]bool
	map2      map[string][]string
	fileListe string
}

type listeFichiers struct {
	listeFiles string
	nbFiles    int
}

// Execution States
const (
	EsSystemRequired = 0x00000001
	EsContinuous     = 0x80000000
)

var pulseTime = 10 * time.Second

func testEqSuffixSlice(suffix, tab []string) bool {
	if len(suffix) > len(tab) {
		return false
	}
	for i := len(suffix) - 1; i >= 0; i-- {
		if suffix[i] != tab[i] {
			return false
		}
	}

	return true
}

func parcourt(res backup, complet bool, date time.Time, configGlobal backupGlobal) (int, error) {
	nbFichier := 0
	f, err := os.OpenFile(res.fileListe, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return 0, err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)

	for i := range res.rep {
		root0 := res.rep[i]
		root := convertie(root0, configGlobal)
		log.Printf("Parcourt de %q (%v)\n", root0, root)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Erreur d'accès à %q: %v\n", path, err)
				return err
			}
			fileName := filepath.Base(path)

			_, ok := res.set[fileName]
			if ok {
				fmt.Printf("Répertoire ignoré: %q\n", path)
				return filepath.SkipDir
			}

			_, ok2 := res.map2[fileName]
			if ok2 {
				tab := strings.Split(path, "\\")
				if testEqSuffixSlice(res.map2[fileName], tab) {
					fmt.Printf("Répertoire ignoré: %q\n", path)
					return filepath.SkipDir
				}
			}

			if !info.IsDir() {

				traitement := false

				if complet {
					traitement = true
				} else {
					if info.ModTime().After(date) {
						traitement = true
					}
				}

				if traitement {
					nbFichier++
					if _, err = f.WriteString(fmt.Sprintf("%s\n", path)); err != nil {
						return err
					}
				}
			}

			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("Erreur lors du parcours : %v\n", err)
		}
	}

	return nbFichier, nil
}

func convertie(root string, global backupGlobal) string {
	if len(root) >= 2 && root[1] == ':' && len(global.lettreVss) > 0 {
		lettre := root[0]
		if link, ok := global.lettreVss[string(rune(lettre))]; ok {
			root2 := link + root[2:]
			return root2
		}
	}
	return root
}

func addMap(map2 *map[string][]string, s string) {
	tab := strings.Split(s, "\\")
	(*map2)[tab[len(tab)-1]] = tab
}

func initialisationConfig(filename string) (backupGlobal, error) {
	var res = backupGlobal{}

	file, err := os.Open(filename)
	if err != nil {
		return backupGlobal{}, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	mapConfig := map[string]string{}
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// on ignore les commentaires commençant par #
		} else {
			i := strings.IndexRune(line, '=')
			if i >= 0 {
				mapConfig[line[:i]] = line[i+1:]
			}
		}
	}

	rep7zip, ok := mapConfig["global.rep_7zip"]
	if ok {
		res.rep7zip = strings.TrimSpace(rep7zip)
	}
	repgpg, ok := mapConfig["global.rep_gpg"]
	if ok {
		res.repGpg = strings.TrimSpace(repgpg)
	}
	repCompression, ok := mapConfig["global.rep_compression"]
	if ok {
		res.repCompression = strings.TrimSpace(repCompression)
	}
	repCryptage, ok := mapConfig["global.rep_cryptage"]
	if ok {
		res.repCryptage = strings.TrimSpace(repCryptage)
	}
	nbBackupIncremental, ok := mapConfig["global.nb_backup_incremental"]
	if ok {
		nbBackupIncremental = strings.TrimSpace(nbBackupIncremental)
		if len(nbBackupIncremental) > 0 {
			res.nbBackupIncremental, err = strconv.Atoi(nbBackupIncremental)
			if err != nil {
				return backupGlobal{}, fmt.Errorf("le paramètre global.nb_backup_incremental n'est pas un nombre", err)
			}
		} else {
			res.nbBackupIncremental = 0
		}
	}
	recipient, ok := mapConfig["global.recipient"]
	if ok {
		res.recipient = strings.TrimSpace(recipient)
	}
	activeVss, ok := mapConfig["global.activeVss"]
	if ok {
		res.activeVss = strings.TrimSpace(activeVss) == "true"
	}
	logdir, ok := mapConfig["global.logdir"]
	if ok {
		res.logDir = strings.TrimSpace(logdir)
	}

	res.dateHeure = strings.ReplaceAll(time.Now().Format("20060102_150405.000"), ".", "")

	listeBackup, ok := mapConfig["global.liste_backups"]
	if ok {

		liste := strings.Split(listeBackup, ",")

		for _, v := range liste {

			var res2 = backup{}
			res2.nom = v
			debut := "backup." + v
			key := debut + ".rep_a_sauver"
			if aSauver, ok := mapConfig[key]; ok {
				tab := strings.Split(aSauver, ",")
				res2.rep = tab
			}
			key = debut + ".rep_nom_a_ignorer"
			if repNomAIgnorer, ok := mapConfig[key]; ok {
				tab := strings.Split(repNomAIgnorer, ",")
				set := map[string]bool{}
				for _, v := range tab {
					set[v] = true
				}
				res2.set = set
			}
			key = debut + ".rep_a_ignorer"
			if repAIgnorer, ok := mapConfig[key]; ok {
				tab := strings.Split(repAIgnorer, ",")
				map2 := map[string][]string{}
				for _, v := range tab {
					addMap(&map2, v)
				}
				res2.map2 = map2
			}

			fileTemp, err := createTempFile("listeFichiers_" + res2.nom)
			if err != nil {
				return backupGlobal{}, fmt.Errorf("erreur pour creer le fichier temporaire : %v", err)
			}
			res2.fileListe = fileTemp

			res.listeBackup = append(res.listeBackup, res2)
		}

	}

	if err := scanner.Err(); err != nil {
		return backupGlobal{}, err
	}

	if len(res.listeBackup) == 0 {
		return backupGlobal{}, errors.New("no liste backup")
	}

	if len(res.rep7zip) == 0 {
		return backupGlobal{}, errors.New("no 7zip directory")
	}

	if len(res.repGpg) == 0 {
		return backupGlobal{}, errors.New("no gpg directory")
	}

	if len(res.repCompression) == 0 {
		return backupGlobal{}, errors.New("no compress directory")
	}

	if len(res.repCryptage) == 0 {
		return backupGlobal{}, errors.New("no crypt directory")
	}

	if res.nbBackupIncremental < 0 {
		return backupGlobal{}, errors.New("nbBackupIncremental doit être superieur ou égal à 0")
	}

	if len(res.recipient) == 0 {
		return backupGlobal{}, errors.New("le paramètre recipient est vide")
	}

	return res, nil
}

func createTempFile(name string) (string, error) {
	f, err := os.CreateTemp("", name)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)
	if err != nil {
		return "", fmt.Errorf("erreur pour creer le fichier temporaire : %v", err)
	} else {
		return f.Name(), nil
	}
}

func listeFiles(backup backup, complet bool, date time.Time, global backupGlobal) (listeFichiers, error) {

	log.Printf("ecriture de la liste des fichiers dans  %s (complet=%v) ...\n", backup.fileListe, complet)

	start := time.Now()

	nbFichiers, err := parcourt(backup, complet, date, global)
	if err != nil {
		return listeFichiers{}, err
	}

	elapsed := time.Since(start)

	log.Printf("parcourt %s", elapsed)

	return listeFichiers{backup.fileListe, nbFichiers}, nil
}

func pasSleep() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setThreadExecStateProc := kernel32.NewProc("SetThreadExecutionState")

	pulse := time.NewTicker(pulseTime)

	log.Println("Starting keep alive poll... (silence)")
	for {
		select {
		case <-pulse.C:
			_, _, err := setThreadExecStateProc.Call(uintptr(EsSystemRequired))
			if err != nil {
				s := fmt.Sprintf("%v", err)
				if s != "L’opération a réussi." {
					log.Printf("Erreur pour changer l'état de veille: %v\n", err)
				}
			}
		}
	}
}

func main() {
	var configGlobal backupGlobal
	var configFile string
	var err error

	// Capture le temps de début
	startTime := time.Now()

	args := os.Args

	if len(args) > 1 {
		configFile = args[1]
	} else {
		log.Fatal("Le fichier de config n'est pas indiqué")
	}

	go pasSleep()

	configGlobal, err = initialisationConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if len(configGlobal.logDir) > 0 {
		logFile, err := os.OpenFile(configGlobal.logDir+fmt.Sprintf("/app-%s.log", time.Now().Format("20060102")), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.Fatal(err)
		}
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
		defer func(logFile *os.File) {
			err := logFile.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(logFile)
	}

	log.Printf("current pid: %v", os.Getpid())

	if configGlobal.activeVss {
		err = initVss(&configGlobal)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("mapVss apres init: %v", configGlobal.lettreVss)

		configGlobalCopy := configGlobal
		defer func(global backupGlobal) {
			log.Printf("mapVss avant fermeture: %v", configGlobalCopy.lettreVss)
			err := fermeVss(global)
			if err != nil {
				log.Fatal(err)
			}
		}(configGlobalCopy)
	}

	for _, backup := range configGlobal.listeBackup {

		log.Printf("traitement de %v", backup.nom)

		fileCompressed, err := compress(backup, configGlobal)
		if err != nil {
			log.Fatal(err)
		}

		if len(fileCompressed) > 0 {
			err = crypt(fileCompressed, backup, configGlobal)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Capture le temps de fin
	endTime := time.Now()

	// Calcul de la durée écoulée
	duration := endTime.Sub(startTime)

	// Affichage de la durée écoulée
	log.Printf("Duree totale = %v\n", duration)
}

func fermeVss(global backupGlobal) error {

	log.Printf("VSS à supprimer : %v", global.lettreVss)

	for lettre, link := range global.lettreVss {
		log.Printf("Suppression de %v (%v) ...", lettre, link)
		err := vss.Remove(link)
		if err != nil {
			return fmt.Errorf("erreur pour supprimer %v : %v", lettre, err)
		}
		log.Printf("Suppression de %v (%v) OK", lettre, link)
	}
	return nil
}

func initVss(configGlobal *backupGlobal) error {

	var listeDisque []string

	log.Printf("initialisation de VSS ...")

	for _, b := range configGlobal.listeBackup {
		for _, p := range b.rep {
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

	configGlobal.lettreVss = mapVss

	log.Printf("map VSS : %v", configGlobal.lettreVss)

	log.Printf("initialisation de VSS ok")

	return nil
}

func crypt(fileCompressed string, b backup, global backupGlobal) error {

	rep := path.Dir(fileCompressed)
	filename := path.Base(fileCompressed)
	repCrypt := fmt.Sprintf("%v/%v", global.repCryptage, b.nom)
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

func cryptFile(fileCompressed string, fileCrypted string, global backupGlobal) (string, error) {
	var program string
	var args []string

	program = global.repGpg
	args = []string{"-v", "--encrypt", "--recipient=" + global.recipient, "--output=" + fileCrypted,
		fileCompressed}

	log.Printf("cryptage de %v ...", path.Base(fileCompressed))

	err := execution(program, args)

	log.Printf("cryptage terminé")

	if err != nil {
		return "", err
	} else {
		return fileCrypted, nil
	}
}

func compress(backup backup, global backupGlobal) (string, error) {

	var res string

	repCompression := fmt.Sprintf("%v/%v", global.repCompression, backup.nom)
	err := os.MkdirAll(repCompression, os.ModePerm)
	if err != nil {
		return "", err
	}

	complet, date, err := calculComplet(repCompression, backup, global)
	if err != nil {
		return "", err
	}

	listeFichiers, err := listeFiles(backup, complet, date, global)
	if err != nil {
		return "", err
	}

	if listeFichiers.nbFiles == 0 {
		log.Printf("Aucun fichier à sauvegarder")
		return "", nil
	} else {
		log.Printf("%d fichiers à sauvegarder", listeFichiers.nbFiles)
		res, err = compression(backup, global, listeFichiers.listeFiles, repCompression, complet, date)
		if err != nil {
			return "", fmt.Errorf("erreur pour compresser le fichier %s (%s) : %v", backup.nom, listeFichiers.listeFiles, err)
		} else {
			return res, nil
		}
	}
}

func calculComplet(repCompression string, backup backup, global backupGlobal) (bool, time.Time, error) {
	files, err := os.ReadDir(repCompression)
	if err != nil {
		return false, time.Time{}, err
	}
	var liste []string

	debutComplet := fmt.Sprintf("backupc_%v_", backup.nom)
	debutIncrement := fmt.Sprintf("backupi_%v_", backup.nom)

	for _, file := range files {
		if !file.IsDir() && (strings.HasPrefix(file.Name(), debutComplet) || strings.HasPrefix(file.Name(), debutIncrement)) {
			s := file.Name()
			if strings.HasSuffix(s, ".gpg") {
				s = strings.TrimSuffix(s, ".gpg")
			}
			var re = regexp.MustCompile(`\.[0-9]+$`)
			s = re.ReplaceAllString(s, ``)
			if strings.HasSuffix(s, ".7z") {
				s = strings.TrimSuffix(s, ".7z")
			}
			re2 := regexp.MustCompile("^(" + debutComplet + ")|(" + debutIncrement + `)[0-9]+_[0-9]+$`)
			if re2.MatchString(s) {
				if !slices.Contains(liste, s) {
					liste = append(liste, s)
				}
			}
		}
	}

	sort.Sort(sort.StringSlice(liste))

	nbBackupIncremental := 0
	var dateDebut time.Time
	var dateDebutTrouve = false
	var t1 time.Time
	var backupComplet bool
	if global.nbBackupIncremental > 0 {
		for i := len(liste) - 1; i >= 0; i-- {
			s := liste[i]
			if strings.HasPrefix(s, "backupc_") {
				break
			} else {
				if !dateDebutTrouve {
					s0 := strings.TrimPrefix(s, debutIncrement)
					if len(s0) == 18 {
						s0 = s0[0:len(s0)-3] + "." + s0[len(s0)-3:]
						tt, err0 := time.Parse("20060102_150405.000", s0)
						if err0 != nil {
							// erreur de parsing => on ignore le fichier
						} else {
							dateDebutTrouve = true
							dateDebut = tt
						}
					}
				}
				nbBackupIncremental++
			}
		}
	}

	log.Printf("date: %v (%v), nbBackupIncr: %d", dateDebut, dateDebutTrouve, nbBackupIncremental)

	if !dateDebutTrouve || nbBackupIncremental > global.nbBackupIncremental {
		backupComplet = true
	} else {
		backupComplet = false
		t1 = time.Date(dateDebut.Year(), dateDebut.Month(), dateDebut.Day(), 0, 0, 0, 0, dateDebut.Location())
	}

	log.Printf("liste %v", liste)
	log.Printf("backup complet: %v date: %v", backupComplet, t1)

	return backupComplet, t1, nil
}

func compression(backup backup, global backupGlobal, fileList string, repCompression string, complet bool, date time.Time) (string, error) {
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
	res = fmt.Sprintf("%v/backup%s_%v_%s.7z", repCompression, c, backup.nom, global.dateHeure)

	program = global.rep7zip
	args = []string{"a", "-t7z", "-spf", "-bt", "-v1g", res, "@" + fileList}

	log.Printf("compression ...")

	err := execution(program, args)

	log.Printf("compression terminé")

	if err != nil {
		return "", err
	} else {
		return res, nil
	}
}

func execution(program string, arguments []string) error {
	var exitCode int

	log.Printf("exec : %s %v", program, arguments)

	// Préparer la commande avec cmd /c start
	cmd := exec.Command("cmd", append([]string{"/c", program}, arguments...)...)

	// Rediriger la sortie standard et la sortie d'erreur vers la console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Capture le temps de début
	startTime := time.Now()

	// Exécuter la commande
	err := cmd.Start() // Démarre la commande sans bloquer
	if err != nil {
		return fmt.Errorf("erreur lors de l'exécution de la commande: %v", err)
	}

	// Attendre la fin du programme
	err = cmd.Wait()
	// Capture le temps de fin
	endTime := time.Now()
	if err != nil {

		// Si une erreur survient, obtenir le code de retour
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		if ok {
			// Récupérer le code de retour du processus
			exitCode = exitError.ExitCode()
		} else {
			// Autre erreur
			return fmt.Errorf("erreur lors de l'attente de la commande : %v", err)
		}

	}

	// Calcul de la durée écoulée
	duration := endTime.Sub(startTime)

	// Affichage de la durée écoulée
	log.Printf("Duree = %v, Code sortie = %v\n", duration, exitCode)
	return err
}
