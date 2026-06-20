-- ============================================================
-- BankingApp — Script de creación de tablas (PostgreSQL)
-- ============================================================
-- Este script crea el esquema completo usado para usuarios,
-- cuentas y transacciones bancarias.
--
-- Uso:
--   psql -U postgres -d banking -f schema.sql
--
-- También se ejecuta automáticamente al arrancar el backend
-- (ver backend/internal/repository/postgres.go → CreateTables),
-- por lo que correr este script manualmente es opcional.
-- ============================================================

-- Extensión necesaria para generar UUIDs en transacciones
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ------------------------------------------------------------
-- Tabla: users
-- Usuarios registrados en el sistema. Las contraseñas se
-- almacenan hasheadas con bcrypt, nunca en texto plano.
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY,
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,
    full_name  VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ------------------------------------------------------------
-- Tabla: accounts
-- Cuentas bancarias. Cada usuario puede tener varias cuentas
-- (corriente, ahorros). El balance se actualiza directamente
-- en cada depósito, retiro o transferencia.
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS accounts (
    id         VARCHAR(20) PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id),
    type       VARCHAR(20) NOT NULL,            -- 'checking' | 'savings'
    nickname   VARCHAR(100),                    -- nombre personalizado, editable por el usuario
    balance    DECIMAL(15,2) DEFAULT 0,
    currency   VARCHAR(10) DEFAULT 'USD',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);

-- ------------------------------------------------------------
-- Tabla: transactions
-- Registro de todos los movimientos: depósitos, retiros y
-- transferencias. 'EXTERNAL' representa dinero que entra o
-- sale del sistema bancario (depósitos y retiros en efectivo).
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS transactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_account VARCHAR(20),                   -- 'EXTERNAL' en depósitos
    to_account   VARCHAR(20),                    -- 'EXTERNAL' en retiros
    amount       DECIMAL(15,2) NOT NULL,
    type         VARCHAR(30) NOT NULL,           -- 'deposit' | 'withdrawal' | 'transfer'
    description  VARCHAR(255),
    status       VARCHAR(20) DEFAULT 'completed',
    timestamp    TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_from_account ON transactions(from_account);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account ON transactions(to_account);
CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp DESC);

-- ============================================================
-- Fin del script
-- ============================================================