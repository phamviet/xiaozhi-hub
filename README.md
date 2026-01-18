# Xiaozhi Hub

Xiaozhi Hub is a specialized backend management system designed to support intelligent AI hardware (such as ESP32-based voice assistants). Built on top of [PocketBase](https://pocketbase.io/), it provides a robust and extensible platform for managing AI agents, device configurations, OTA updates, and conversation history.

## 🚀 Key Features

- **Device Management & Binding**: Secure 6-digit code binding flow to link physical hardware with AI agents and user accounts.
- **AI Agent Configuration**: Centralized management of AI module configurations (ASR, VAD, LLM, TTS, Intent, and Memory).
- **OTA (Over-The-Air) Updates**: Automated firmware delivery system with secure HmacSHA256 token-based authentication.
- **Conversation History**: Persistent storage of chat interactions, including support for binary audio file storage via PocketBase's filesystem API.
- **Long-term Memory & Summarization**: Incremental chat summarization using LLMs to maintain agent persona and user context over multiple sessions.
- **Secret Management**: Secure resolution of API keys via credential references, preventing sensitive data from being exposed in plain text configurations.

## 🛠 Architecture

Xiaozhi Hub is implemented as a custom plugin for a Go-based [PocketBase](https://pocketbase.io/) application. It leverages PocketBase's core features for:

- **Auth**: User and device authentication.
- **Database**: Real-time SQLite storage for all collections.
- **Filesystem**: Managed storage for audio recordings and firmware files.
- **Admin UI**: Built-in interface for managing data and system parameters.

## 📂 Project Structure

- `/xiaozhi`: Core logic for the Xiaozhi Hub plugin (API handlers, managers, and types).
- `/internal/hub`: General-purpose framework for PocketBase extensions.
- `/docs`: Detailed API documentation for all endpoints.
- `/migrations`: Database schema versioning and initial setup.
- `/ui`: (Planned) Frontend management interface.

## 📖 API Documentation

Detailed documentation for individual API routes can be found in the `/docs` directory:

- [OTA Endpoint](docs/api-ota.md)
- [Agent Models Configuration](docs/api-agent-models.md)
- [Device Binding](docs/api-device-bind.md)
- [Chat History Report](docs/api-report-chat.md)
- [Chat Summary & Memory](docs/api-chat-summary.md)
- [Database Schema Overview](docs/database-schema.md)

## 🏁 Getting Started

1. **Environment Setup**: Ensure you have Go 1.25+ installed.
2. **Build**: Run `go build -o pb main.go` to compile the application.
3. **Seed**: Run `./pb seeds` to load default system parameters and model configurations.
4. **Run**: Execute `./pb serve` to start the PocketBase server.
5. **Admin UI**: Access the dashboard at `http://127.0.0.1:8090/_/` to configure system parameters and manage collections.
