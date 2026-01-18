### API Endpoint: `/xiaozhi/ota`

This endpoint handles initial contact and Over-The-Air (OTA) update requests from devices. It serves as a discovery and authentication point for devices to receive their WebSocket server connection details and firmware update information.

#### 1. General Information
- **URL:** `/xiaozhi/ota`
- **Method:** `POST`
- **Auth:** Public (Device identification via header)

#### 2. Request Structure

##### 2.1. Headers
| Header | Required | Format | Description |
| :--- | :--- | :--- | :--- |
| `device-id` | Yes | `XX:XX:XX:XX:XX:XX` | The MAC address of the device. |
| `client-id` | Yes | `string` | Unique identifier for the client application. |

##### 2.2. Body
The request expects a JSON body containing device and application metadata.

```json
{
  "version": 0,
  "uuid": "...",
  "application": {
    "name": "xiaozhi-esp32",
    "version": "1.0.0",
    "compile_time": "2025-01-01 12:00:00"
  },
  "board": {
    "type": "esp32-s3-box",
    "ssid": "MyWiFi",
    "ip": "192.168.1.100",
    "mac": "XX:XX:XX:XX:XX:XX"
  }
}
```

#### 3. Processing Logic
1.  **Header Validation:** Verifies that the `device-id` header is present and follows a valid MAC address format.
2.  **Request Logging:** Stores the raw request body as a JSON string in the `ota_requests` collection, indexed by the `mac_address`.
3.  **Configuration Retrieval:** Fetches the following parameters from the `sys_params` collection:
    -   `server.websocket`: The URL for the real-time communication server.
    -   `server.secret`: The secret key used for HmacSHA256 signing.
4.  **Token Generation:** Generates an HmacSHA256 signature of the string `client-id|device-id|timestamp`. The signature is then Base64 URL-safe encoded (without padding). The final token format is `signature.timestamp`.
5.  **Response Construction:** Returns the server time, firmware version/URL, and WebSocket connection details.

#### 4. Response Structure
```json
{
  "server_time": {
    "timestamp": 1737215340000,
    "timeZone": "Asia/Ho_Chi_Minh",
    "timezone_offset": 420
  },
  "firmware": {
    "version": "1.0.0",
    "url": "http://..."
  },
  "websocket": {
      "url": "ws://...",
      "token": "Base64URLSafeSignature.1737215340"
    }
}
```

##### 4.1. Response Fields
| Field | Type | Description |
| :--- | :--- | :--- |
| `server_time.timestamp` | `long` | Current server time in milliseconds. |
| `server_time.timeZone` | `string` | Server timezone (Fixed: `Asia/Ho_Chi_Minh`). |
| `server_time.timezone_offset` | `int` | Timezone offset in minutes (Fixed: `420`). |
| `firmware.version` | `string` | Latest available firmware version. |
| `firmware.url` | `string` | Download URL for the firmware binary. |
| `websocket.url` | `string` | The WebSocket endpoint for the device to connect to. |
| `websocket.token` | `string` | Combined `Base64URLSafeSignature.timestamp` token for authentication. |

#### 5. Database Mapping
- **`ota_requests`**: Stores historical requests from devices (MAC, Board Type, and Raw JSON).
- **`sys_params`**:
    - `server.websocket`: Source for `websocket.url`.
    - `server.secret`: Used for signing the token.
