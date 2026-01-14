package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/logging"
	"github.com/viplounge/platform/internal/domain"
)

type CloudLogger struct {
	client *logging.Client
	logger *logging.Logger
}

type LogEntry struct {
	Timestamp       string                 `json:"timestamp"`
	Severity        string                 `json:"severity"` // INFO, WARNING, ERROR, CRITICAL
	Service         string                 `json:"service"`
	RequestID       string                 `json:"request_id"`
	LeadID          string                 `json:"lead_id"`
	Event           string                 `json:"event"`
	Details         map[string]interface{} `json:"details"`
	UserIP          string                 `json:"user_ip,omitempty"`
	UserAgent       string                 `json:"user_agent,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
}

// NewCloudLogger inicializa o logger do Google Cloud
func NewCloudLogger(ctx context.Context, projectID string) (*CloudLogger, error) {
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("erro criar cloud logging client: %w", err)
	}

	logger := client.Logger("viplounge-api")

	return &CloudLogger{
		client: client,
		logger: logger,
	}, nil
}

// LogValidationStarted registra o início de uma validação
func (cl *CloudLogger) LogValidationStarted(ctx context.Context, requestID, cpf, condoID string) {
	entry := LogEntry{
		Timestamp: fmt.Sprintf("%v", ctx.Value("timestamp")),
		Severity:  "INFO",
		Service:   "viplounge-api",
		RequestID: requestID,
		LeadID:    fmt.Sprintf("%s_%s", condoID, cpf),
		Event:     "VALIDATION_STARTED",
		Details: map[string]interface{}{
			"cpf":     cpf,
			"condo_id": condoID,
		},
	}
	cl.logEntry(ctx, entry)
}

// LogSuperlogicaSuccess registra sucesso na busca Superlogica
func (cl *CloudLogger) LogSuperlogicaSuccess(ctx context.Context, requestID string, lead *domain.Lead) {
	entry := LogEntry{
		Timestamp: fmt.Sprintf("%v", ctx.Value("timestamp")),
		Severity:  "INFO",
		Service:   "viplounge-api",
		RequestID: requestID,
		LeadID:    fmt.Sprintf("%s_%s", lead.CondoID, lead.CPF),
		Event:     "SUPERLOGICA_SUCCESS",
		Details: map[string]interface{}{
			"name":       lead.Name,
			"email":      lead.Email,
			"response_ms": lead.SuperlogicaResponseMs,
		},
	}
	cl.logEntry(ctx, entry)
}

// LogRegistrationCompleted registra conclusão do cadastro
func (cl *CloudLogger) LogRegistrationCompleted(ctx context.Context, requestID string, lead *domain.Lead) {
	entry := LogEntry{
		Timestamp: fmt.Sprintf("%v", ctx.Value("timestamp")),
		Severity:  "INFO",
		Service:   "viplounge-api",
		RequestID: requestID,
		LeadID:    fmt.Sprintf("%s_%s", lead.CondoID, lead.CPF),
		Event:     "REGISTRATION_COMPLETED",
		Details: map[string]interface{}{
			"status":                    lead.RedeParceriasStatus,
			"rede_parcerias_response_ms": lead.RedeParceriasResponseMs,
			"attempts":                  lead.RedeParceriasAttempts,
		},
	}
	cl.logEntry(ctx, entry)
}

// LogError registra um erro
func (cl *CloudLogger) LogError(ctx context.Context, requestID, leadID, errorMsg string, details map[string]interface{}) {
	entry := LogEntry{
		Timestamp:    fmt.Sprintf("%v", ctx.Value("timestamp")),
		Severity:     "ERROR",
		Service:      "viplounge-api",
		RequestID:    requestID,
		LeadID:       leadID,
		Event:        "ERROR",
		ErrorMessage: errorMsg,
		Details:      details,
	}
	cl.logEntry(ctx, entry)
}

// logEntry faz o log de uma entrada
func (cl *CloudLogger) logEntry(ctx context.Context, entry LogEntry) {
	// Converter para JSON
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("erro serializar log: %v", err)
		return
	}

	// Se o logger do Cloud Logging não estiver disponível, fazer fallback
	if cl.logger != nil {
		cl.logger.Log(logging.Entry{
			Payload:  string(data),
			Severity: logging.Severity(entry.Severity),
		})
	} else {
		// Fallback para stdout
		log.Println(string(data))
	}
}

// Close fecha a conexão com Cloud Logging
func (cl *CloudLogger) Close() error {
	if cl.client != nil {
		return cl.client.Close()
	}
	return nil
}
