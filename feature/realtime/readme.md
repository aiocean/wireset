# Realtime Feature

The realtime feature provides WebSocket-based real-time communication capabilities for your application. It allows users to join rooms, send messages, and receive updates in real-time.

## Value Proposition

1. Real-time communication: Enable instant messaging and updates within your application.
2. Room-based interactions: Organize users into separate rooms for focused discussions or group activities.
3. Scalable architecture: Designed to work with multiple instances, though there are known scaling issues to be addressed.
4. Event-driven: Utilizes CQRS (Command Query Responsibility Segregation) for handling commands and events.

## How to Use

### 1. Setup

Ensure you have the necessary dependencies installed and the feature is properly initialized in your application.

### 2. Connecting to WebSocket

To connect to the WebSocket, make a GET request to:

```
/api/v1/ws?username=<USERNAME>&roomID=<ROOM_ID>
```

Replace `<USERNAME>` with the user's name and `<ROOM_ID>` with the desired room identifier.

### 3. Sending Messages

Once connected, you can send messages using the following format:

```json
{
  "topic": "<TOPIC>",
  "payload": <PAYLOAD>
}
```

Replace `<TOPIC>` with the desired message topic and `<PAYLOAD>` with the message content.

### 4. Handling Messages

The server uses a registry system to handle different message topics. Implement handlers for specific topics to process incoming messages.

### 5. Disconnecting

The server automatically handles disconnections and cleans up resources when a user leaves.

## Implementation Details

- The main handler is in `feature/realtime/api/handle.go`.
- Room management is handled in `feature/realtime/room/room.go` and `feature/realtime/room/manager.go`.
- Message sending is managed by `feature/realtime/command/sendWsMessageHandler.go`.
- The feature is initialized in `feature/realtime/feature.go`.

## Known Issues

As mentioned in `feature/realtime/feature.go`:

This service can scale to multiple pods. If a user connects via WebSocket to pod A, but pod B receives a SendMessage command, pod B will attempt to send the message to the user via WebSocket, but pod B doesn't have the connection. This will lead to an error.

This scaling issue needs to be addressed for proper multi-instance deployment.

## Future Improvements

1. Resolve the scaling issue mentioned above.
2. Implement authentication and authorization for secure communications.
3. Add more robust error handling and logging.
4. Develop client-side libraries for easier integration.