import pytest
import sys
sys.path.insert(0, '.')
from generate_dicom_mri import parse_size


def test_parse_size_kilobytes():
    assert parse_size("100KB") == 100 * 1024


def test_parse_size_megabytes():
    assert parse_size("50MB") == 50 * 1024 * 1024


def test_parse_size_gigabytes():
    assert parse_size("4.5GB") == int(4.5 * 1024 * 1024 * 1024)


def test_parse_size_case_insensitive():
    assert parse_size("10mb") == 10 * 1024 * 1024
    assert parse_size("10Mb") == 10 * 1024 * 1024


def test_parse_size_invalid_format():
    with pytest.raises(ValueError):
        parse_size("invalid")


def test_parse_size_invalid_unit():
    with pytest.raises(ValueError):
        parse_size("100TB")
