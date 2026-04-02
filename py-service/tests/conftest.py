import pytest
from main import app

@pytest.fixture
def test_client():
    """Фикстура для тестового клиента FastAPI"""
    from fastapi.testclient import TestClient
    return TestClient(app)