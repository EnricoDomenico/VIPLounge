// VIP Lounge - Backend Configuration
// Configure your backend URL here

const API_CONFIG = {
  // Desenvolvimento local
  development: {
    BASE_URL: 'http://localhost:8080',
    API_VERSION: 'v1'
  },

  // Staging (Firebase Hosting - roteia para Cloud Run)
  staging: {
    BASE_URL: window.location.origin,
    API_VERSION: 'v1'
  },

  // Produção (Firebase Hosting - roteia para Cloud Run via rewrite)
  production: {
    BASE_URL: window.location.origin,
    API_VERSION: 'v1'
  }
};

// Detectar ambiente
const ENV = (() => {
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return 'development';
  }
  if (window.location.hostname.includes('staging')) {
    return 'staging';
  }
  return 'production';
})();

const API = API_CONFIG[ENV];

// Função para fazer chamadas ao backend
async function callBackendAPI(endpoint, options = {}) {
  const url = `${API.BASE_URL}/api/${API.API_VERSION}/${endpoint}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers
    },
    credentials: 'include' // Para CORS com cookies se necessário
  });

  if (!response.ok) {
    console.error(`API Error: ${response.status} ${response.statusText} - ${url}`);
    throw new Error(`API Error: ${response.status} ${response.statusText}`);
  }

  return await response.json();
}

// Exportar para uso global
window.API_CONFIG = API;
window.callBackendAPI = callBackendAPI;

console.log(`🔌 Backend conectado: ${API.BASE_URL} [${ENV}]`);
