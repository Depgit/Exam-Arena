# AWS EC2 Production Deployment Guide for ExamArena Backend

This guide provides step-by-step instructions for deploying the **ExamArena Backend** Go application onto an **AWS EC2 Virtual Machine** (Ubuntu 24.04 LTS or Amazon Linux 2023) using **Nginx**, **systemd**, and **Let's Encrypt SSL**.

---

## 1. Why AWS EC2 for ExamArena?

ExamArena features real-time multiplayer capabilities:
- **Long-lived WebSockets (`gorilla/websocket`)**: Persistent TCP connections for 1v1 battle rooms.
- **In-Memory Matchmaking Engine**: Background goroutines running continuously (`go mmEngine.Run()`).
- **Zero-Latency In-Memory Caching**: Active question banks pre-warmed in RAM.

AWS EC2 supports all stateful features out-of-the-box with **zero code refactoring**.

---

## 2. AWS Infrastructure Setup

### Step 1: Launch EC2 Instance
1. Go to **AWS Console -> EC2 -> Launch Instance**.
2. **AMI**: Ubuntu 24.04 LTS (or Amazon Linux 2023).
3. **Architecture**: 
   - `arm64` for `t4g.small` (Recommended: ~20% lower cost, high performance).
   - `x86_64` for `t3.small` / `t3.medium`.
4. **Key Pair**: Select or create an SSH key pair (`.pem`).

### Step 2: Configure AWS Security Group
Add the following Inbound Rules:

| Type | Protocol | Port Range | Source | Description |
| :--- | :--- | :--- | :--- | :--- |
| **SSH** | TCP | 22 | `Your-Admin-IP/32` | Secure SSH management |
| **HTTP** | TCP | 80 | `0.0.0.0/0` | Web & Certbot SSL validation |
| **HTTPS** | TCP | 443 | `0.0.0.0/0` | Secure REST API & Secure WebSockets (`wss://`) |

> [!NOTE]
> Do **NOT** open Port 8080 to the internet. The Go server will bind locally to `127.0.0.1:8080`, and Nginx will reverse proxy requests securely.

### Step 3: Attach AWS Elastic IP
1. Navigate to **EC2 -> Elastic IPs -> Allocate Elastic IP**.
2. Associate the Elastic IP with your launched EC2 instance to ensure a static public IP address.

---

## 3. Building the Linux Binary

### Windows PowerShell:
```powershell
# For ARM64 (t4g.small):
$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="arm64"
go build -ldflags="-w -s" -o bin/exam-arena-linux ./cmd/server

# For AMD64 / x86_64 (t3.small):
$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="amd64"
go build -ldflags="-w -s" -o bin/exam-arena-linux ./cmd/server
```

### Linux / macOS:
```bash
# Using Makefile:
make build-linux-arm64
# OR
make build-linux-amd64
```

---

## 4. Deploying to EC2

### Step 1: Transfer Files to EC2
```bash
# Copy binary and configuration files to EC2
scp -i your-key.pem bin/exam-arena-linux ubuntu@<EC2-PUBLIC-IP>:/tmp/exam-arena
scp -i your-key.pem scripts/examarena.service ubuntu@<EC2-PUBLIC-IP>:/tmp/
scp -i your-key.pem scripts/nginx_examarena.conf ubuntu@<EC2-PUBLIC-IP>:/tmp/
```

### Step 2: System Setup on EC2
Connect to EC2 via SSH:
```bash
ssh -i your-key.pem ubuntu@<EC2-PUBLIC-IP>
```

Run installation setup:
```bash
# Create application directories & user
sudo useradd -r -s /bin/false examarena || true
sudo mkdir -p /opt/examarena/bin /var/log/examarena

# Move binary into place
sudo mv /tmp/exam-arena /opt/examarena/bin/exam-arena
sudo chmod +x /opt/examarena/bin/exam-arena

# Create production .env file
sudo nano /opt/examarena/.env
```

Add your production environment variables to `/opt/examarena/.env`:
```env
ENVIRONMENT=production
SERVER_HOST=127.0.0.1
SERVER_PORT=8080

# Database Configuration (SQLite or AWS RDS PostgreSQL)
DATABASE_DRIVER=sqlite
DATABASE_URL=/opt/examarena/exam_arena.db

# JWT & Security
JWT_SECRET=your_super_secret_jwt_key_here
CORS_ALLOWED_ORIGINS=*
```

Fix ownership:
```bash
sudo chown -R examarena:examarena /opt/examarena /var/log/examarena
```

---

## 5. Systemd Service Setup

Install the systemd unit file:
```bash
sudo cp /tmp/examarena.service /etc/systemd/system/examarena.service
sudo systemctl daemon-reload
sudo systemctl enable examarena
sudo systemctl start examarena
```

Check process status:
```bash
sudo systemctl status examarena
```

---

## 6. Nginx & Free SSL/TLS (Let's Encrypt)

### Step 1: Configure Nginx Reverse Proxy
```bash
sudo apt-get update
sudo apt-get install -y nginx

sudo cp /tmp/nginx_examarena.conf /etc/nginx/sites-available/examarena
sudo ln -sf /etc/nginx/sites-available/examarena /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl restart nginx
```

### Step 2: Obtain Free SSL Certificate via Certbot
```bash
sudo apt-get install -y certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

Certbot automatically updates Nginx configuration to support HTTPS (`https://yourdomain.com`) and secure WebSockets (`wss://yourdomain.com/ws`).

---

## 7. Automated SQLite Database Backup to AWS S3 (If using SQLite)

To protect your SQLite database (`/opt/examarena/exam_arena.db`) from server failure:

1. Install AWS CLI:
   ```bash
   sudo apt-get install -y awscli
   ```
2. Create automated backup cron job (`crontab -e`):
   ```cron
   # Backup SQLite database to S3 every 3 hours
   0 */3 * * * sqlite3 /opt/examarena/exam_arena.db ".backup '/tmp/exam_arena_backup.db'" && aws s3 cp /tmp/exam_arena_backup.db s3://your-backup-bucket/exam_arena_$(date +\%Y\%m\%d_\%H\%M\%S).db
   ```

---

## 8. Logs & Troubleshooting

- **Check Go Application Logs**:
  ```bash
  sudo tail -f /var/log/examarena/output.log
  ```
- **Check Systemd Logs**:
  ```bash
  sudo journalctl -u examarena -f
  ```
- **Check Nginx Access & Error Logs**:
  ```bash
  sudo tail -f /var/log/nginx/access.log
  sudo tail -f /var/log/nginx/error.log
  ```
