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


from generate_dicom_mri import calculate_dimensions
import math


def test_calculate_dimensions_basic():
    """Test basic dimension calculation."""
    total_bytes = 1024 * 1024 * 100  # 100 MB
    num_images = 10
    width, height = calculate_dimensions(total_bytes, num_images)

    # Should return square dimensions
    assert width == height
    # Should be multiple of 256 or close to sqrt of pixels
    assert width > 0 and height > 0


def test_calculate_dimensions_large_file():
    """Test with 4.5GB / 120 images."""
    total_bytes = int(4.5 * 1024 * 1024 * 1024)
    num_images = 120
    width, height = calculate_dimensions(total_bytes, num_images)

    # Check dimensions are reasonable for MRI
    assert width >= 512
    assert width == height

    # Verify size is close to target (within 10%)
    metadata_overhead = 100 * 1024  # 100KB
    pixel_bytes = (total_bytes - metadata_overhead)
    expected_pixels = pixel_bytes // 2  # 2 bytes per pixel (uint16)
    expected_per_frame = expected_pixels // num_images
    actual_per_frame = width * height

    tolerance = 0.1
    assert abs(actual_per_frame - expected_per_frame) / expected_per_frame < tolerance


def test_calculate_dimensions_rounds_to_reasonable():
    """Test that dimensions are rounded to reasonable values."""
    total_bytes = 1024 * 1024 * 50  # 50 MB
    num_images = 5
    width, height = calculate_dimensions(total_bytes, num_images)

    # Should be multiple of 256 or close
    assert width % 256 == 0 or (width % 128 == 0)
