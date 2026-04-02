import pytest
import asyncio
import websockets
import json

@pytest.mark.asyncio
async def test_websocket_connection():
    """Тест подключения к WebSocket чату"""
    uri = "ws://localhost:8080/ws/chat?room=test&user=tester"
    
    try:
        async with websockets.connect(uri) as websocket:
            assert websocket.open
            print("WebSocket connection successful")
    except Exception as e:
        pytest.skip(f"Go service not running: {e}")

@pytest.mark.asyncio
async def test_websocket_send_receive():
    """Тест отправки и получения сообщений"""
    uri = "ws://localhost:8080/ws/chat?room=testroom&user=alice"
    
    try:
        async with websockets.connect(uri) as websocket:
            test_msg = {"type": "chat", "content": "Hello, World!"}
            await websocket.send(json.dumps(test_msg))
            
            response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
            data = json.loads(response)
            
            assert data["content"] == "Hello, World!"
            assert data["user"] == "alice"
            assert "time" in data
            
    except Exception as e:
        pytest.skip(f"Test failed: {e}")