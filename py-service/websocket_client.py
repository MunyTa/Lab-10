import asyncio
import websockets
import json
from datetime import datetime

class ChatClient:
    def __init__(self, uri: str, room: str, user: str):
        self.uri = uri
        self.room = room
        self.user = user
        self.websocket = None
        
    async def connect(self):
        url = f"{self.uri}?room={self.room}&user={self.user}"
        self.websocket = await websockets.connect(url)
        print(f"Connected to chat room '{self.room}' as '{self.user}'")
        
    async def send_message(self, content: str):
        message = {"type": "chat", "content": content}
        await self.websocket.send(json.dumps(message))
        print(f"[{self.user}] {content}")
        
    async def receive_messages(self):
        try:
            async for message in self.websocket:
                data = json.loads(message)
                timestamp = data.get("time", datetime.now().strftime("%H:%M:%S"))
                user = data.get("user", "unknown")
                content = data.get("content", "")
                msg_type = data.get("type", "chat")
                
                if msg_type == "system":
                    print(f"\n[SYSTEM] {content}")
                else:
                    print(f"\n[{timestamp}] {user}: {content}")
        except websockets.exceptions.ConnectionClosed:
            print("Connection closed")
            
    async def close(self):
        if self.websocket:
            await self.websocket.close()
            print("Disconnected from chat")
            
async def main():
    WS_URL = "ws://localhost:8080/ws/chat"
    ROOM = "general"
    USER = "python_user"
    
    client = ChatClient(WS_URL, ROOM, USER)
    
    try:
        await client.connect()
        receive_task = asyncio.create_task(client.receive_messages())
        
        await asyncio.sleep(1)
        await client.send_message("Hello from Python client!")
        await asyncio.sleep(1)
        await client.send_message("This is a WebSocket chat test")
        await asyncio.sleep(1)
        await client.send_message("Working with Go server!")
        
        await asyncio.sleep(3)
        receive_task.cancel()
        
    except Exception as e:
        print(f"Error: {e}")
    finally:
        await client.close()

if __name__ == "__main__":
    asyncio.run(main())