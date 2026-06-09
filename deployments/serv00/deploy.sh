#!/bin/bash
# deployments/serv00/deploy.sh
# Script untuk deploy server hub Go ke hosting serv00.com.
# Jalankan ini di terminal serv00 SSH Anda.

set -e

# Warna output
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}=== Memulai Deploy Server Kiosk Desa di Serv00 ===${NC}"

# 1. Pastikan port sudah direservasi di panel serv00 (misalnya port 3000)
# Buat direktori app
mkdir -p ~/apps/kiosk-desa

# 2. Upload file binary linux 'server' ke serv00 via SFTP/SCP ke: ~/apps/kiosk-desa/server
# Pastikan permissions file binary executable
chmod +x ~/apps/kiosk-desa/server

# 3. Buat file .env di serv00
if [ ! -f ~/apps/kiosk-desa/.env ]; then
    echo -e "${GREEN}Menyiapkan file .env baru di serv00...${NC}"
    read -p "Masukkan Database URL PostgreSQL (contoh: postgres://user:pass@host:5432/dbname): " DB_URL
    read -p "Masukkan Port yang direservasi di Serv00 (contoh: 3000): " PORT
    read -p "Masukkan JWT Secret rahasia: " JWT_SEC
    read -p "Masukkan Allowed Origins (CORS, default: *): " ORIGINS
    
    if [ -z "$ORIGINS" ]; then ORIGINS="*"; fi

    cat <<EOT >> ~/apps/kiosk-desa/.env
LISTEN_ADDR=:$PORT
DATABASE_URL=$DB_URL
JWT_SECRET=$JWT_SEC
ACCESS_TOKEN_EXPIRY=60
REFRESH_TOKEN_EXPIRY=168
ALLOWED_ORIGINS=$ORIGINS
EOT
    echo -e "${GREEN}File .env berhasil dibuat.${NC}"
fi

# 4. Daftarkan cronjob untuk memastikan server selalu berjalan jika restart
# Menggunakan crontab serv00: @reboot ~/apps/kiosk-desa/server > /dev/null 2>&1
echo -e "${GREEN}Mendaftarkan cronjob auto-restart...${NC}"
(crontab -l 2>/dev/null; echo "@reboot cd ~/apps/kiosk-desa && ./server > ~/apps/kiosk-desa/server.log 2>&1 &") | crontab -

# 5. Jalankan server di background
echo -e "${GREEN}Menjalankan server hub...${NC}"
cd ~/apps/kiosk-desa
nohup ./server > server.log 2>&1 &

echo -e "${GREEN}=== DEPLOYMENT BERHASIL! ===${NC}"
echo -e "${GREEN}Server Hub berjalan di port dan mencatat log di ~/apps/kiosk-desa/server.log${NC}"
