# Local Synapse 🧠

**Local Synapse** es un proxy ligero escrito en Go diseñado para conectar un servidor local de [Ollama](https://ollama.com/) con el mundo exterior. 

Actualmente, el proyecto se encuentra en una fase pre-MVP, sirviendo como una herramienta de validación para asegurar que el servidor local es accesible de forma segura y eficiente desde redes externas.

## 🚀 Estado Actual

Actualmente, el proyecto ofrece:
- **Proxy Ollama**: Streaming de alta fidelidad, soporte para respuestas en plano (`format=plain`) o SSE.
- **Métricas del Sistema**: Endpoint para monitorear CPU, RAM, Disco y Red.
- **Servidor MCP (Model Context Protocol)**: Servidor basado en Stdio para exponer herramientas locales a LLMs.

## 🛠 Instalación y Uso

### Prerrequisitos
- **Go 1.25+**
- **Ollama** corriendo localmente (opcional si solo usas métricas).

### Configuración
Crea un archivo `.env` basado en la configuración necesaria:
```env
PORT=8080
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=llama3
OLLAMA_SYSTEM_PROMPT="Eres un asistente útil."
```

### Ejecución

#### 1. API Principal
Ejecuta el servidor API que incluye el proxy de Ollama y las métricas:
```bash
make run-api
```
Endpoint de métricas: `GET /api/v1/system/stats`

#### 2. Servidor MCP (Stdio)
Si deseas usar las herramientas locales con un host MCP (como `mcphost` o Claude Desktop):
```bash
go run ./cmd/mcp/main.go
```

## 🔗 Configuración de `mcphost` (Remoto)

Para que un LLM en un servidor remoto o local pueda interactuar con las herramientas de este proyecto, se recomienda usar [mcphost](https://github.com/mark3labs/mcphost).

### Pasos en el servidor remoto:

1. **Instalar mcphost**:
   ```bash
   go install github.com/mark3labs/mcphost@latest
   ```

2. **Configurar el puente**:
   Debes configurar `mcphost` para que use Ollama como proveedor y se conecte a este proyecto como un servidor de herramientas.

   Ejemplo de configuración para `mcphost`:
   ```bash
   mcphost config set provider ollama
   mcphost config set ollama-model llama3:latest
   mcphost config set ollama-url http://tu-ip-u-host:11434
   ```

3. **Registrar Local Synapse como servidor MCP**:
   Ya que el servidor MCP usa Stdio, si `mcphost` corre en una máquina distinta, podrías necesitar un túnel (como SSH) o ejecutarlo localmente donde reside el proyecto.

   Si `mcphost` tiene acceso al binario compilado de `mcp`:
   ```bash
   mcphost server add local-synapse -- ./mcp
   ```

---
*Desarrollado con ❤️ por poch.*
