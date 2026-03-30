import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from unittest.mock import patch, MagicMock
from solution.client import get_user, User


def test_get_user_returns_user_object():
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.json.return_value = {"id": 1, "name": "Alice", "email": "alice@example.com"}
    with patch("solution.client.requests.get", return_value=mock_response):
        user = get_user(1)
    assert isinstance(user, User)
    assert user.id == 1
    assert user.name == "Alice"


def test_get_user_correct_url():
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.json.return_value = {"id": 42, "name": "Bob", "email": "bob@example.com"}
    with patch("solution.client.requests.get", return_value=mock_response) as mock_get:
        get_user(42)
    mock_get.assert_called_once_with("http://api.example.internal/users/42")


def test_get_user_raises_on_404():
    mock_response = MagicMock()
    mock_response.status_code = 404
    with patch("solution.client.requests.get", return_value=mock_response):
        with pytest.raises(ValueError, match="User 99 not found"):
            get_user(99)
