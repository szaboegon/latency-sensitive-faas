import os
import pytest
from unittest.mock import patch, MagicMock
from unittest import mock
from typing import Any, Generator
from route import RoutingTable
import sys

@pytest.fixture(autouse=True)
def patch_tracing_before_import(monkeypatch: Any) -> None:
    # Ensure module is not already imported
    if "func" in sys.modules:
        del sys.modules["func"]

    monkeypatch.setattr("tracing.instrument_app", MagicMock(return_value=MagicMock()))

@pytest.fixture
def mock_context() -> MagicMock:
    """Fixture to create a mock context."""
    context = MagicMock()
    context.request.headers = {"X-Forward-To": "component1"}
    context.request.json = {
        "value": "test",
    }
    return context


@pytest.fixture
def mock_routing_table() -> RoutingTable:
    """Fixture to provide a mock routing table."""
    return {
        "component1": [
            {"component": "component2", "url": "local"},
            {"component": "component3", "url": "http://example.com"}
        ],
        "component2": []
    }


@pytest.fixture
def mock_env_vars(monkeypatch: Any) -> Generator[None, None, None]:
    with mock.patch.dict(os.environ, clear=True):
            envvars = {
                "NODE_IP": "192.168.100.1",
                "FUNCTION_NAME": "test_func",
                "APP_NAME": "ar13jaksdh21",
            }
            for k, v in envvars.items():
                monkeypatch.setenv(k, v)
            yield


@patch("func.read_config")
@patch("func.handle_request")
@patch("func.forward_request")
def test_main_successful_execution(
    mock_forward_request: MagicMock,
    mock_handle_request: MagicMock,
    mock_read_config: MagicMock,
    mock_context: MagicMock,
    mock_routing_table: RoutingTable,
    mock_env_vars: None
) -> None:
    from func import main
    # Mock the routing table
    mock_read_config.return_value = mock_routing_table

    # Mock handle_request to return dummy output and span context
    mock_handle_request.side_effect = [
        ({"key": "value"}, "span_context1"),
        ({"key2": "value2"}, "span_context2"),
    ]

    # Call the main function
    result = main(mock_context)

    # Assertions
    assert result == ("ok", 200)
    mock_read_config.assert_called_once()
    assert mock_handle_request.call_count == 2
    mock_forward_request.assert_called_once_with(
        {"component": "component3", "url": "http://example.com"},
        {"key": "value"},
        "span_context1"
    )


@patch("func.read_config")
def test_main_no_component_in_headers(
    mock_read_config: MagicMock,
    mock_context: MagicMock,
    mock_env_vars: None
) -> None:
    from func import main
    # Remove "X-Forward-To" header
    mock_context.request.headers = {}

    # Call the main function
    result = main(mock_context)

    # Assertions
    assert result == ("No component found with name None", 400)
    mock_read_config.assert_not_called()


@patch("func.read_config")
@patch("func.handle_request")
def test_main_invalid_routing_table(
    mock_handle_request: MagicMock,
    mock_read_config: MagicMock,
    mock_context: MagicMock,
    mock_env_vars: None
) -> None:
    from func import main
    # Mock the routing table with an invalid structure
    mock_read_config.return_value = {
        "component1": [
            {"url": "local"}  # Missing "component" key
        ]
    }

    # Mock handle_request to return dummy output and span context
    mock_handle_request.return_value = ({"key": "value"}, "span_context1")

    # Call the main function
    result = main(mock_context)

    # Assertions
    assert result == ("Invalid routing table entry: 'component' is missing", 400)  # Adjusted to match expected behavior
    mock_read_config.assert_called_once()
    mock_handle_request.assert_called_once()


@patch("func.read_config")
@patch("func.handle_request")
def test_main_empty_routing_table(
    mock_handle_request: MagicMock,
    mock_read_config: MagicMock,
    mock_context: MagicMock,
    mock_env_vars: None
) -> None:
    from func import main
    # Mock an empty routing table
    mock_read_config.return_value = {}
    mock_handle_request.return_value = ({"key": "value"}, "span_context1")

    # Call the main function
    result = main(mock_context)

    # Assertions
    assert result == ("ok", 200)
    mock_read_config.assert_called_once()
    mock_handle_request.assert_called_once()


@patch("func.read_config")
@patch("func.handle_request")
@patch("func.forward_request")
def test_main_handler_returns_list_of_dicts(
    mock_forward_request: MagicMock,
    mock_handle_request: MagicMock,
    mock_read_config: MagicMock,
    mock_context: MagicMock,
    mock_routing_table: RoutingTable,
    mock_env_vars: None
) -> None:
    from func import main
    mock_read_config.return_value = mock_routing_table

    # Mock handle_request to return a list of dicts and a span context
    mock_handle_request.side_effect = [
        ([{"key": "value"}, {"key2": "value2"}], "span_context1"),  # component1
        ([{"key3": "value3"}], "span_context2"),                    # component2 for first dict
        ([{"key4": "value4"}], "span_context3"),                    # component2 for second dict
    ]

    result = main(mock_context)

    assert result == ("ok", 200)
    mock_read_config.assert_called_once()
    assert mock_handle_request.call_count == 3
    assert mock_forward_request.call_count == 2
    mock_forward_request.assert_any_call(
        {"component": "component3", "url": "http://example.com"},
        {"key": "value"},
        "span_context1"
    )
    mock_forward_request.assert_any_call(
        {"component": "component3", "url": "http://example.com"},
        {"key2": "value2"},
        "span_context1"
    )


@patch("func.read_config")
@patch("func.handle_request")
@patch("func.forward_request")
def test_main_handler_returns_list_of_bytes(
    mock_forward_request: MagicMock,
    mock_handle_request: MagicMock,
    mock_read_config: MagicMock,
    mock_context: MagicMock,
    mock_routing_table: RoutingTable,
    mock_env_vars: None
) -> None:
    from func import main
    mock_read_config.return_value = mock_routing_table

    # Mock handle_request to return a list of bytes and a span context
    mock_handle_request.side_effect = [
        ([b"abc", b"def"], "span_context1"),  # component1
        ([b"ghi"], "span_context2"),          # component2 (from b"abc")
        ([b"jkl"], "span_context3"),          # component2 (from b"def")
    ]

    result = main(mock_context)

    assert result == ("ok", 200)
    mock_read_config.assert_called_once()
    assert mock_handle_request.call_count == 3
    
    assert mock_forward_request.call_count == 2
    mock_forward_request.assert_any_call(
        {"component": "component3", "url": "http://example.com"},
        b"abc",
        "span_context1"
    )
    mock_forward_request.assert_any_call(
        {"component": "component3", "url": "http://example.com"},
        b"def",
        "span_context1"
    )