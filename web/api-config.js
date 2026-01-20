// VIP Lounge - Backend Configuration

let API = null;

// Função para fazer chamadas ao backend
async function callBackendAPI(endpoint, options = {}) {
  const url = `/api/v1/${endpoint}`;

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
console.log(`🔌 Backend pronto: /api/v1/`);

// Exportar para uso global
window.callBackendAPI = callBackendAPI;
