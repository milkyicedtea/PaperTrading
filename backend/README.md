# PaperTrading Backend

Go-based REST API server for the PaperTrading platform.

## 🛠 Tech Stack

- **Go 1.21+**
- **go-chi** - HTTP router and middleware
- **pgx/pgxpool** - PostgreSQL driver and connection pooling
- **JWT** - Authentication (access & refresh tokens)
- **PostgreSQL** - Database

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher

## 🚀 Getting Started

### Installation

```bash
# Clone the repository (if not already done)
git clone https://github.com/yourusername/papertrading.git
cd papertrading/backend

# Install dependencies
go mod download
```