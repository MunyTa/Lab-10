import pytest
from main import app

@pytest.fixture
def test_client():
    from fastapi.testclient import TestClient
    return TestClient(app)