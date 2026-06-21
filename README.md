# BankingApp — Sistema de Banca en Línea

Sistema de banca en línea con autenticación, gestión de cuentas, transacciones y un asistente bancario por chat impulsado por IA usando **Model Context Protocol (MCP)**.

## Tecnologías utilizadas

| Capa | Tecnología |
|------|------------|
| Backend | Go (Chi router) |
| Frontend | Vue 3 + Vite + Tailwind CSS |
| Base de datos de usuarios | PostgreSQL |
| Motor financiero | TigerBeetle |
| IA / Chat | OpenRouter (LLM) + MCP (Go SDK oficial) |
| Autenticación | JWT + bcrypt + TOTP (2FA) |
| Gráficas | Chart.js |
| Infraestructura | Docker + Docker Compose |

## Arquitectura

El sistema está compuesto por **5 servicios** orquestados con Docker Compose:

```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│  Frontend   │────▶│   Backend   │────▶│  PostgreSQL  │
│  Vue (nginx)│     │     Go      │     │ (usuarios,   │
│  :5173      │     │   :8080     │     │  metadatos   │
└─────────────┘     └──────┬──────┘     │  de cuentas) │
                            │            └──────────────┘
                            ├──────────────────┐
                            │ (cliente MCP)     │ (cliente TigerBeetle)
                            ▼                   ▼
                     ┌─────────────┐     ┌──────────────┐
                     │ MCP Server  │────▶│ TigerBeetle  │
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

### Arquitectura dual de datos

- **PostgreSQL** almacena usuarios, credenciales (incluyendo secretos TOTP de 2FA), y los metadatos de cada cuenta (nickname, tipo, a qué usuario pertenece). Es la fuente de verdad para todo lo que **no** es dinero.
- **TigerBeetle** es la fuente de verdad financiera: cada cuenta bancaria tiene una cuenta contable espejo en TigerBeetle (su ID se deriva determinísticamente del ID de cuenta en Postgres mediante un hash SHA-256). Todos los depósitos, retiros y transferencias se registran como `transfers` de doble entrada en TigerBeetle, y el balance mostrado al usuario se lee **en tiempo real** directamente desde TigerBeetle, no desde una copia en Postgres.
- Existe una cuenta especial `EXTERNAL` (ID `1`) en TigerBeetle que representa dinero entrando o saliendo del sistema bancario (depósitos y retiros en efectivo).

### Backend como cliente dual

El backend Go mantiene dos conexiones simultáneas:
1. **Cliente PostgreSQL** — para usuarios y metadatos de cuentas.
2. **Cliente TigerBeetle** — para todas las operaciones financieras y lectura de balances.

Además, el backend actúa como **cliente MCP**, conectándose al servidor MCP independiente para que el chat con IA pueda ejecutar operaciones bancarias reales.

### Chat con IA (MCP real)

- **MCP Server (Go):** expone las operaciones bancarias (`get_balance`, `get_history`, `deposit`, `withdraw`, `transfer`) como *tools* MCP, usando el [SDK oficial de Go](https://github.com/modelcontextprotocol/go-sdk). Este servidor también está conectado directamente a TigerBeetle, por lo que las operaciones que ejecuta el chat son igual de reales que las del dashboard.
- El backend recibe el mensaje del usuario, lo envía a un LLM vía OpenRouter junto con las tools disponibles del servidor MCP. El modelo decide qué operación ejecutar, y el backend la dispara a través del cliente MCP.
- **Confirmación obligatoria:** antes de ejecutar cualquier depósito, retiro o transferencia, el asistente describe la operación y pide confirmación explícita al usuario en el mismo chat. Solo ejecuta la tool tras una respuesta afirmativa.
- **Memoria de conversación:** el frontend envía el historial completo de mensajes en cada request, permitiendo que el modelo recuerde el contexto (por ejemplo, la operación pendiente de confirmar).
- **Restricción de seguridad:** el chat con IA solo puede ejecutar transferencias entre cuentas del mismo usuario autenticado. Las transferencias a terceros deben hacerse desde el formulario normal del dashboard.

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

## Datos de prueba

El proyecto incluye un script de carga (`seed`) que inserta una muestra representativa de usuarios, cuentas y transacciones —tomada del JSON oficial de datos de prueba— directamente en PostgreSQL y TigerBeetle, usando la misma lógica de IDs determinísticos que usa el backend.

### Ejecutar la carga

Con el sistema ya levantado (`docker compose up -d`):

```bash
docker compose --profile seed run --rm seed
```

Esto crea **10 usuarios**, **17 cuentas** (con sus saldos iniciales reales en TigerBeetle) y **8 transacciones** de transferencia entre ellas.

### Credenciales de prueba

Tras ejecutar el seed, puedes iniciar sesión con cualquiera de los 10 usuarios generados. Ejemplo:

```
Email: ihernandez@email.com
Password: Isabel2024!
```
```
Email: miguel.perez@email.com
Password: Miguel2024!
```

El script imprime en consola un usuario de ejemplo al finalizar la carga; el email exacto puede variar según el orden en que se procesó el JSON. Ninguno de los usuarios de prueba tiene 2FA activado por defecto, así que el login es de un solo paso.

El servicio `seed` usa Docker Compose **profiles**, por lo que nunca se levanta automáticamente con `docker compose up` — solo se ejecuta cuando se invoca explícitamente con `--profile seed`.

Archivos relevantes:
- `database/sample-data.json` — muestra de datos extraída del dataset oficial
- `database/seed.go` — script de carga
- `database/seed.Dockerfile` — imagen específica para ejecutar el seed dentro del entorno Docker

## Funcionalidades

### Autenticación
- Registro de usuarios con contraseña hasheada (bcrypt)
- Login con JWT (expira en 20 minutos)
- Logout
- **Autenticación de dos factores (2FA) con TOTP** — compatible con Google Authenticator, Authy y similares, configurable desde la página de Seguridad
- Middleware de autenticación en rutas protegidas

### Gestión de cuentas
- Creación de cuenta bancaria automática al registrarse (tipo ahorros por defecto), con su cuenta contable espejo en TigerBeetle
- Soporte para múltiples cuentas por usuario (corriente / ahorros)
- Edición de nickname de cuenta
- Eliminación de cuenta (solo si el saldo en TigerBeetle es $0)

### Transacciones (procesadas en TigerBeetle)
- Depósito y retiro, con selector explícito de cuenta de origen/destino
- Retiro con validación de saldo (TigerBeetle bloquea balances negativos por defecto)
- Transferencia con dos modos: **entre cuentas propias** o **a un tercero**, cada uno con sus propios selectores
- Historial de transacciones por usuario o filtrado por cuenta, con exportación a CSV
- Gráfica de ingresos vs gastos por mes en el dashboard

### Chat con IA (MCP)
El usuario puede interactuar en lenguaje natural desde un widget de chat flotante en el dashboard:

- *"¿Cuánto dinero tengo?"*
- *"Deposita 50 dólares a mi cuenta de ahorros"*
- *"Transfiere 20 dólares de mi cuenta corriente a mi cuenta de ahorros"*
- *"Muéstrame mis últimas transacciones"*
- *"¿Por qué no puedo cerrar mi cuenta?"*

El LLM interpreta la intención, confirma la operación antes de ejecutarla, llama a la tool MCP correspondiente (que opera sobre TigerBeetle real), y responde en lenguaje natural con el resultado. El dashboard se refresca automáticamente tras cualquier operación realizada por chat.

## Estructura del proyecto

```
banking-app/
├── backend/              # API REST en Go
│   ├── cmd/server/       # Punto de entrada
│   ├── internal/
│   │   ├── handlers/     # Controladores HTTP (incluye 2FA)
│   │   ├── services/     # Lógica de negocio, cliente MCP, OpenRouter
│   │   ├── repository/   # Acceso a PostgreSQL y TigerBeetle
│   │   └── models/       # Structs de datos
│   └── pkg/config/       # Configuración por variables de entorno
├── mcp-server/           # Servidor MCP independiente
│   ├── internal/         # Repository, cliente TigerBeetle y definición de tools
│   └── main.go
├── frontend/             # SPA en Vue 3
│   └── src/
│       ├── pages/        # Login, registro, dashboard, transacciones, historial, seguridad
│       ├── components/   # ChatWidget, IncomeExpenseChart
│       ├── stores/       # Pinia (auth, account)
│       └── router/       # Vue Router
├── database/             # Scripts SQL, inicialización de TigerBeetle y datos de prueba
│   ├── schema.sql
│   ├── init-tigerbeetle.sh
│   ├── seed.go
│   ├── seed.Dockerfile
│   └── sample-data.json
└── docker-compose.yml    # Orquestación de los 5 servicios + servicio de seed opcional
```

## Endpoints principales

| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/api/auth/register` | Registro de usuario |
| POST | `/api/auth/login` | Login (paso 1: credenciales, paso 2: código 2FA si aplica) |
| POST | `/api/auth/logout` | Logout |
| GET | `/api/account` | Cuentas del usuario (balance leído de TigerBeetle) |
| POST | `/api/account/create` | Crear cuenta nueva (Postgres + TigerBeetle) |
| PUT | `/api/account/{id}/nickname` | Editar nombre de cuenta |
| DELETE | `/api/account/{id}` | Eliminar cuenta (saldo $0 en TigerBeetle) |
| POST | `/api/transactions/deposit` | Depositar (vía TigerBeetle), requiere `account_id` |
| POST | `/api/transactions/withdraw` | Retirar (vía TigerBeetle), requiere `account_id` |
| POST | `/api/transactions/transfer` | Transferir (vía TigerBeetle), requiere `from_account_id` y `to_account_id` |
| GET | `/api/transactions/history` | Historial, con filtro opcional por cuenta |
| POST | `/api/chat` | Chat con el asistente bancario IA (vía MCP) |
| POST | `/api/2fa/setup` | Genera secreto y código QR para activar 2FA |
| POST | `/api/2fa/confirm` | Confirma el código inicial y activa 2FA |
| POST | `/api/2fa/disable` | Desactiva 2FA |

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

Cada carpeta (`backend/`, `mcp-server/`) requiere su propio archivo `.env` con `DATABASE_URL`, `JWT_SECRET`, `TIGERBEETLE_ADDR`, `OPENROUTER_API_KEY`, etc. Revisa `pkg/config/config.go` para ver las variables soportadas.

**Nota sobre TigerBeetle y CGO:** el cliente Go de TigerBeetle usa CGO con una librería nativa que requiere `io_uring`. En Windows con Docker Desktop/WSL2 esto puede requerir habilitar `kernel.io_uring_disabled=0` y correr los contenedores con `privileged: true` (ya configurado en `docker-compose.yml`). Por este motivo, scripts que usan el cliente de TigerBeetle (como `seed.go`) están diseñados para ejecutarse **dentro** de un contenedor Docker, no directamente en el host.

## Decisiones técnicas y notas

- **Moneda:** todas las cuentas operan en USD.
- **Arquitectura dual real:** se optó por una separación estricta de responsabilidades — PostgreSQL nunca almacena el balance autoritativo, solo TigerBeetle. Esto cumple el espíritu del requisito de "arquitectura dual" de la prueba: dos bases de datos, cada una responsable de un dominio distinto.
- **IDs determinísticos:** los IDs de cuenta de TigerBeetle (`uint64`) se derivan de los IDs de cuenta de PostgreSQL (`string`, formato `4001-XXXX-XXXX-XXXX`) mediante SHA-256, garantizando que ambos sistemas siempre referencien la misma cuenta sin necesidad de una tabla de mapeo adicional.
- **Uint128 y endianness:** TigerBeetle representa los valores de 128 bits (`Uint128`) en formato little-endian a nivel de bytes. Al convertir a `big.Int` de Go (que espera big-endian), es necesario invertir el orden de los bytes antes de la conversión; de lo contrario, los balances se calculan incorrectamente.
- **2FA con TOTP:** implementado con la librería `pquerna/otp`, compatible con el estándar Google Authenticator (SHA1, 6 dígitos, periodo de 30s). El secreto se guarda en PostgreSQL solo tras la primera confirmación exitosa.
- **MCP real:** se optó por implementar un servidor MCP independiente (en vez de function calling directo) para cumplir fielmente el protocolo, permitiendo que el backend actúe como cliente MCP estándar, reutilizable por cualquier otro cliente compatible con el protocolo.
- **Modelo de IA:** se usa un modelo gratuito de OpenRouter compatible con tool calling, configurable en `internal/services/openrouter.go`.
- **Datos de prueba:** se cargó una muestra representativa (no el dataset completo de 1000 usuarios) para mantener tiempos de carga y pruebas manejables; el script `seed.go` es fácilmente extensible a un dataset mayor ajustando el archivo JSON de entrada.
- **Seguridad:** las contraseñas se almacenan con bcrypt; los tokens JWT expiran en 20 minutos; los archivos `.env` están excluidos del control de versiones; el chat con IA tiene restricciones explícitas sobre qué operaciones puede ejecutar y bajo qué condiciones.
- **Límite de OpenRouter (modelo gratuito):** la cuenta gratuita de OpenRouter tiene un límite diario de ~50 requests. Si el chat deja de responder con un error de "Rate limit exceeded", significa que se alcanzó ese límite — se resuelve esperando al día siguiente o agregando una key con saldo propio en OpenRouter.
- **Limitación conocida — confirmación del chat:** la regla de "confirmar antes de ejecutar" vive en el system prompt del LLM, no en el código del backend. Por la naturaleza no determinística de los modelos de lenguaje (especialmente modelos gratuitos), existe una probabilidad baja pero no nula de que el modelo ejecute una operación sin pedir confirmación explícita. Una mejora futura sería forzar la confirmación a nivel de código.