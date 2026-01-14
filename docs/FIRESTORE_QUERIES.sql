// Firestore Queries para Manutencao do VIP Lounge
// Use estas queries no Console do Firestore para monitorar e fazer manutencao

// ============================================================================
// 1. LEADS COM ERRO NA INTEGRACAO REDE PARCERIAS
// ============================================================================
// Retorna todos os leads com falha na integracao da Rede Parcerias
// Use quando: receber reclamacoes sobre cadastros nao sincronizados

db.collection('leads')
  .where('rede_parcerias_status', '==', 'FAILED')
  .orderBy('updated_at', 'desc')
  .limit(100)

// Resultado esperado:
// {
//   "cpf": "00933733844",
//   "condo_id": "13",
//   "name": "Ailton Geraldo Junior",
//   "rede_parcerias_status": "FAILED",
//   "rede_parcerias_error": "HTTP_500",
//   "rede_parcerias_attempts": 1,
//   "updated_at": "2026-01-14T04:30:00Z"
// }

// ============================================================================
// 2. LEADS COM RETRY PENDENTE (menos de 3 tentativas)
// ============================================================================
// Ideal para um job de retry automatico

db.collection('partner_integration')
  .where('status', '==', 'RETRY_PENDING')
  .where('attempts', '<', 3)
  .orderBy('next_retry', 'asc')
  .limit(50)

// Resultado:
// {
//   "lead_id": "13_00933733844",
//   "partner": "REDE_PARCERIAS",
//   "status": "RETRY_PENDING",
//   "attempts": 1,
//   "next_retry": "2026-01-14T05:00:00Z",
//   "error_code": "NETWORK_ERROR"
// }

// ============================================================================
// 3. ERROS NOS ULTIMOS 24H
// ============================================================================
// Usado para alertas e dashboard

db.collection('audit_logs')
  .where('status', '==', 'ERROR')
  .where('timestamp', '>', new Date(Date.now() - 86400000))
  .orderBy('timestamp', 'desc')
  .limit(100)

// ============================================================================
// 4. TAXA DE SUCESSO POR CONDOMINIO
// ============================================================================
// Mapear sucesso de cada condominio
// Nota: Firestore nao tem group_by nativo, use BigQuery para analytics

db.collection('leads')
  .where('status', '==', 'APPROVED')
  .orderBy('condo_id')

// SQL para BigQuery (apos exportar dados):
// SELECT 
//   condo_id,
//   COUNT(*) as total_leads,
//   SUM(CASE WHEN rede_parcerias_status = 'REGISTERED' THEN 1 ELSE 0 END) as registered,
//   ROUND(100.0 * SUM(CASE WHEN rede_parcerias_status = 'REGISTERED' THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
// FROM `project.dataset.leads`
// WHERE status = 'APPROVED'
// GROUP BY condo_id
// ORDER BY success_rate DESC

// ============================================================================
// 5. LEADS REJEITADOS (CPF NAO ENCONTRADO)
// ============================================================================

db.collection('leads')
  .where('status', '==', 'REJECTED')
  .orderBy('created_at', 'desc')
  .limit(100)

// ============================================================================
// 6. PERFORMANCE - Tempo medio de resposta por dia
// ============================================================================

db.collection('leads')
  .where('created_at', '>', new Date(Date.now() - 86400000))
  .orderBy('created_at', 'desc')

// Depois processar em Python/Node.js:
// leads = db.collection('leads').where('created_at', '>', yesterday).stream()
// for doc in leads:
//   total_time = doc['superlogica_response_ms'] + doc['rede_parcerias_response_ms']
//   print(f"Lead: {doc['cpf']}, Total: {total_time}ms")

// ============================================================================
// 7. AUDIT - ULTIMAS ACOES DE UM LEAD ESPECIFICO
// ============================================================================

db.collection('audit_logs')
  .where('lead_id', '==', '13_00933733844')
  .orderBy('timestamp', 'desc')

// Resultado mostra timeline completa:
// [
//   { "action": "VALIDATION_COMPLETED", "timestamp": "2026-01-14T04:30:02Z" },
//   { "action": "REGISTRATION_COMPLETED", "timestamp": "2026-01-14T04:30:01Z" },
//   { "action": "SUPERLOGICA_SUCCESS", "timestamp": "2026-01-14T04:30:00Z" },
//   { "action": "VALIDATION_STARTED", "timestamp": "2026-01-14T04:30:00Z" }
// ]

// ============================================================================
// 8. IDENTIFICAR CONDOMINIOS COM PROBLEMAS
// ============================================================================

db.collection('leads')
  .where('status', '==', 'APPROVED')
  .where('rede_parcerias_status', '!=', 'REGISTERED')
  .orderBy('condo_id')

// Se retornar muitos resultados de um condo especifico, pode indicar:
// - Problema na API da Rede Parcerias para aquele condo
// - Email invalido ou duplicado para aquele condo
// - CPF malformado

// ============================================================================
// 9. VERIFICAR DUPLICATAS (MESMO CPF + CONDO)
// ============================================================================

db.collection('leads')
  .orderBy('cpf')

// Depois verificar em codigo:
// if document.id = "{condo_id}_{cpf}" entao nao ha duplicatas
// Se houver 2 leads com mesmo CPF/Condo, eh um problema

// ============================================================================
// 10. ALERTAS RECOMENDADOS (IMPLEMENTAR COM CLOUD MONITORING)
// ============================================================================

// Alert 1: Taxa de erro > 5% em 10 minutos
SELECT
  COUNT(*) as total,
  SUM(CASE WHEN rede_parcerias_status = 'FAILED' THEN 1 ELSE 0 END) as failed,
  ROUND(100.0 * SUM(CASE WHEN rede_parcerias_status = 'FAILED' THEN 1 ELSE 0 END) / COUNT(*), 2) as error_rate
FROM `project.dataset.leads`
WHERE created_at > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 10 MINUTE)
HAVING error_rate > 5

// Alert 2: Sem validacoes por 30 minutos
SELECT COUNT(*) as validations_last_30min
FROM `project.dataset.audit_logs`
WHERE action = 'VALIDATION_STARTED'
  AND timestamp > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 MINUTE)
HAVING COUNT(*) = 0

// Alert 3: Tempo medio Superlogica > 1s
SELECT
  AVG(superlogica_response_ms) as avg_response_ms
FROM `project.dataset.leads`
WHERE created_at > TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 HOUR)
HAVING AVG(superlogica_response_ms) > 1000

// ============================================================================
// INDICES RECOMENDADOS NO FIRESTORE
// ============================================================================

// Criar indices compostos para as queries acima:

// Index 1: Para buscar errors recentes
Collection: leads
Fields: rede_parcerias_status (Ascending), updated_at (Descending)

// Index 2: Para retry queue
Collection: partner_integration
Fields: status (Ascending), attempts (Ascending), next_retry (Ascending)

// Index 3: Para audit logs
Collection: audit_logs
Fields: status (Ascending), timestamp (Descending)

// Index 4: Para performance por condo
Collection: leads
Fields: status (Ascending), condo_id (Ascending), created_at (Descending)

// ============================================================================
// SCRIPT PYTHON PARA MANUTENCAO MANUAL
// ============================================================================

# Para reprocessar leads com falha:
from firebase_admin import firestore, initialize_app

db = firestore.client()

def retry_failed_leads():
    docs = db.collection('leads') \
        .where('rede_parcerias_status', '==', 'FAILED') \
        .stream()
    
    for doc in docs:
        lead = doc.to_dict()
        if lead['rede_parcerias_attempts'] < 3:
            # Tentar novo cadastro aqui
            print(f"Reprocessing: {lead['cpf']}")
            # call RegisterUser again

if __name__ == '__main__':
    initialize_app()
    retry_failed_leads()
