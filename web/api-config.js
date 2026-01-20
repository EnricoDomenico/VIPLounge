// VIP Lounge - Backend Configuration

let API = null;

// Função para descobrir URL do backend
async function initializeBackend() {
  const environment = (() => {
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
      return 'development';
    }
    return 'production';
  })();

  // URLs candidatas
  const candidates = [];
  
  if (environment === 'development') {
    candidates.push('http://localhost:8080');
  } else {
    // Em produção, tentar descobrir a URL do Cloud Run
    // Primeiro tenta chamar um endpoint de health local
    candidates.push(`${window.location.origin}/api`);
    
    // Se tiver uma variável global com a URL do backend, usar
    if (window.__BACKEND_URL__) {
      candidates.unshift(window.__BACKEND_URL__);
    }
  }

  // Tentar cada candidata
  for (const baseUrl of candidates) {
    try {
      const response = await fetch(`${baseUrl}/api/v1/health`, { 
        method: 'GET',
        mode: 'cors'
      });
      
      if (response.ok) {
        console.log(`✅ Backend encontrado: ${baseUrl}`);
        API = {
          BASE_URL: baseUrl,
          API_VERSION: 'v1'
        };
        window.API_CONFIG = API;
        return;
      }
    } catch (e) {
      console.log(`⚠️  Candidata falhou: ${baseUrl}`);
    }
  }

  // Se nenhuma funcionou, usar a padrão
  const defaultUrl = environment === 'development' 
    ? 'http://localhost:8080'
    : 'https://viplounge-service-dn8vwf3nrq-uc.a.run.app';
  
  console.warn(`⚠️  Usando backend padrão: ${defaultUrl}`);
  API = {
    BASE_URL: defaultUrl,
    API_VERSION: 'v1'
  };
  window.API_CONFIG = API;
}

// Função para fazer chamadas ao backend
async function callBackendAPI(endpoint, options = {}) {
  if (!API) {
    throw new Error('Backend não inicializado. Aguarde...');
  }

  const url = `${API.BASE_URL}/api/${API.API_VERSION}/${endpoint}`;

  console.log(`📡 Chamando: ${url}`);

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      },
      credentials: 'include',
      mode: 'cors'
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error(`❌ Erro ${response.status}: ${errorText}`);
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
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
initializeBackend().then(() => {
  console.log(`🔌 Backend pronto: ${API.BASE_URL}`);
});

// Exportar para uso global
window.callBackendAPI = callBackendAPI;
