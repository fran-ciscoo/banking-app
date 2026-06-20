#!/bin/sh
# ============================================================
# BankingApp — Script de inicialización de TigerBeetle
# ============================================================
#
# Este script formatea el cluster de TigerBeetle y prepara
# el archivo de datos necesario para que el servidor arranque.
#
# TigerBeetle no usa SQL: su modelo de datos son "accounts"
# (cuentas contables) y "transfers" (movimientos entre ellas),
# definidos mediante su API binaria, no mediante scripts de texto.
#
# Este script realiza el formateo inicial del cluster —
# equivalente a "crear la base de datos" en un motor relacional.
# El mismo proceso ya está automatizado en docker-compose.yml,
# por lo que correr esto manualmente es solo necesario si se
# quiere levantar TigerBeetle fuera de Docker.
#
# Uso:
#   sh init-tigerbeetle.sh
# ============================================================

set -e

DATA_FILE="/data/0_0.tigerbeetle"
CLUSTER_ID=0
REPLICA_ID=0
REPLICA_COUNT=1

echo "Formateando cluster de TigerBeetle..."

if [ ! -f "$DATA_FILE" ]; then
  /tigerbeetle format \
    --cluster=$CLUSTER_ID \
    --replica=$REPLICA_ID \
    --replica-count=$REPLICA_COUNT \
    "$DATA_FILE"
  echo "Cluster formateado correctamente en $DATA_FILE"
else
  echo "El archivo de datos ya existe, se omite el formateo."
fi

echo "Iniciando TigerBeetle en 0.0.0.0:3000..."
/tigerbeetle start --addresses=0.0.0.0:3000 "$DATA_FILE"