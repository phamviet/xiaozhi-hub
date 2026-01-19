# Xiaozhi Hub - AI Context

## Project Overview
Xiaozhi Hub is a backend management system for intelligent AI hardware (specifically ESP32-based voice assistants), built on top of **PocketBase** using **Go**. It acts as the central control plane for managing devices, AI agents, firmware updates (OTA), and conversation history.

## Core Domain Entities

### 1. Devices (`ai_device`)
- Represents physical hardware, identified by a unique ID (often MAC address based).
- **Key Attributes**: `mac_address`, `firmware_version`, `board` type.
- **Relationships**: Can be bound to a `user` and assigned an `ai_agent`.

### 2. AI Agents (`ai_agent`)
- Represents the "persona" or "brain" running on a device.
- **Configuration**:
  - `system_prompt`: Defines behavior and personality.
  - `lang_code`: Primary language.
- **Modular Architecture**: An agent is composed of references to specific model configurations:
  - **ASR** (Automatic Speech Recognition)
  - **LLM** (Large Language Model)
  - **TTS** (Text-to-Speech)
  - **VAD** (Voice Activity Detection)
  - **Memory** (Context/Summarization)
  - **Intent** (Action recognition)

### 3. Models (`model_config`, `model_providers`)
- **Providers**: Definitions of services like OpenAI, Anthropic, Azure, etc.
- **Configs**: Specific instances of a provider (e.g., "GPT-4o" with specific parameters).
- **Credentials**: API keys are stored separately in `user_credentials` and referenced, ensuring security.

### 4. Chat History (`ai_agent_chat_history`)
- Stores the interaction log between the user and the agent.
- **Content**: Text transcripts and optional binary audio files.
- **Memory**: Used as input for the summarization process to build long-term memory.

## Key Workflows

### Device Binding
- **Mechanism**: 6-digit code pairing.
- **Process**:
  1. Device requests a binding code.
  2. Server generates a short-lived code.
  3. User enters the code in the frontend/app.
  4. Server links the Device to the User.
- **Code**: `xiaozhi/device_bind.go`, `xiaozhi/device_binding.go`.

### OTA (Over-The-Air) Updates
- **Mechanism**: Token-based secure firmware delivery.
- **Process**: Device polls for updates; Server checks version and serves binary if a newer version exists.
- **Code**: `xiaozhi/ota.go`.

### Agent Configuration Sync
- **Mechanism**: JSON configuration delivery.
- **Process**: Device requests its config; Server resolves all assigned models (LLM, TTS, etc.) and credentials, returning a consolidated JSON object.
- **Code**: `xiaozhi/config_agent_models.go`.

### Chat Reporting & Summarization
- **Mechanism**: REST API upload + LLM processing.
- **Process**:
  1. Device uploads chat records (text/audio).
  2. Server stores them in `ai_agent_chat_history`.
  3. System (periodically or on-trigger) summarizes recent chats to update the agent's "Memory".
- **Code**: `xiaozhi/report_chat.go`, `xiaozhi/summary_chat.go`.

## Architecture & Tech Stack
- **Language**: Go.
- **Framework**: PocketBase (provides Auth, DB, Admin UI, Filesystem).
- **Database**: SQLite (embedded, WAL mode).
- **API Style**: RESTful, extending PocketBase's router.

## Directory Map
- `/xiaozhi`: **Core Business Logic**. Contains API handlers, model managers, and type definitions.
- `/internal/hub`: **Framework Extensions**. Generic utilities for PocketBase.
- `/docs`: **Documentation**. API specs and schema details.
- `/migrations`: **Database Schema**. Go-based migrations for PocketBase collections.
- `/pb_data`: **Runtime Data**. Database file and uploaded media (usually gitignored).
