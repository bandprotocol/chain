
import pytest

from .util import convert_proto_str_to_object


# Test Cases
@pytest.mark.parametrize(
    "input_str, expected",
    [
        # Simple input with double quotes and unquoted value
        (
            'signal_id:"test value" encoder:ENCODER_FIXED_POINT_ABI ',
            {"signal_id": "test value", "encoder": "ENCODER_FIXED_POINT_ABI"},
        ),
        # Unquoted values
        (
            "key1:value1 key2:value2 ",
            {"key1": "value1", "key2": "value2"},
        ),
        # Multiple types of quotes
        (
            'key1:value1 key2:"value 2" key3:\'value 3\' key4:value4',
            {"key1": "value1", "key2": "value 2", "key3": "value 3", "key4": "value4"},
        ),
        # Empty input
        (
            "",
            {},
        ),
    ],
)

def test_convert_proto_str_to_object(input_str, expected):
    assert convert_proto_str_to_object(input_str) == expected
