### API Endpoint: `/xiaozhi/agent/chat-summary/{chatId}/save`

This endpoint is called by the WebSocket server when a chat session ends to summarize the conversation and update the agent's long-term memory.

#### 1. General Information
- **URL:** `/xiaozhi/agent/chat-summary/{chatId}/save`
- **Method:** `POST`
- **Auth:** Protected by manager secret (Middleware level)

#### 2. Path Parameters
| Parameter | Type | Description |
| :--- | :--- | :--- |
| `chatId` | `string` | The unique identifier of the `ai_agent_chat` session to be summarized. |

#### 3. Processing Logic
1.  **Session Retrieval:** Fetches the `ai_agent_chat` record using the provided `chatId`.
2.  **History Retrieval:** Fetches all chat messages from the `ai_agent_chat_history` collection linked to this `chat` ID.
3.  **Agent Identification:** Retrieves the `agent` ID from the `ai_agent_chat` record and loads the corresponding `ai_agent`.
4.  **Enabled Check:** Verifies if `chat_history_enabled` is true for the agent. If false, processing stops.
5.  **Memory Model Loading:**
    - Checks if `mem_model_id` is configured for the agent.
    - Loads the model configuration and resolves any referenced LLM (e.g., if the memory module uses an external LLM provider).
    - Resolves API keys from `user_credentials` via `secret_ref`.
6.  **Prompt Retrieval:** Fetches the summary system prompt from `sys_params` named `memory.system_prompt`. Defaults to a standard summarization prompt if not found.
7.  **Conversation Construction:** 
    - Formats the retrieved chat messages into a "Role: Content" string.
    - If the agent already has an existing `summary_memory`, it is included as "Previous Memories" to ensure continuity.
8.  **LLM Interaction:**
    - Initializes an LLM client (currently supports `openai` type).
    - Sends the combined prompt (System Prompt + Previous Memories + Current Conversation) to the LLM to generate an updated summary.
9.  **Persistence:** 
    - Updates the `summary_memory` field of the `ai_agent` record with the generated summary.
    - Updates the `summary` field of the `ai_agent_chat` record with a short summary of this specific session.

#### 4. Response Structure
Returns the generated summary on success.

```json
{
    "code": 0,
    "msg": "success",
    "data": true
}
```

#### 5. Database Mapping
- **`ai_agent_chat`**: Represents the conversation session.
- **`ai_agent_chat_history`**: Source of individual conversation messages.
- **`ai_agent`**: Target for storing the long-term `summary_memory`.
- **`sys_params`**: Source for `memory.system_prompt`.
- **`model_config` / `model_providers`**: Used to configure the LLM for summarization.
- **`user_credentials`**: Source for LLM API keys.
