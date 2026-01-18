### API Endpoint: `/xiaozhi/config/agent-models`

This endpoint retrieves the full configuration for an AI agent associated with a specific device. It includes the prompt, chat history settings, and the detailed configurations for various AI modules (ASR, LLM, TTS, etc.).

#### 1. General Information
- **URL:** `/xiaozhi/config/agent-models`
- **Method:** `POST`
- **Auth:** Protected by manager secret (Middleware level)

#### 2. Request Structure
The request expects a JSON body with the device's identification.

```json
{
  "clientId": "unique_client_id",
  "macAddress": "XX:XX:XX:XX:XX:XX",
  "selectedModule": {}
}
```

| Field | Type | Description |
| :--- | :--- | :--- |
| `clientId` | `string` | A unique identifier for the client session. |
| `macAddress` | `string` | The MAC address of the device requesting the config. |
| `selectedModule` | `object` | (Optional) Currently selected modules on the client. |

#### 3. Processing Logic
The server follows these steps to generate the configuration:

1.  **Device Identification:** Looks up the device in the `ai_device` collection using the provided `macAddress`.
2.  **Device Binding Flow:** If the device is not found, the server initiates the binding flow:
    - Generates a 6-digit random code.
    - Stores the code in memory, mapped to the `macAddress` and `clientId`.
    - Returns a response with error code `10042` and the generated code in the `msg` field.
3.  **Agent Retrieval:** Finds the associated AI Agent in the `ai_agent` collection via the `agent` ID on the device record.
4.  **Base Response Initialization:**
    - Loads the `system_prompt` and `summary_memory` from the Agent record.
    - Sets `chat_history_conf`: `2` (voice & message) if enabled on the agent, otherwise `0`.
4.  **Module Configuration Loading:** For each required module type (`ASR`, `VAD`, `TTS`, `LLM`, `Memory`, `Intent`), it:
    - Retrieves the `model_config` record. If the agent doesn't specify one, it falls back to the system default for that type.
    - Retrieves the `provider_code` from the `model_providers` collection.
    - Flattens the `config_json` from the database and injects the `provider_code` as the `type` field.
5.  **Secret Resolution:** If a configuration contains a `secret_ref`, the server looks up the corresponding `api_key` from the `user_credentials` collection and replaces the reference with the actual key.
6.  **LLM Reference Handling:** Some modules (like Intent or Memory) might reference an LLM by ID in their parameters (e.g., `{"llm": "some_llm_id"}`). The system automatically detects these references, loads the corresponding LLM configuration, and attaches it to the `LLM` key in the root `data` object to ensure the client has all necessary model details.

#### 4. Response Structure
##### 4.1. Success Response (code 0)
The response is wrapped in a standard success envelope. See section 5 for the detailed `data` object structure.

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

##### 4.2. Binding Needed Response (code 10042)
If the device is not bound to any agent, this response is returned.

```json
{
    "code": 10042,
    "data": null,
    "msg": "272728"
}
```

| Field | Type | Description |
| :--- | :--- | :--- |
| `code` | `int` | `10042` indicates that the device is not bound. |
| `data` | `null` | No data is returned in this case. |
| `msg` | `string` | A 6-digit randomly generated code to be displayed to the user for binding. |

#### 5. Detailed Response Structure (Success)
The response is wrapped in a standard success envelope.

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "prompt": "The system prompt for the AI...",
    "summaryMemory": "",
    "chat_history_conf": 2,
    "device_max_output_size": "0",
    "selected_module": {
      "ASR": "ASR_ID",
      "LLM": "LLM_ID",
      "TTS": "TTS_ID",
      "Intent": "Intent_ID",
      "Memory": "Memory_ID"
    },
    "ASR": {
      "ASR_ID": {
        "type": "openai",
        "api_key": "...",
        "base_url": "...",
        "model_name": "..."
      }
    },
    "LLM": {
       "LLM_ID": {
         "type": "openai",
         "api_key": "...",
         "model_name": "..."
       }
    },
    "TTS": {
      "TTS_ID": {
        "type": "edge",
        "voice": "..."
      }
    },
    "Intent": {
      "Intent_ID": {
        "type": "intent_llm",
        "llm": "LLM_ID"
      }
    },
    "Memory": {
      "Memory_ID": {
        "type": "mem_local_short",
        "llm": "LLM_ID"
      }
    },
    "plugins": {
      "get_weather": "{\"api_key\": \"test\", \"api_host\": \"mj7p3y7naa.re.qweatherapi.com\", \"default_location\": \"广州\"}",
      "play_music": "{}"
    }
  }
}
```

##### 4.1. Data Fields

| Field | Description |
| :--- | :--- |
| `prompt` | The primary instructions/persona for the AI agent. |
| `summaryMemory` | Long-term memory or context summary. |
| `chat_history_conf` | `0`: Disabled, `2`: Store voice and text messages. |
| `device_max_output_size` | Maximum size for device output (default "0"). |
| `selected_module` | Maps module types (ASR, LLM, etc.) to the specific config ID used. |
| `[ModuleType]` | Objects containing configurations for each module, keyed by their ID. |
| `plugins` | Configuration for various agent capabilities (e.g., weather, music). |

##### 4.2. Module Configuration Object
Each module (ASR, LLM, etc.) contains specific parameters depending on its provider, but generally includes:

| Field | Description |
| :--- | :--- |
| `type` | The provider type (e.g., `openai`, `edge`, `gemini`). |
| `api_key` | Resolved API key (if applicable). |
| `model_name` | The specific model identifier (e.g., `gpt-4o`). |
| `base_url` | API endpoint for the provider. |

#### 5. Database Mapping
- **`ai_device`**: Links physical hardware (MAC) to an AI Agent.
- **`ai_agent`**: Defines the agent's behavior and which models it uses by default.
- **`model_config`**: Stores the specific parameters (JSON) for a model.
- **`model_providers`**: Defines the `type` (code) of the model provider.
- **`user_credentials`**: Stores encrypted/protected API keys referenced in configs.
