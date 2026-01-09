#!/usr/bin/env python3
"""
DICOM MRI Generator
Generate valid DICOM multi-frame MRI files for testing medical interfaces.
"""

import re


def parse_size(size_str):
    """
    Parse size string (e.g., '4.5GB', '100MB') into bytes.

    Args:
        size_str: Size string with unit (KB, MB, GB)

    Returns:
        int: Size in bytes

    Raises:
        ValueError: If format is invalid or unit not supported
    """
    pattern = r'^(\d+(?:\.\d+)?)(KB|MB|GB)$'
    match = re.match(pattern, size_str.upper())

    if not match:
        raise ValueError(f"Format invalide: '{size_str}'. Utilisez format comme '100MB', '4.5GB'")

    value = float(match.group(1))
    unit = match.group(2)

    multipliers = {
        'KB': 1024,
        'MB': 1024 * 1024,
        'GB': 1024 * 1024 * 1024
    }

    if unit not in multipliers:
        raise ValueError(f"Unité non supportée: '{unit}'. Utilisez KB, MB ou GB")

    return int(value * multipliers[unit])
