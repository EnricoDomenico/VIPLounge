package repository

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/viplounge/platform/internal/domain"
)

type FirestoreRepository struct {
	client *firestore.Client
	collectionName string
}

func NewFirestoreRepository(ctx context.Context, projectID string) (*FirestoreRepository, error) {
	// Se projectID não for passado, tenta pegar do ambiente (comum no Cloud Run)
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return &FirestoreRepository{
		client: client,
		collectionName: "leads",
	}, nil
}

func (r *FirestoreRepository) Save(ctx context.Context, lead domain.Lead) error {
	// Se client é nil, retorna silenciosamente (dev local sem credenciais)
	if r == nil || r.client == nil {
		log.Printf("INFO: Firestore client não disponível, pulando save")
		return nil
	}
	
	// Usa o CPF e CondoID como chave composta ou ID do documento para evitar duplicatas fáceis
	docID := fmt.Sprintf("%s_%s", lead.CondoID, lead.CPF)
	
	// Set (upsert) para criar ou atualizar
	_, err := r.client.Collection(r.collectionName).Doc(docID).Set(ctx, lead)
	if err != nil {
		log.Printf("Erro ao salvar no Firestore: %v", err)
		return err
	}
	
	return nil
}

func (r *FirestoreRepository) Close() error {
	return r.client.Close()
}


