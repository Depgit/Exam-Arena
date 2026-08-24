#!/usr/bin/env bash
set -e

# Automated EC2 Installation & Service Setup Script for ExamArena Backend
# Run this script as root / sudo on Ubuntu 22.04/24.04 or Amazon Linux 2023

echo "=== [1/5] Creating examarena system user & directories ==="
id -u examarena &>/dev/null || useradd -r -s /bin/false examarena

mkdir -p /opt/examarena/bin
mkdir -p /var/log/examarena

echo "=== [2/5] Setting directory permissions ==="
chown -R examarena:examarena /opt/examarena
chown -R examarena:examarena /var/log/examarena
chmod 755 /opt/examarena
chmod 755 /opt/examarena/bin

echo "=== [3/5] Installing systemd service ==="
if [ -f "./scripts/examarena.service" ]; then
    cp ./scripts/examarena.service /etc/systemd/system/examarena.service
    systemctl daemon-reload
    systemctl enable examarena
    echo "systemd service installed successfully."
else
    echo "Warning: ./scripts/examarena.service not found."
fi

echo "=== [4/5] Configuring Nginx reverse proxy ==="
if command -v nginx &> /dev/null; then
    if [ -f "./scripts/nginx_examarena.conf" ]; then
        cp ./scripts/nginx_examarena.conf /etc/nginx/sites-available/examarena
        ln -sf /etc/nginx/sites-available/examarena /etc/nginx/sites-enabled/default
        nginx -t && systemctl reload nginx
        echo "Nginx configuration updated and reloaded."
    fi
else
    echo "Nginx is not installed. Installing Nginx..."
    apt-get update && apt-get install -y nginx || yum install -y nginx
fi

echo "=== [5/5] Setup Summary ==="
echo "Place your Linux binary at: /opt/examarena/bin/exam-arena"
echo "Place your environment variables file at: /opt/examarena/.env"
echo "Start the service using: sudo systemctl start examarena"
echo "Check status using: sudo systemctl status examarena"
