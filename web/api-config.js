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

  // URLs candidatas para o backend
  const candidates = [];
  
  if (environment === 'development') {
    candidates.push('http://localhost:8080');
  } else {
    // Em produção, backend está no Cloud Run
    // Usar uma URL externa conhecida ou deixar vazio se for local
    if (window.__BACKEND_URL__) {
      candidates.push(window.__BACKEND_URL__);
    }
  }

  // Tentar cada candidata
  for (const baseUrl of candidates) {
    try {
      const response = await fetch(`${baseUrl}/api/v1/health`, { 
        method: 'GET',
        mode: 'cors'
      });
      
      if (response.ok && response.headers.get('content-type')?.includes('application/json')) {
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
  // Em produção, sem candidatas, significa que o backend está em um domínio externo
  // ou está como uma cloud function proxy que redireciona
  
  let defaultUrl = environment === 'development' 
    ? 'http://localhost:8080'
    : ''; // Vazio significa mesmo domínio (Firebase relays)
  
  console.warn(`⚠️  Usando backend: ${defaultUrl || 'same-origin'}`);
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

  // Construir URL corretamente
  let url;
  if (API.BASE_URL) {
    url = `${API.BASE_URL}/api/${API.API_VERSION}/${endpoint}`;
  } else {
    // Same-origin (Firebase redireciona)
    url = `/api/${API.API_VERSION}/${endpoint}`;
  }

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

    // Verificar se response é JSON
    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      const text = await response.text();
      console.error(`❌ Resposta não é JSON. Tipo: ${contentType}, Conteúdo: ${text.substring(0, 100)}`);
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
initializeBackend().then(() => {
  console.log(`🔌 Backend pronto: ${API.BASE_URL || 'same-origin'}`);
});

// Exportar para uso global
window.callBackendAPI = callBackendAPI;
