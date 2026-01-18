### API Endpoint: `/xiaozhi/agent/chat-history/report`

This endpoint allows devices to report user chat messages, including optional audio data, to be stored in the AI agent's chat history.

#### 1. General Information
- **URL:** `/xiaozhi/agent/chat-history/report`
- **Method:** `POST`
- **Auth:** Protected by manager secret (Middleware level)

#### 2. Request Structure
The request expects a JSON body with chat details and optional Base64 encoded audio.

```json
{
    "audioBase64": "...",
    "chatType": 2,
    "content": "User message",
    "macAddress": "57:3B:18:35:12:A1",
    "reportTime": 1768575812,
    "sessionId": "b5b1887c-acd0-42df-8596-dca78e3a5309"
}
```

| Field | Type | Description                                                            |
| :--- | :--- |:-----------------------------------------------------------------------|
| `audioBase64` | `string` | (Optional) Base64 encoded audio data.                                  |
| `chatType` | `int` | Type of chat message (e.g., `1` for User or `2` for Assistant).        |
| `content` | `string` | The text content of the user message.                                  |
| `macAddress` | `string` | The MAC address of the device.                                         |
| `reportTime` | `long` | Timestamp when the message was reported.                               |
| `sessionId` | `string` | Unique session/conversation identifier.                                |

#### 3. Processing Logic
1.  **Request Parsing:** Binds the JSON body to a structured object.
2.  **Device Identification:** Looks up the device in the `ai_device` collection using `macAddress` to retrieve the associated `agent_id`.
3.  **Audio Processing:** If `audioBase64` is provided:
    - Decodes the Base64 string into raw bytes.
    - Creates a new file object using PocketBase's filesystem API.
4.  **Record Creation:** Inserts a new record into the `ai_agent_chat_history` collection:
    - `mac_address`: From request.
    - `agent_id`: From the identified device.
    - `conversation_id`: Mapped from `sessionId`.
    - `content`: From request.
    - `chat_type`: From request (as string).
    - `audio`: The processed audio file (if provided).
5.  **Persistence:** Saves the record to the database and filesystem.

#### 4. Response Structure
Returns a standard success response.

```json
{
    "code": 0,
    "msg": "success",
    "data": true
}
```

#### 5. Database Mapping
- **`ai_device`**: Used to link `macAddress` to `agent_id`.
- **`ai_agent_chat_history`**: Stores the chat history records and audio files.
