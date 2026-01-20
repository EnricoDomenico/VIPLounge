// VIP Lounge - Backend Configuration

let BACKEND_URL = null;

// Carregar configuração do backend
async function loadBackendConfig() {
  try {
    const response = await fetch('/backend-config.json');
    const config = await response.json();
    BACKEND_URL = config.backendUrl;
    console.log(`✅ Backend configurado: ${BACKEND_URL}`);
  } catch (error) {
    console.error('❌ Erro ao carregar backend-config.json:', error);
    // Fallback para localhost
    BACKEND_URL = window.location.hostname === 'localhost' 
      ? 'http://localhost:8080' 
      : 'https://viplounge-service-dn8vwf3nrq-uc.a.run.app';
    console.warn(`⚠️  Usando backend fallback: ${BACKEND_URL}`);
  }
}

// Função para fazer chamadas ao backend
async function callBackendAPI(endpoint, options = {}) {
  if (!BACKEND_URL) {
    console.warn('⚠️  Backend ainda não configurado, aguardando...');
    await loadBackendConfig();
  }

  const url = `${BACKEND_URL}/v1/${endpoint}`;

  console.log(`📡 Chamando: ${url}`);

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      },
      mode: 'cors'
    });

    console.log(`📊 Status: ${response.status}, Content-Type: ${response.headers.get('content-type')}`);

    // Verificar se response é JSON
    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      const text = await response.text();
      console.error(`❌ Resposta não é JSON. Tipo: ${contentType}`);
      console.error(`Conteúdo: ${text.substring(0, 200)}`);
      throw new Error(`Resposta inválida: ${contentType}. Esperado application/json`);
    }

    if (!response.ok) {
      const errorData = await response.json();
      console.error(`❌ Erro ${response.status}:`, errorData);
      throw new Error(`API Error: ${response.status}`);
    }

    const data = await response.json();
    console.log(`✅ Resposta:`, data);
    return data;
  } catch (error) {
    console.error(`❌ Erro ao chamar API: ${error.message}`);
    throw error;
  }
}

// Inicializar no carregamento
loadBackendConfig().then(() => {
  console.log(`🔌 Backend pronto: ${BACKEND_URL}`);
});

// Exportar para uso global
window.callBackendAPI = callBackendAPI;
window.BACKEND_URL = () => BACKEND_URL;
