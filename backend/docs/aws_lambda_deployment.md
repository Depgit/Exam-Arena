# AWS Lambda Deployment Guide for ExamArena Backend

This guide provides step-by-step instructions for deploying the **ExamArena Backend** REST APIs to **AWS Lambda** with **AWS API Gateway** and **AWS RDS PostgreSQL**.

---

## 1. Prerequisites

Before starting the deployment, ensure you have installed:
1. **Go 1.25+**
2. **AWS CLI** (configured via `aws configure`)
3. **AWS SAM CLI** (Serverless Application Model CLI)
4. An active **AWS Account** with permissions for Lambda, API Gateway, and RDS.

---

## 2. Architecture Overview

- **REST API Function**: Compiled Go binary (`cmd/lambda/main.go`) running on AWS Lambda (`provided.al2023`, `arm64` Graviton2 architecture).
- **API Gateway**: HTTP API routing requests (`/{proxy+}`) directly to the Lambda function.
- **Database**: AWS RDS PostgreSQL or Aurora Serverless v2 PostgreSQL.
- **Real-Time Services**: Real-time WebSockets and in-memory matchmaking queue can run alongside on **AWS ECS Fargate** or **AWS App Runner** (Hybrid Architecture).

---

## 3. Database Setup (AWS RDS PostgreSQL)

1. Provision an **AWS RDS PostgreSQL** instance or **Aurora Serverless v2 PostgreSQL** cluster.
2. Note your connection string format:
   ```env
   DATABASE_DRIVER=postgres
   DATABASE_URL=postgresql://<username>:<password>@<rds-endpoint>:5432/<dbname>?sslmode=require
   ```
3. The backend will automatically execute schema migrations on startup when connected to the database.

---

## 4. Local Build & Packaging

To compile the Go binary for AWS Lambda (Linux ARM64):

### Windows PowerShell:
```powershell
$env:GOOS="linux"
$env:GOARCH="arm64"
go build -tags lambda.norpc -o bootstrap ./cmd/lambda
```

### Linux / macOS:
```bash
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/lambda
```

---

## 5. Deployment with AWS SAM

1. Validate the SAM Template:
   ```bash
   sam validate
   ```

2. Build and Package:
   ```bash
   sam build
   ```

3. Deploy to AWS:
   ```bash
   sam deploy --guided
   ```

Follow the interactive prompts:
- **Stack Name**: `exam-arena-backend`
- **AWS Region**: `us-east-1` (or your preferred region)
- **Confirm changes before deploy**: `y`
- **Allow SAM CLI IAM role creation**: `y`
- **Save arguments to configuration file**: `y`

---

## 6. Environment Variables

Configure the following environment variables in AWS Lambda (or in `template.yaml`):

| Variable Name | Example Value | Description |
| :--- | :--- | :--- |
| `DATABASE_DRIVER` | `postgres` | Switch database driver to PostgreSQL |
| `DATABASE_URL` | `postgresql://user:pass@rds-endpoint:5432/exam_arena?sslmode=require` | RDS PostgreSQL Connection String |
| `JWT_SECRET` | `your-secure-random-jwt-secret-key` | Secret key used for signing JWT tokens |
| `CORS_ALLOWED_ORIGINS` | `https://yourdomain.com` | Allowed origin header for CORS |
| `ENVIRONMENT` | `production` | Environment name |

---

## 7. Verification & Health Check

Once deployment completes, SAM output will provide your API Gateway endpoint URL:
```text
ExamArenaApiUrl = https://abc123xyz.execute-api.us-east-1.amazonaws.com
```

Test the health check endpoint:
```bash
curl https://abc123xyz.execute-api.us-east-1.amazonaws.com/health
```

Expected Response:
```json
{"status":"ok","environment":"aws_lambda"}
```

---

## 8. Hybrid Architecture (WebSockets & Matchmaking on ECS Fargate)

For multiplayer real-time battles and WebSockets (`/ws`), deploy the containerized server (`cmd/server/main.go`) to **AWS ECS Fargate** or **AWS App Runner**:

1. Build Docker image using `./docker/Dockerfile`
2. Push to **AWS ECR**:
   ```bash
   aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <aws-account-id>.dkr.ecr.us-east-1.amazonaws.com
   docker build -t exam-arena-backend .
   docker tag exam-arena-backend:latest <aws-account-id>.dkr.ecr.us-east-1.amazonaws.com/exam-arena-backend:latest
   docker push <aws-account-id>.dkr.ecr.us-east-1.amazonaws.com/exam-arena-backend:latest
   ```
3. Launch ECS Fargate Service connected to the same AWS RDS PostgreSQL instance.
