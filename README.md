# PaperTrading

A realistic stock market simulation platform where you can practice trading with virtual money that mirrors real market conditions.

## 🚀 Overview

PaperTrading is a full-stack application that allows users to practice stock trading without financial risk. Users can trade with virtual money using real market data, learn trading strategies, and track their performance over time.

### Key Features

- **Risk-Free Practice**: Trade with virtual money using real market data
- **Affordable Top-ups**: Refill virtual funds at minimal cost
- **Real Market Simulation**: Live market data integration for authentic trading experience
- **Portfolio Tracking**: Comprehensive analytics and performance tracking
- **Educational Resources**: Learn trading strategies without financial consequences
- **Dual Theme Support**: Dark and light mode options

## 🏗 Architecture

This project consists of two main components:

### Frontend
- React + Vite + TypeScript
- Mantine UI components with theming
- TanStack Router for navigation
- [Frontend Documentation →](./frontend/README.md)

### Backend
- Go with go-chi router
- PostgreSQL database with pgx/pgxpool
- JWT authentication system
- [Backend Documentation →](./backend/README.md)

## 🚀 Quick Start

1. **Clone the repository**
	```bash
	git clone https://github.com/yourusername/papertrading.git
	cd papertrading
	```
   
2. **Set up the backend**
	```bash
	cd backend
	# Follow backend/README.md for detailed setup
	```
	
3. **Set up the frontend**
	```bash
	cd frontend
	# Follow frontend/README.md for detailed setup
	```
	
### 📂 Repository Structure
	```plaintext
	papertrading/
	├── frontend/               # React application
	│   ├── src/
	│   ├── package.json
	│   └── README.md
	├── backend/               # Go API server
	│   ├── cmd/
	│   ├── internal/
	│   ├── go.mod
	│   └── README.md
	├── docs/                  # Additional documentation
	└── README.md             # This file
	```