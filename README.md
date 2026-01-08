# Local Synapse 🧠

**Local Synapse** es un proxy ligero escrito en Go diseñado para conectar un servidor local de [Ollama](https://ollama.com/) con el mundo exterior. 

Actualmente, el proyecto se encuentra en una fase pre-MVP, sirviendo como una herramienta de validación para asegurar que el servidor local es accesible de forma segura y eficiente desde redes externas.

## 🚀 Estado Actual

Hoy en día, Local Synapse actúa como un proxy robusto para la API de Chat de Ollama, ofreciendo:
- **Streaming de alta fidelidad**: Soporte para respuestas largas mediante un búfer optimizado de 1MB.
- **Eficiencia de recursos**: Cancelación automática de peticiones a Ollama si el cliente se desconecta.
- **Configuración simple**: Configurable mediante variables de entorno (`PORT`, `OLLAMA_URL`, `SYSTEM_PROMPT`).

## 🛠 Instalación y Uso

### Prerrequisitos
- Go 1.25+
- Ollama corriendo localmente.

### Ejecución
1. Clona el repositorio.
2. Ejecuta el servidor:
   ```bash
   go run main.go
   ```
### Ejemplo con cURL
Para ver el streaming en tiempo real desde la terminal:
```bash
curl -N "http://localhost:8080/chat?prompt=Explícame+Go+en+una+frase"
```
*(El flag `-N` es importante para desactivar el buffering de cURL).*

### Ejemplo con JavaScript (Frontend)
Si quieres consumirlo desde una web:
```javascript
const eventSource = new EventSource('http://localhost:8080/chat?prompt=Hola');

eventSource.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log(data.message.content); // Aquí llega el fragmento de texto
};

eventSource.onerror = () => {
    eventSource.close();
};
```

### Ejecución con Podman/Docker
El proyecto se construye automáticamente en cada cambio a `main`. Puedes obtener la imagen desde GitHub Packages:
```bash
podman pull ghcr.io/${{ github.repository }}:latest
```

O ejecutarlo directamente con Compose:
```bash
podman-compose up -d
```

## 🔮 Visión a Futuro

Este proyecto no se detendrá en ser un simple proxy. El objetivo es evolucionar hacia una plataforma integrada que permita:

1.  **Gestión de Proyectos Electrónicos**: Una interfaz para visualizar datos de sensores, controlar actuadores y organizar esquemáticos/documentación técnica de mis proyectos.
2.  **Asistente LLM Genérico**: Un compañero de IA personalizado que no solo responda preguntas, sino que entienda el contexto de mis desarrollos locales.
4.  **Integración con MCP (Model Context Protocol)**: Actuar como un host o cliente MCP para permitir que el LLM interactúe dinámicamente con herramientas externas y bases de conocimiento.
5.  **Frontend Interactivo**: Un panel de control moderno para visualizar el estado de los proyectos electrónicos y chatear con los modelos de forma fluida.
6.  **Puente Hardware-IA**: Utilizar la potencia de los LLMs locales para analizar telemetría de hardware en tiempo real.

---
*Desarrollado con ❤️ por poch.*
