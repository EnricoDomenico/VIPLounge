/**
 * Import function triggers from their respective submodules:
 *
 * import {onCall} from "firebase-functions/v2/https";
 * import {onDocumentWritten} from "firebase-functions/v2/firestore";
 *
 * See a full list of supported triggers at https://firebase.google.com/docs/functions
 */

import {setGlobalOptions} from "firebase-functions";
import {onRequest} from "firebase-functions/v2/https";

// Configurar opções globais
setGlobalOptions({region: "us-central1", maxInstances: 100});

// URL do backend Cloud Run
const BACKEND_URL = process.env.BACKEND_URL || "https://viplounge-service-dn8vwf3nrq-uc.a.run.app";

/**
 * Proxy Cloud Function - roteia /apiProxy/** para o backend Cloud Run
 * Tratamento de CORS incluído
 */
export const apiProxy = onRequest(
  {cors: true, maxInstances: 100},
  async (req, res) => {
    try {
      const path = req.path.replace(/^\/apiProxy/, ""); // Remove /apiProxy do início
      const backendUrl = `${BACKEND_URL}${path}`;

      console.log(`📡 Proxying: ${req.method} ${backendUrl}`);

      // Preparar headers
      const headers: HeadersInit = {
        "Content-Type": "application/json",
      };

      // Copiar authorization se existir
      if (req.get("authorization")) {
        headers["Authorization"] = req.get("authorization")!;
      }

      // Preparar body
      let body: string | undefined;
      if (req.method !== "GET" && req.method !== "HEAD") {
        body = JSON.stringify(req.body);
      }

      // Fazer requisição para o backend
      const response = await fetch(backendUrl, {
        method: req.method,
        headers,
        body,
      });

      // Obter dados da resposta
      const contentType = response.headers.get("content-type") || "application/json";
      let data;

      if (contentType.includes("application/json")) {
        data = await response.json();
      } else {
        data = await response.text();
      }

      // Copiar headers de resposta
      res.set("Content-Type", contentType);
      if (response.headers.get("cache-control")) {
        res.set("Cache-Control", response.headers.get("cache-control")!);
      }

      // Enviar resposta
      res.status(response.status).send(data);
    } catch (error) {
      console.error("❌ Erro no proxy:", error);
      res.status(500).json({
        error: "Internal Server Error",
        message: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }
);
