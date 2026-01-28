# 🧠 Personal MCP Tools Repository

**Personal MCP Tools Repository** es mi colección personal de herramientas para el [Model Context Protocol (MCP)](https://modelcontextprotocol.io/). Este repositorio contiene herramientas MCP que he desarrollado para uso personal y que pueden ser integradas en cualquier proyecto que utilice MCP.

## 🚀 Características

### Herramientas MCP Disponibles

#### 1. **system-stats** 📊
- **Descripción**: Obtiene métricas del sistema en tiempo real
- **Métricas incluidas**:
  - Uso de CPU (porcentaje)
  - Uso de RAM (porcentaje y GB usados/totales)
  - Uso de Disco (porcentaje y GB usados/totales)
  - Tráfico de Red (bytes enviados/recibidos)
- **Uso**: `system-stats`

### Arquitectura Modular
- **Estructura clara**: Cada herramienta en su propio archivo dentro de `internal/pkg/mcp_tools/`
- **Fácil extensión**: Añade nuevas herramientas siguiendo el patrón existente
- **Servidor MCP independiente**: Ejecutable standalone en `cmd/mcp/main.go`

## 🛠️ Uso

### 1. Como Servidor MCP Standalone

El servidor MCP funciona via Stdio y puede ser utilizado por cualquier cliente MCP:

```bash
# Ejecutar el servidor MCP
go run ./cmd/mcp/main.go

# Probar manualmente (JSON-RPC)
echo '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | go run ./cmd/mcp/main.go
```

### 2. Integración con Clientes MCP

#### Con Claude Desktop:
```json
{
  "mcpServers": {
    "personal-tools": {
      "command": "go",
      "args": ["run", "/ruta/a/tu/proyecto/cmd/mcp/main.go"]
    }
  }
}
```

#### Con Cline u otros clientes:
Configura el servidor para ejecutar el binario compilado o el comando `go run`.

### 3. Como Biblioteca en Otros Proyectos

Puedes importar las herramientas individualmente en tus proyectos Go:

```go
import (
    "github.com/metalpoch/local-synapse/internal/pkg/mcp_tools"
)

// Usar la herramienta system-stats
tool, handler := mcptools.SystemStats()
```

## 🏗️ Desarrollo de Nuevas Herramientas

### Estructura de una Herramienta MCP

Cada herramienta sigue este patrón en `internal/pkg/mcp_tools/`:

```go
package mcptools

import (
    "context"
    "fmt"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

func MiNuevaHerramienta() (tool mcp.Tool, handler server.ToolHandlerFunc) {
    return mcp.NewTool(
            "mi-herramienta",
            mcp.WithDescription("Descripción de mi herramienta"),
            mcp.WithStringSchema("parametro", mcp.WithDescription("Descripción del parámetro")),
        ),
        func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            // Lógica de la herramienta aquí
            return mcp.NewToolResultText("Resultado"), nil
        }
}
```

### Pasos para Añadir una Nueva Herramienta:

1. **Crear el archivo** en `internal/pkg/mcp_tools/`
2. **Implementar la función** que retorna `(mcp.Tool, server.ToolHandlerFunc)`
3. **Registrar la herramienta** en `cmd/mcp/main.go`:
   ```go
   s.AddTool(mcptools.MiNuevaHerramienta())
   ```

### Ejemplos de Herramientas Potenciales:
- **file-operations**: Operaciones básicas de archivos
- **git-utils**: Comandos Git comunes
- **docker-status**: Estado de contenedores Docker
- **weather-check**: Clima local
- **todo-manager**: Gestor de tareas personal

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

## 📦 Despliegue (Local)

### Compilar Binarios y ejecutar

```bash
make run-api
```

## 🔧 Configuración

### Variables de Entorno

Crea un archivo `.env` en la raíz:

```env
PORT=8080
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=qwen3:4b
OLLAMA_SYSTEM_PROMPT="Eres un asistente muy cute y algo tsundere que termina cada parrafo con 'datebayo'"
```

## 🧪 Testing

```bash
# Ejecutar tests unitarios
go test ./...

# Ejecutar tests con cobertura
go test -cover ./...

# Ejecutar tests específicos
go test ./internal/pkg/mcp_tools/...
```

## 📁 Estructura del Proyecto

```
.
├── cmd/
│   ├── mcp/          # Servidor MCP principal
│   └── api/          # API HTTP (opcional, para integración web)
├── internal/
│   ├── pkg/mcp_tools/ # Todas las herramientas MCP
│   │   ├── system_stats.go
│   │   └── [nueva_herramienta].go
│   ├── infrastructure/ # Infraestructura compartida
│   └── usecase/       # Casos de uso (si aplica)
├── mcp                # Script MCP para ejecución directa
├── Makefile          # Automatización
├── compose.yml       # Docker Compose
├── Containerfile     # Dockerfile
└── README.md         # Este archivo
```

## 🤝 Contribución

Este es un repositorio personal, pero si encuentras errores o tienes sugerencias:

1. **Reporta issues** para bugs o mejoras
2. **Sugiere nuevas herramientas** que podrían ser útiles
3. **Sigue los patrones existentes** para consistencia

## 📄 Licencia

Este proyecto es de uso personal. Consulta el archivo LICENSE para más detalles.

---

*Desarrollado con ❤️ por poch. Mantenido como mi repositorio personal de herramientas MCP.*
