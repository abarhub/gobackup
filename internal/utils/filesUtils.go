package utils

import (
	"fmt"
	"log"
	"os"
)

func CreateEmptyFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error for create file %s : %w", filename, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panicf("error for close file %s : %v", filename, err)
		}
	}(f)
	return nil
}

func CreateFile(filename string, contenu string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error for create file %s : %w", filename, err)
	}
	_, err = f.WriteString(contenu)
	if err != nil {
		return fmt.Errorf("error for write file %s : %w", filename, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Panicf("error for close file %s : %v", filename, err)
		}
	}(f)
	return nil
}

func ReadFile(filename string) (string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error for read file %s : %w", filename, err)
	}
	str := string(b)
	return str, nil
}
