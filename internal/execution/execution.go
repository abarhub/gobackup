package execution

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func Execution(program string, arguments []string) error {
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
