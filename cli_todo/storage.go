package main

import (
	"encoding/json"
	"log"
	"os"
)

type Storage[T any] struct {
	fileName string
}

func NewStorage[T any](fileName string) *Storage[T] {
	return &Storage[T]{fileName: fileName}
}

func (s *Storage[T]) save(data *T) error {
	b, err := json.MarshalIndent(data, "", "")

	if err != nil {
		return err
	}

	return os.WriteFile(s.fileName, b, 0644)
}

func (s *Storage[T]) Load(dst *T) error {
	b, err := os.ReadFile(s.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Starting with empty data", s.fileName)
			return nil
		}
		return err
	}

	if len(b) == 0 {
		log.Printf("Starting with empty data", s.fileName)
		return nil
	}

	return json.Unmarshal(b, dst)
}
