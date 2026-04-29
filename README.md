# JGC Attendance System

An advanced, enterprise-grade Employee Attendance and Workforce Management System built with a high-performance **Go (Fiber)** backend and a reactive **SvelteKit** frontend.

## Overview

This system provides precise tracking for employee check-ins, lunch breaks, and check-outs, featuring strict **geofencing perimeter validation** and automated incident detection.

### Key Features

- **Geofenced Check-In/Out**: Validates user GPS coordinates against Work Center boundaries.
- **Mobile Workforce (Field Service)**: Support for employees with dynamic locations, allowing multiple daily check-ins and disabling geofence penalties for `field` shift types.
- **Mandatory Visual Evidence**: Enforces photo capture before `check-out` to ensure auditability of field operations.
- **Automated Incident Detection**: Real-time identification of tardiness (`late`) and GPS proximity failures (`out_of_range`).
- **Unified Temporal Accuracy**: Migrated to `TIMESTAMPTZ` schema to ensure robust 24/7 duration arithmetic and prevent data corruption across midnight shifts.
- **Dynamic Financial Reporting**: Automated generation of payroll cycle reports with incident-based penalty deductions.
- **PDF Export**: editorial-grade payroll and attendance statements.
- **Security Hardening**: Strict JWT algorithm validation, standardized error masking to prevent DB leakage, and restrictive CORS policies.

---

## Technology Stack

### Backend

- **Core**: Go (Golang)
- **Framework**: [Fiber](https://gofiber.io/)
- **Database**: PostgreSQL (Relational persistence)
- **Connectivity**: SQLX & Lib/PQ

### Frontend

- **Framework**: [SvelteKit](https://kit.svelte.dev/)
- **Logic**: TypeScript/JavaScript
- **Styling**: Vanilla CSS (Tailored UI system)
- **State Management**: Reactive Stores

---

## Architecture & Infrastructure

The application follows a decoupled Client-Server architecture designed for scalability and clear separation of concerns.

### Backend Infrastructure (Go)

- **Host**: Go 1.22+ compiled binary.
- **API Framework**: [Fiber](https://gofiber.io/) — A minimalist, Express-inspired web framework for Go.
- **Database**: PostgreSQL 15+ — Chosen for its strong ACID compliance and native `TIMESTAMPTZ` / `Interval` support, essential for workforce time-tracking.
- **Data Access**: `sqlx` (General purpose SQL extensions) for clean relational mapping.
- **Hot Reload**: `Air` — Enhances DX by automatically recompiling the binary on file changes.

### Frontend Infrastructure (SvelteKit)

- **Framework**: [SvelteKit](https://kit.svelte.dev/) via [Vite](https://vitejs.dev/).
- **Package Management**: Managed via `Bun` for optimized speed, fallback to `NPM`.
- **Styling Engine**: Modular Vanilla CSS with a global design token system (Aesthetic: Executive Ethereal).

---

## Backend Folder Structure

The backend follows a clean architecture pattern, isolating logic from transport layers:

```bash
backend/
├── cmd/
│   └── api/             # Entry point: App initialization, route mapping, and listener start.
├── internal/
│   ├── config/          # .env parsing and database connection pooling.
│   ├── database/        # Migration runners and initialization logic.
│   ├── handlers/        # Controllers: HTTP request processing, param validation, and response mapping.
│   ├── middleware/      # Interceptors: JWT Auth, Role-based protection, and logging.
│   ├── models/          # Data Layer: DB schemas, DTOs, and conversion mappers.
│   ├── repository/      # Persistence logic (if separate from handlers).
│   └── services/        # Business Logic: Calculations (GPS distance, PDF creation) that are handler-agnostic.
├── migrations/          # SQL scripts (UP/DOWN) for version-controlled schema evolution.
└── scratch/             # DevOps utilities, migration tools, and manual verification scripts.
```

## Quick Start

### Backend (Developing)

1. Navigate to `/backend`.
2. Configure your `.env` (DB credentials, Ports).
3. Run with Live Reload:
   ```bash
   air
   ```
4. **Configuración**: Copia `.env.example` a `.env` y ajusta las variables de seguridad y base de datos.

### Frontend (Developing)

1. Navigate to `/frontend`.
2. Install dependencies:
   ```bash
   bun install # or npm install
   ```
3. Start the dev server:
   ```bash
   bun run dev
   ```

---

## Quality Assurance & Testing

The system includes a comprehensive testing suite to ensure business logic integrity, especially around financial calculations and temporal arithmetic.

### Backend Testing (Go)

We use `go test` along with `testify` for assertions and `sqlmock` for database isolation.

**Run all tests:**

```bash
go test ./...
```

**Key Test Areas:**

- **Attendance Logic**: Verification of distance calculations (Haversine), lateness detection (including midnight shifts), and compliance policy enforcement.
- **Financial Calculations**: Validation of gross/net earnings and penalty deductions based on job positions.
- **Report Generation**: Ensuring data consistency in aggregated payroll cycles.

---

## Documentation

For deeper context on agents, operational entities, and business logic, refer to:

- [AGENTS.md](./AGENTS.md) - Core logic and architectural entities.
- [implementation_plan.md](./implementation_plan.md) - Recent refactor history and feature roadmaps.
