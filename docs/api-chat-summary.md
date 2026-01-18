### API Endpoint: `/xiaozhi/agent/chat-summary/{sessionId}/save`

This endpoint is called by the WebSocket server when a chat session ends to summarize the conversation and update the agent's long-term memory.

#### 1. General Information
- **URL:** `/xiaozhi/agent/chat-summary/{sessionId}/save`
- **Method:** `POST`
- **Auth:** Protected by manager secret (Middleware level)

#### 2. Path Parameters
| Parameter | Type | Description |
| :--- | :--- | :--- |
| `sessionId` | `string` | The unique identifier of the conversation/session to be summarized. |

#### 3. Processing Logic
1.  **Session Retrieval:** Fetches all chat messages from the `ai_agent_chat_history` collection matching the provided `sessionId` (conversation_id).
2.  **Agent Identification:** Retrieves the `agent_id` from the chat history records and loads the corresponding `ai_agent`.
3.  **Enabled Check:** Verifies if `chat_history_enabled` is true for the agent. If false, processing stops.
4.  **Memory Model Loading:**
    - Checks if `mem_model_id` is configured for the agent.
    - Loads the model configuration and resolves any referenced LLM (e.g., if the memory module uses an external LLM provider).
    - Resolves API keys from `user_credentials` via `secret_ref`.
5.  **Prompt Retrieval:** Fetches the summary system prompt from `sys_params` named `memory.system_prompt`. Defaults to a standard summarization prompt if not found.
6.  **Conversation Construction:** 
    - Formats the retrieved chat messages into a "Role: Content" string.
    - If the agent already has an existing `summary_memory`, it is included as "Previous Memories" to ensure continuity.
7.  **LLM Interaction:**
    - Initializes an LLM client (currently supports `openai` type).
    - Sends the combined prompt (System Prompt + Previous Memories + Current Conversation) to the LLM to generate an updated summary.
8.  **Persistence:** Updates the `summary_memory` field of the `ai_agent` record with the generated summary.

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
- **`ai_agent_chat_history`**: Source of conversation messages.
- **`ai_agent`**: Target for storing the generated summary.
- **`sys_params`**: Source for `memory.system_prompt`.
- **`model_config` / `model_providers`**: Used to configure the LLM for summarization.
- **`user_credentials`**: Source for LLM API keys.
