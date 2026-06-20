# BankingApp — Sistema de Banca en Línea

Sistema de banca en línea con autenticación, gestión de cuentas, transacciones y un asistente bancario por chat impulsado por IA usando **Model Context Protocol (MCP)**.

## Tecnologías utilizadas

| Capa | Tecnología |
|------|------------|
| Backend | Go (Chi router) |
| Frontend | Vue 3 + Vite + Tailwind CSS |
| Base de datos transaccional | PostgreSQL |
| Base de datos financiera | TigerBeetle |
| IA / Chat | OpenRouter (LLM) + MCP (Go SDK oficial) |
| Autenticación | JWT + bcrypt |
| Infraestructura | Docker + Docker Compose |

## Arquitectura

El sistema está compuesto por **5 servicios** orquestados con Docker Compose:

```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│  Frontend   │────▶│   Backend   │────▶│  PostgreSQL  │
│  Vue (nginx)│     │     Go      │     │ (usuarios,   │
│  :5173      │     │   :8080     │     │  cuentas,    │
└─────────────┘     └──────┬──────┘     │  transacc.)  │
                            │            └──────────────┘
                            │ (cliente MCP)
                            ▼
                     ┌─────────────┐     ┌──────────────┐
                     │ MCP Server  │     │ TigerBeetle  │
                     │     Go      │     │   :3000      │
                     │   :9090     │     └──────────────┘
                     └──────┬──────┘
                            │
                            ▼
                     ┌─────────────┐
                     │ OpenRouter  │
                     │   (LLM)     │
                     └─────────────┘
```

- **Backend (Go):** expone la API REST, maneja autenticación JWT y actúa como **cliente MCP**.
- **MCP Server (Go):** expone las operaciones bancarias (`get_balance`, `get_history`, `deposit`, `withdraw`, `transfer`) como *tools* MCP, usando el [SDK oficial de Go](https://github.com/modelcontextprotocol/go-sdk).
- **Chat con IA:** el backend recibe el mensaje del usuario, lo envía a un LLM vía OpenRouter junto con las tools disponibles del servidor MCP. El modelo decide qué operación ejecutar, y el backend la dispara a través del cliente MCP.

> **Nota sobre TigerBeetle:** el contenedor de TigerBeetle está configurado y corriendo en el `docker-compose.yml`, listo para integrarse como motor contable. Actualmente las operaciones financieras (depósitos, retiros, transferencias) se procesan sobre PostgreSQL.

## Cómo levantar el proyecto

### Requisito único: Docker

```bash
git clone https://github.com/fran-ciscoo/banking-app.git
cd banking-app
```

### Variables de entorno

Crea un archivo `.env` en la raíz del proyecto:

```env
OPENROUTER_API_KEY=tu-api-key-de-openrouter
```

Puedes obtener una key gratuita en [openrouter.ai/keys](https://openrouter.ai/keys).

### Levantar todo el sistema

```bash
docker compose up -d --build
```

Esto levanta:
- PostgreSQL en `localhost:5432`
- TigerBeetle en `localhost:3000`
- Servidor MCP en `localhost:9090`
- Backend (API REST) en `localhost:8080`
- Frontend en `localhost:5173`

### Acceder a la aplicación

Abre [http://localhost:5173](http://localhost:5173) en tu navegador.

## Funcionalidades

### Autenticación
- Registro de usuarios con contraseña hasheada (bcrypt)
- Login con JWT (expira en 24 horas)
- Logout
- Middleware de autenticación en rutas protegidas

### Gestión de cuentas
- Creación de cuenta bancaria automática al registrarse
- Soporte para múltiples cuentas por usuario (corriente / ahorros)
- Edición de nickname de cuenta
- Eliminación de cuenta (solo si el saldo es $0)

### Transacciones
- Depósito
- Retiro (con validación de saldo)
- Transferencia entre cuentas (propias o de terceros)
- Historial de transacciones por usuario (todas sus cuentas)

### Chat con IA (MCP)
El usuario puede interactuar en lenguaje natural desde un widget de chat flotante en el dashboard:

- *"¿Cuánto dinero tengo?"*
- *"Deposita 50 dólares a mi cuenta de ahorros"*
- *"Transfiere 20 dólares a la cuenta 4001-XXXX-XXXX-XXXX"*
- *"Muéstrame mis últimas transacciones"*

El LLM interpreta la intención, llama a la tool MCP correspondiente, y responde en lenguaje natural con el resultado real de la operación.

## Estructura del proyecto

```
banking-app/
├── backend/              # API REST en Go
│   ├── cmd/server/       # Punto de entrada
│   ├── internal/
│   │   ├── handlers/     # Controladores HTTP
│   │   ├── services/     # Lógica de negocio, cliente MCP, OpenRouter
│   │   ├── repository/   # Acceso a PostgreSQL
│   │   └── models/       # Structs de datos
│   └── pkg/config/       # Configuración por variables de entorno
├── mcp-server/           # Servidor MCP independiente
│   ├── internal/         # Repository y definición de tools
│   └── main.go
├── frontend/             # SPA en Vue 3
│   └── src/
│       ├── pages/        # Login, registro, dashboard, transacciones, historial
│       ├── components/   # ChatWidget y otros componentes
│       ├── stores/       # Pinia (auth, account)
│       └── router/       # Vue Router
├── database/             # Datos de prueba
└── docker-compose.yml    # Orquestación de los 5 servicios
```

## Endpoints principales

| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/api/auth/register` | Registro de usuario |
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/logout` | Logout |
| GET | `/api/account` | Cuentas del usuario |
| POST | `/api/account/create` | Crear cuenta nueva |
| PUT | `/api/account/{id}/nickname` | Editar nombre de cuenta |
| DELETE | `/api/account/{id}` | Eliminar cuenta (saldo $0) |
| POST | `/api/transactions/deposit` | Depositar |
| POST | `/api/transactions/withdraw` | Retirar |
| POST | `/api/transactions/transfer` | Transferir |
| GET | `/api/transactions/history` | Historial de transacciones |
| POST | `/api/chat` | Chat con el asistente bancario IA |

## Desarrollo local (sin Docker)

Si prefieres correr cada servicio manualmente:

```bash
# Terminal 1 — bases de datos
docker compose up -d postgres tigerbeetle

# Terminal 2 — servidor MCP
cd mcp-server
go run main.go

# Terminal 3 — backend
cd backend
go run cmd/server/main.go

# Terminal 4 — frontend
cd frontend
pnpm install
pnpm run dev
```

Cada carpeta (`backend/`, `mcp-server/`) requiere su propio archivo `.env` con `DATABASE_URL`, `JWT_SECRET`, `OPENROUTER_API_KEY`, etc. Revisa `pkg/config/config.go` para ver las variables soportadas.

## Decisiones técnicas y notas

- **Moneda:** todas las cuentas operan en USD.
- **MCP real:** se optó por implementar un servidor MCP independiente (en vez de function calling directo) para cumplir fielmente el protocolo, permitiendo que el backend actúe como cliente MCP estándar, reutilizable por cualquier otro cliente compatible con el protocolo.
- **Modelo de IA:** se usa un modelo gratuito de OpenRouter compatible con tool calling, configurable en `internal/services/openrouter.go`.
- **Seguridad:** las contraseñas se almacenan con bcrypt; los tokens JWT expiran en 24 horas; los archivos `.env` están excluidos del control de versiones.