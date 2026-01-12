# Local Synapse 🧠

**Local Synapse** es un proxy ligero y servidor de herramientas escrito en Go. Su propósito es conectar un servidor local de [Ollama](https://ollama.com/) con el mundo exterior y exponer métricas del sistema a través del [Model Context Protocol (MCP)](https://modelcontextprotocol.io/).

Este proyecto está diseñado para desarrolladores que desean exponer sus modelos locales y el estado de su servidor de manera segura y estandarizada.

---

## 🚀 Características Principales

### 1. Proxy para Ollama
- **Streaming de alta fidelidad**: Soporte completo para respuestas en tiempo real.
- **Formatos flexibles**: Soporte nativo de Ollama o texto plano (`format=plain`).
- **Persistencia de Contexto**: Mantiene el historial de conversación por usuario utilizando **Valkey** (cache) y **SQLite** (persistencia persistente).
- **Identificación de Usuario**: El modelo reconoce al usuario actual (nombre y email) mediante inyección automática en el prompt de sistema.

### 2. Monitor de Sistema
- Endpoint REST para métricas en tiempo real: `GET /api/v1/system/stats`.
- Monitoreo de: CPU, RAM, Disco y Red.

### 3. Servidor MCP (Model Context Protocol) 
- Expone herramientas locales a LLMs (Modelos de Lenguaje).
- **Herramienta actual**: `system-stats` (Consultar estado del servidor desde el LLM).
- **Transporte**: Stdio (Entrada/Salida estándar).

---

## 🛠 Flujo de Trabajo (Desarrollo)

Sigue estos pasos para configurar tu entorno de desarrollo local.

### Prerrequisitos
- **Go 1.25+** instalado.
- **Ollama** corriendo localmente (por defecto en port 11434).
- **Valkey** o Redis accesible para el cache de contexto.
- **SQLite** para la persistencia de mensajes.
- **Make** (opcional, para usar el Makefile).

### 1. Configuración del Entorno
Crea un archivo `.env` en la raíz del proyecto:
```env
PORT=8080
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=llama3
OLLAMA_SYSTEM_PROMPT="Eres un asistente útil."
VALKEY_ADDRESS=localhost:6379
SQLITE_ADDR=test.db
JWT_SECRET=tu_secreto_super_seguro
```

### 2. Ejecutar la API (Proxy + Métricas)
La API maneja el proxy hacia Ollama y el endpoint de métricas.
```bash
# Usando Make
make run-api

# O comando directo
go run ./cmd/api/main.go
```
La API estará disponible en `http://localhost:8080`.

### 3. Ejecutar el Servidor MCP
El servidor MCP funciona via Stdio, por lo que se ejecuta generalmente a través de un host MCP o para pruebas manuales.

**Prueba manual (JSON-RPC):**
```bash
echo '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | go run ./cmd/mcp/main.go
```

---

## 📦 Despliegue (Producción)

### 1. CI/CD con GitHub Actions
El proyecto cuenta con un flujo automatizado de CI/CD. Cada vez que se realiza un push a la rama principal:
- Se compilan los binarios y se construye la imagen de contenedor.
- La imagen se publica automáticamente en **GitHub Container Registry (GHCR)**.

### 2. Despliegue con Podman Compose
Para desplegar en producción, simplemente utiliza el archivo `compose.yml` incluido. Este descargará la imagen pre-construida desde GHCR.

```bash
# Iniciar todo el stack usando la imagen de GHCR
podman-compose up -d
```

Esto levantará el contenedor `go-local-synapse-proxy` exponiendo el puerto 8080 y configurando todas las variables de entorno necesarias desde tu archivo `.env`.

## 🧠 Uso de la API (Nativo)

¡Tu API ahora es inteligente! No necesitas software extra. El contenedor ya incluye todo lo necesario para orquestar herramientas.

### Consumo desde Web/Mobile
Simplemente consulta el endpoint de chat. La API se encargará de:
1.  Dialogar con Ollama.
2.  Ejecutar herramientas locales (como obtener métricas) si Ollama lo pide.
3.  Devolverte la respuesta final enriquecida.

**Ejemplo de Request:**
```bash
curl -X POST "http://localhost:8080/api/v1/ollama/chat?prompt=Dame%20el%20estado%20del%20servidor"
```

**Respuesta (Automática):**
> "El servidor está estable. El uso de CPU es del 15% y quedan 8GB de RAM libres..."

---

## 🔧 Uso CLI (Opcional con `mcphost`)
Si prefieres interactuar desde la terminal usando `mcphost` en lugar de tu API web, puedes seguir haciéndolo.

1.  **Instalar mcphost** (en el host):
    ```bash
    go install github.com/mark3labs/mcphost@latest
    ```

2.  **Conectar**:
    ```bash
    mcphost config set provider ollama
    mcphost config set ollama-url http://localhost:11434
    mcphost server add local-synapse -- podman exec -i go-local-synapse-proxy /root/mcp
    ```

3.  **Chatear**:
    ```bash
    mcphost chat "Estado del sistema"
    ```

---
*Desarrollado con ❤️ por poch.*
