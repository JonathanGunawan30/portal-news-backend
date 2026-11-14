# Portal News Backend

A robust and secure backend API for a news portal application, built with Go using the Fiber framework. This project follows a clean architecture pattern and includes modern security features like global rate limiting and Google reCAPTCHA v3 integration.

## 🚀 Features

- **User Authentication**: Secure login with JWT (JSON Web Tokens).
- **Admin Management**: Dedicated endpoints for categories and content management.
- **Content Storage**: Integrated with **Cloudflare R2** for scalable object storage.
- **Database**: PostgreSQL with **GORM** for efficient ORM and migrations.
- **Security**:
  - **Global Rate Limiting**: 60 requests per 10 minutes per IP to prevent DDoS and brute-force attacks.
  - **Google reCAPTCHA v3**: Mandatory token verification for admin login to ensure human-only access.
- **API Documentation**: Interactive **Swagger UI** available in development mode.
- **Validation**: Strict request validation using `go-playground/validator`.
- **Environment Driven**: Fully configurable via `.env` or environment variables.

## 🛠 Tech Stack

- **Language**: [Go (1.27+)](https://go.dev/)
- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **ORM**: [GORM](https://gorm.io/)
- **Configuration**: [Viper](https://github.com/spf13/viper)
- **Object Storage**: [Cloudflare R2 (S3-compatible)](https://www.cloudflare.com/products/r2/)
- **Documentation**: [Swagger / OpenAPI](https://swagger.io/)
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)

## 📋 Prerequisites

- Go 1.27.4 or higher
- PostgreSQL
- Cloudflare R2 Account (for file uploads)
- Google reCAPTCHA v3 Secret Key

## ⚙️ Installation & Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/JonathanGunawan30/portal-news-backend.git
   cd portal-news-backend
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Configure Environment Variables**:
   Create a `.env` file in the root directory and configure the following:

   ```env
   # App Configuration
   APP_PORT=8000
   APP_ENV=development # development | production
   
   # Database Configuration
   DATABASE_HOST=localhost
   DATABASE_PORT=5432
   DATABASE_USER=postgres
   DATABASE_PASSWORD=your_password
   DATABASE_NAME=portal_news
   DATABASE_MAX_OPEN_CONNECTION=100
   DATABASE_MAX_IDLE_CONNECTION=10
   
   # JWT Configuration
   JWT_SECRET_KEY=your_jwt_secret
   JWT_ISSUER=portal-news
   
   # Cloudflare R2 Configuration
   CLOUDFLARE_R2_BUCKET_NAME=your_bucket_name
   CLOUDFLARE_R2_API_KEY=your_r2_api_key
   CLOUDFLARE_R2_API_SECRET=your_r2_api_secret
   CLOUDFLARE_R2_TOKEN=your_r2_token
   CLOUDFLARE_R2_ACCOUNT_ID=your_account_id
   CLOUDFLARE_R2_PUBLIC_URL=https://your-public-url.com
   
   # Security
   RECAPTCHA_SECRET_KEY=your_recaptcha_v3_secret_key
   ```

## 🏃 Running the Application

### Using Go CLI
To start the server:
```bash
go run main.go start
```

### Using Docker
1. **Build the image**:
   ```bash
   docker build -t portal-news-backend .
   ```
2. **Run the container**:
   ```bash
   docker run -p 8000:8000 --env-file .env portal-news-backend
   ```

## 📖 API Documentation

In **development mode** (`APP_ENV=development`), you can access the Swagger documentation at:
`http://localhost:8000/api/docs`

The specification is located at `docs/apispec.yaml`.

## 🗄 Database Migrations

Migrations are located in `database/migrations/`. You can use your preferred migration tool (e.g., `golang-migrate`) or apply them manually to your PostgreSQL instance.

## 📁 Project Structure

```text
├── cmd/                # CLI commands (Cobra)
├── config/             # Configuration logic and structs
├── database/           # DB Migrations and Seeders
├── docs/               # API Documentation (Swagger)
├── internal/
│   ├── adapter/        # Handlers, Repositories, External Adapters
│   ├── app/            # Application entry point & Middleware setup
│   └── core/           # Domain Entities, Services, Models
├── lib/                # Shared libraries (JWT, reCAPTCHA, etc.)
└── validator/          # Custom validation logic
```

## 🛡 Security Notes

- **Rate Limiting**: Globally applied at 60 requests/ 10 minutes.
- **reCAPTCHA**: The `/api/login` endpoint requires a valid `recaptcha_token` in the request body. Verification is performed against Google's API with a minimum required score of **0.5**.
