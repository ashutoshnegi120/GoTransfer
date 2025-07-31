# 🚀 GoTransfer

A minimal, high-performance file transfer backend written in **Go** using the `net/http` package — built from scratch with concurrency, JWT authentication, and in-memory SQLite support. Designed as a real-world portfolio project.

---

## ✨ Features

- ✅ User Signup & Login (JWT Auth)
- ✅ Upload & Download files with access control
- ✅ In-memory & persistent SQLite DB support
- ✅ YAML-based configuration
- ✅ Built-in middleware for DB and config injection
- ✅ Minimal dependencies, pure `net/http`
- ✅ Jenkins CI with automated email notifications

---

## 📁 Project Structure

```text
GoTransfer/
├── cmd/Go_transfer/         # Main application entry
│   └── main.go              # Server setup and routes
├── internal/
│   ├── config/              # YAML config loader (MustConfig)
│   ├── midlleware/          # JWT middleware
│   └── utile/               # Helpers: DB, JWT claim, keys
├── uploads/                 # Uploaded files (by user UUID)
├── .config/config.yml       # App configuration
├── Jenkinsfile              # CI pipeline
└── README.md
```

---

## 🔐 API Endpoints

| Method | Path                                  | Auth       | Description         |
|--------|---------------------------------------|------------|---------------------|
| POST   | `/signup/{username}/{password}/{email}` | ❌         | Register new user   |
| POST   | `/login/{password}/{email}`           | ❌         | Login, get JWT      |
| POST   | `/upload`                             | ✅ JWT      | Upload a file       |
| GET    | `/download/{fileID}`                  | ✅ JWT      | Download a file     |
| GET    | `/stats/{fileID}`                     | ✅ JWT      | View file stats     |

---

## ⚙️ Configuration

App is configured using a YAML file.

`.config/config.yml`:

```yaml
environment: "dev"
db_location: ":memory:" # or "app.db"
secretKey: "super-secret-key"
http_client:
  port: "8080"
  host: "127.0.0.1"
```

You can pass config in 2 ways:

- **Via env**: `CONFIG_PATH=.config/config.yml go run ...`
- **Via flag**: `go run main.go -config .config/config.yml`

---

## 🚀 Running the App

```bash
go run cmd/Go_transfer/main.go -config .config/config.yml
```

---

## ✅ Running Tests

Includes logic tests for:

- Signup
- Login
- Upload
- Download(pending)

Run tests with:

```bash
go test ./cmd/Go_transfer/ -arg -config=../../.config/config.yml
```

Testable features use in-memory DB and temporary uploads folder.

---

## 🧰 CI/CD with Jenkins

This project includes a Jenkins pipeline with:

- ✅ Build & Test stages
- ✅ Email notifications on success/failure
- ✅ Upload cleanup after test runs

`Jenkinsfile`:

```groovy
pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        sh 'go build cmd/Go_transfer'
      }
    }
    stage('Test') {
      steps {
        sh 'go test ./cmd/Go_transfer/'
      }
    }
  }
  post {
    success {
      emailext(
        subject: "✅ Build SUCCESS: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
        body: "Build passed. Check console: ${env.BUILD_URL}",
        to: "atulsgrr10@gmail.com"
      )
    }
    failure {
      emailext(
        subject: "❌ Build FAILED: ${env.JOB_NAME} #${env.BUILD_NUMBER}",
        body: "Build failed. Check console: ${env.BUILD_URL}console",
        to: "atulsgrr10@gmail.com"
      )
    }
    always {
      echo 'Build completed'
      sh 'rm -rf uploads/' // Cleanup
    }
  }
}
```

---

## 🛠️ Dependencies

- Go 1.24+
- SQLite (`github.com/mattn/go-sqlite3`)
- JWT (`github.com/golang-jwt/jwt`)
- YAML (`gopkg.in/yaml.v3`)

---

## 📌 Example cURL Commands

```bash
# Signup
curl -X POST http://localhost:8080/signup/atul/123321/atul@gmail.com

# Login
curl -X POST http://localhost:8080/login/123321/atul@gmail.com

# Upload (JWT required)
curl -X POST -H "Authorization: Bearer <token>" -F "file=@file.txt" http://localhost:8080/upload

# Download (JWT required)
curl -X GET -H "Authorization: Bearer <token>" http://localhost:8080/download/<fileID>
```

---

## 👍 Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss.

---

## 📄 License

MIT License

---

## 💬 Contact

**Author**: @ashutoshnegi120  
**Email**: ashutoshnegisgrr@gmail.com