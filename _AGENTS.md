# Xiaozhi Hub - AI Context

## Project Overview
Xiaozhi Hub is a backend management system for intelligent AI hardware (specifically ESP32-based voice assistants), built on top of **PocketBase** using **Go**. It acts as the central control plane for managing devices, AI agents, firmware updates (OTA), and conversation history.

The frontend is a Single Page Application (SPA) built with **React**, **TypeScript**, and **Vite**, using **Tailwind CSS** for styling and **PocketBase** as the backend service.

## Build, Lint, and Test Commands

### Backend (Go)
*   **Run Development Server:** `make dev-api` (uses `air` for live reload)
*   **Build Docker Image:** `make docker`
*   **Generate Schema Snapshot:** `make snapshot` (runs `go run . migrate collections`)
*   **Run Tests:** `go test ./...` (Run specific test: `go test -v -run TestName ./path/to/package`)
    *   *Note: Currently no tests exist in the codebase. When adding tests, place `*_test.go` files alongside the source code.*
*   **Lint:** Follow standard Go conventions (`go vet`).

### Frontend (TypeScript/React)
*   **Install Dependencies:** `bun install --cwd ./ui`
*   **Run Development Server:** `make dev-ui` (or `bun run --cwd ./ui dev`)
*   **Build:** `make build-ui` (or `bun run --cwd ./ui build`)
*   **Lint & Format (Biome):**
    *   Check (Lint & Format): `bun run --cwd ./ui check`
    *   Fix Issues: `bun run --cwd ./ui check:fix`
    *   Format Only: `bun run --cwd ./ui format`
    *   Lint Only: `bun run --cwd ./ui lint`

### Full Stack
*   **Run All (Dev):** `make dev` (runs both API and UI in parallel)

## Code Style Guidelines

### General
*   **Conventions:** Adhere to existing project conventions. Analyze surrounding code before making changes.
*   **Comments:** Use comments to explain *why*, not *what*.
*   **Pathing:** Always use absolute paths when referencing files in tool calls.

### Go (Backend)
*   **Formatting:** Standard `gofmt` is mandatory.
*   **Naming:**
    *   Use `PascalCase` for exported identifiers and `camelCase` for unexported ones.
    *   Package names should be short, lowercase, and descriptive (e.g., `hub`, `services`).
*   **Imports:** Group imports: Standard Library, Third-party (e.g., `github.com/...`), Local (`github.com/phamviet/xiaozhi-hub/internal/...`).
*   **Error Handling:**
    *   Return errors; do not panic (except during strict initialization).
    *   Wrap errors with context when propagating: `fmt.Errorf("failed to do something: %w", err)`.
*   **Architecture:**
    *   Core logic resides in `internal/`.
    *   Service layer pattern (`internal/hub/services`) separates business logic from handlers.
    *   PocketBase is extended via hooks (`OnBootstrap`, `OnServe`).

### TypeScript (Frontend)
*   **Formatting/Linting:** Strict adherence to **Biome** rules. Run `bun run check:fix` before committing.
*   **Component Style:**
    *   Use Functional Components with Hooks.
    *   Use `lucide-react` for icons.
    *   Use `@radix-ui` primitives for accessible UI components.
*   **State Management:** Use **Nanostores** (`@nanostores/react`) for global state.
*   **Styling:** Use **Tailwind CSS** utility classes. Avoid inline styles or separate CSS files unless necessary.
*   **Data Fetching:** Use the PocketBase JS SDK for interacting with the backend.
*   **Naming:**
    *   Components: `PascalCase` (e.g., `DeviceList.tsx`).
    *   Functions/Variables: `camelCase`.
    *   Files: `camelCase.ts` or `PascalCase.tsx` (for components).

## Environment Setup
*   **Go:** Version 1.23+ (inferred from generic Go projects, check `go.mod` if specific).
*   **Node/Bun:** Bun is the primary runtime/package manager for the frontend.
*   **PocketBase:** Embedded in the Go application.
*   **Air:** Required for backend live reloading (`go install github.com/air-verse/air@latest`).

## Directory Structure
*   `internal/hub/`: Core hub logic, plugin system, and server setup.
*   `internal/hub/ws/`: WebSocket handling (Client, Dispatcher, Handlers).
*   `internal/hub/services/`: Business logic services (Device, Session, History).
*   `ui/`: React frontend application.
*   `migrations/`: PocketBase schema migrations.
*   `pb_data/`: PocketBase data directory (gitignored).
