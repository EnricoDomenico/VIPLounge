// VIP Lounge - Backend Configuration

let BACKEND_URL = null;

// Detectar ambiente
function getEnvironment() {
  const hostname = window.location.hostname;
  return (hostname === 'localhost' || hostname === '127.0.0.1') ? 'development' : 'production';
}

// Carregar configuração do backend
async function loadBackendConfig() {
  const env = getEnvironment();
  
  if (env === 'development') {
    // Localhost - chamar backend local
    BACKEND_URL = 'http://localhost:8081';
    console.log(`✅ Ambiente: DESENVOLVIMENTO`);
    console.log(`✅ Backend: ${BACKEND_URL}`);
    return;
  }

  // Produção - tentar carregar configuração
  try {
    const response = await fetch('/backend-config.json');
    const config = await response.json();
    BACKEND_URL = config.backendUrl;
    console.log(`✅ Ambiente: PRODUÇÃO`);
    console.log(`✅ Backend: ${BACKEND_URL}`);
  } catch (error) {
    console.error('❌ Erro ao carregar backend-config.json:', error);
    // Fallback - usar mesma origem (Firebase rewrite)
    BACKEND_URL = window.location.origin;
    console.warn(`⚠️  Usando backend fallback (same-origin): ${BACKEND_URL}`);
  }
}

// Função para fazer chamadas ao backend
async function callBackendAPI(endpoint, options = {}) {
  if (!BACKEND_URL) {
    console.warn('⚠️  Backend ainda não configurado, inicializando...');
    await loadBackendConfig();
  }

  const url = `${BACKEND_URL}/v1/${endpoint}`;

  console.log(`📡 Chamando: ${url}`);
  console.log(`📦 Método: ${options.method || 'GET'}`);
  if (options.body) {
    console.log(`📦 Body: ${options.body}`);
  }

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      },
      mode: 'cors',
      credentials: 'omit'
    });

    console.log(`📊 Status: ${response.status} ${response.statusText}`);
    console.log(`📄 Content-Type: ${response.headers.get('content-type')}`);

    // Verificar se response é JSON
    const contentType = response.headers.get('content-type');
    
    if (!response.ok) {
      // Tentar ler como texto para ver o erro
      const errorText = await response.text();
      console.error(`❌ Erro HTTP ${response.status}`);
      console.error(`Resposta: ${errorText.substring(0, 200)}`);
      throw new Error(`HTTP ${response.status}: ${errorText.substring(0, 100)}`);
    }

    if (!contentType || !contentType.includes('application/json')) {
      const text = await response.text();
      console.error(`❌ Resposta não é JSON. Tipo: ${contentType}`);
      console.error(`Conteúdo: ${text.substring(0, 200)}`);
      throw new Error(`Resposta inválida: ${contentType}. Esperado application/json`);
    }

    const data = await response.json();
    console.log(`✅ Resposta recebida:`, data);
    return data;
  } catch (error) {
    console.error(`❌ Erro ao chamar API: ${error.message}`);
    console.error(`Stack:`, error);
    throw error;
  }
}

// Inicializar no carregamento
loadBackendConfig().then(() => {
  console.log(`🔌 Backend pronto: ${BACKEND_URL}`);
  console.log(`🌍 Ambiente: ${getEnvironment()}`);
});

// Exportar para uso global
window.callBackendAPI = callBackendAPI;
window.BACKEND_URL = () => BACKEND_URL;
window.getEnvironment = getEnvironment;
