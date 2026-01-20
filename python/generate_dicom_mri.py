#!/usr/bin/env python3
"""
DICOM MRI Generator
Generate valid DICOM multi-frame MRI files for testing medical interfaces.
"""

import re
import math
import hashlib
import glob
import pydicom
from pydicom.dataset import Dataset, FileMetaDataset
from pydicom.uid import generate_uid, ExplicitVRLittleEndian
from datetime import datetime
import random
import numpy as np
import argparse
import sys
import os
from pydicom.fileset import FileSet
from PIL import Image, ImageDraw, ImageFont


def generate_deterministic_uid(seed_string):
    """
    Generate a deterministic DICOM UID from a seed string.

    Args:
        seed_string: String to use as seed for UID generation

    Returns:
        str: Valid DICOM UID (max 64 chars, no leading zeros in components)
    """
    # Use pydicom's prefix for compatibility
    # DICOM UID format: prefix.suffix where suffix is numeric
    prefix = "1.2.826.0.1.3680043.8.498"

    # Generate a hash from the seed string
    hash_obj = hashlib.sha256(seed_string.encode())
    hash_hex = hash_obj.hexdigest()

    # Convert hash to numeric string (take first 30 hex chars to keep UID shorter)
    numeric_suffix = str(int(hash_hex[:30], 16))

    # Create segments, ensuring no segment starts with 0 (DICOM requirement)
    # Split into 10-digit segments
    segments = []
    for i in range(0, len(numeric_suffix), 10):
        segment = numeric_suffix[i:i+10]
        # Ensure segment doesn't start with 0 (unless it's just "0")
        if segment and segment != "0" and segment[0] == "0":
            segment = segment.lstrip("0") or "1"  # Remove leading zeros, use "1" if all zeros
        if segment:
            segments.append(segment)
        if len(segments) >= 3:  # Limit to 3 segments after prefix
            break

    suffix = '.'.join(segments)
    uid = f"{prefix}.{suffix}"

    # Ensure UID is not too long (max 64 chars)
    if len(uid) > 64:
        # Truncate last segment to fit
        uid = uid[:63]
        # Remove trailing dot if any
        uid = uid.rstrip('.')

    return uid


def generate_patient_name(sex):
    """
    Generate a realistic patient name based on sex.

    Args:
        sex: 'M' or 'F'

    Returns:
        str: Patient name in DICOM format (LASTNAME^FIRSTNAME)
    """
    # Prénoms masculins
    male_first_names = [
        'Jean', 'Pierre', 'Michel', 'André', 'Philippe', 'Alain', 'Bernard', 'Jacques',
        'François', 'Christian', 'Daniel', 'Patrick', 'Nicolas', 'Olivier', 'Laurent',
        'Thierry', 'Stéphane', 'Éric', 'David', 'Julien', 'Christophe', 'Pascal',
        'Sébastien', 'Marc', 'Vincent', 'Antoine', 'Alexandre', 'Maxime', 'Thomas',
        'Lucas', 'Hugo', 'Louis', 'Arthur', 'Gabriel', 'Raphaël', 'Paul', 'Jules'
    ]

    # Prénoms féminins
    female_first_names = [
        'Marie', 'Nathalie', 'Isabelle', 'Sylvie', 'Catherine', 'Françoise', 'Valérie',
        'Christine', 'Monique', 'Sophie', 'Patricia', 'Martine', 'Nicole', 'Sandrine',
        'Stéphanie', 'Céline', 'Julie', 'Aurélie', 'Caroline', 'Laurence', 'Émilie',
        'Claire', 'Anne', 'Camille', 'Laura', 'Sarah', 'Manon', 'Emma', 'Léa',
        'Chloé', 'Zoé', 'Alice', 'Charlotte', 'Lucie', 'Juliette', 'Louise'
    ]

    # Noms de famille (neutres)
    last_names = [
        'Martin', 'Bernard', 'Dubois', 'Thomas', 'Robert', 'Richard', 'Petit',
        'Durand', 'Leroy', 'Moreau', 'Simon', 'Laurent', 'Lefebvre', 'Michel',
        'Garcia', 'David', 'Bertrand', 'Roux', 'Vincent', 'Fournier', 'Morel',
        'Girard', 'André', 'Lefevre', 'Mercier', 'Dupont', 'Lambert', 'Bonnet',
        'François', 'Martinez', 'Legrand', 'Garnier', 'Faure', 'Rousseau', 'Blanc',
        'Guerin', 'Muller', 'Henry', 'Roussel', 'Nicolas', 'Perrin', 'Morin',
        'Mathieu', 'Clement', 'Gauthier', 'Dumont', 'Lopez', 'Fontaine', 'Chevalier',
        'Robin', 'Masson', 'Sanchez', 'Gerard', 'Nguyen', 'Boyer', 'Denis', 'Lemaire'
    ]

    if sex == 'M':
        first_name = random.choice(male_first_names)
    else:  # 'F'
        first_name = random.choice(female_first_names)

    last_name = random.choice(last_names)

    # DICOM format: LASTNAME^FIRSTNAME
    return f"{last_name}^{first_name}"


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


def calculate_dimensions(total_size_bytes, num_images):
    """
    Calculate optimal image dimensions to hit target file size.

    Args:
        total_size_bytes: Target total file size in bytes
        num_images: Number of frames/images

    Returns:
        tuple: (width, height) as integers
    """
    # Estimate metadata overhead
    metadata_overhead = 100 * 1024  # 100KB

    # Available bytes for pixel data
    available_bytes = total_size_bytes - metadata_overhead

    # DICOM limit: pixel data must be < 2^32 bytes (4,294,967,296)
    # The length field is 32-bit unsigned, so max is 2^32 - 1
    # Use a safe margin of 10MB below the limit
    MAX_PIXEL_DATA_SIZE = (2**32 - 1) - (10 * 1024 * 1024)  # ~4.28 GB with safety margin

    # If requested size exceeds DICOM limit, cap it
    if available_bytes > MAX_PIXEL_DATA_SIZE:
        available_bytes = MAX_PIXEL_DATA_SIZE
        print(f"Attention: Taille limitée à 4 GB (limite DICOM pour pixel data)")

    # Calculate pixels (2 bytes per pixel for uint16)
    bytes_per_pixel = 2
    total_pixels = available_bytes // bytes_per_pixel
    pixels_per_frame = total_pixels // num_images

    # Calculate square dimension
    dim = int(math.sqrt(pixels_per_frame))

    # Round DOWN to nearest multiple of 256 for realistic MRI dimensions
    # Important: must round down to ensure we don't exceed size limit
    if dim >= 256:
        dim = (dim // 256) * 256  # Floor division to round down
    elif dim >= 128:
        dim = (dim // 128) * 128

    # Ensure minimum size
    dim = max(dim, 128)

    return dim, dim


def generate_metadata(num_images, width, height, instance_number=None, study_uid=None, series_uid=None,
                      patient_id=None, patient_name=None, patient_birth_date=None, patient_sex=None,
                      study_date=None, study_time=None, study_id=None, study_description=None,
                      accession_number=None, series_number=1,
                      pixel_spacing=None, slice_thickness=None, spacing_between_slices=None,
                      echo_time=None, repetition_time=None, flip_angle=None, sequence_name=None,
                      manufacturer=None, model=None, field_strength=None):
    """
    Generate DICOM dataset with realistic MRI metadata.

    Args:
        num_images: Number of frames (used for series info, but each file has 1 frame)
        width: Image width in pixels
        height: Image height in pixels
        instance_number: Instance number for this image (1-based)
        study_uid: Shared Study Instance UID (if None, generates new)
        series_uid: Shared Series Instance UID (if None, generates new)
        patient_id: Shared Patient ID (if None, generates new)
        patient_name: Shared Patient Name (if None, generates new)
        patient_birth_date: Shared Patient Birth Date (if None, generates new)
        patient_sex: Shared Patient Sex (if None, generates new)
        study_date: Shared Study Date (if None, uses current date)
        study_time: Shared Study Time (if None, uses current time)
        study_id: Shared Study ID (if None, generates new)
        study_description: Shared Study Description (if None, generates new)
        accession_number: Shared Accession Number (if None, generates new)
        series_number: Series Number (default: 1)

    Returns:
        pydicom.Dataset: Dataset with metadata
    """
    # Create file meta information
    file_meta = FileMetaDataset()
    file_meta.TransferSyntaxUID = ExplicitVRLittleEndian
    file_meta.MediaStorageSOPClassUID = '1.2.840.10008.5.1.4.1.1.4'  # MR Image Storage
    file_meta.MediaStorageSOPInstanceUID = generate_uid()
    file_meta.ImplementationClassUID = generate_uid()

    # Create main dataset
    ds = Dataset()
    ds.file_meta = file_meta

    # Specify character set for proper encoding of accented characters
    # ISO_IR 192 = UTF-8
    ds.SpecificCharacterSet = 'ISO_IR 192'

    # Patient information (shared across all instances for same patient)
    ds.PatientName = patient_name if patient_name else f"TEST^PATIENT^{random.randint(1000, 9999)}"
    ds.PatientID = patient_id if patient_id else f"PID{random.randint(100000, 999999)}"
    ds.PatientBirthDate = patient_birth_date if patient_birth_date else f"{random.randint(1950, 2000):04d}{random.randint(1, 12):02d}{random.randint(1, 28):02d}"
    ds.PatientSex = patient_sex if patient_sex else random.choice(['M', 'F'])

    # Study information (shared across all instances in same study)
    ds.StudyInstanceUID = study_uid if study_uid else generate_uid()
    if study_date and study_time:
        ds.StudyDate = study_date
        ds.StudyTime = study_time
    else:
        now = datetime.now()
        ds.StudyDate = now.strftime('%Y%m%d')
        ds.StudyTime = now.strftime('%H%M%S')
    ds.StudyID = study_id if study_id else f"STD{random.randint(1000, 9999)}"
    ds.StudyDescription = study_description if study_description else "Test MRI Study"
    ds.AccessionNumber = accession_number if accession_number else f"ACC{random.randint(100000, 999999)}"

    # Series information (shared across all instances in same series)
    ds.SeriesInstanceUID = series_uid if series_uid else generate_uid()
    ds.SeriesNumber = series_number
    ds.SeriesDescription = f"Test MRI Series - {num_images} images"
    ds.Modality = 'MR'

    # Instance number (position in series)
    if instance_number is not None:
        ds.InstanceNumber = instance_number

    # SOP Common
    ds.SOPClassUID = file_meta.MediaStorageSOPClassUID
    ds.SOPInstanceUID = file_meta.MediaStorageSOPInstanceUID

    # MRI-specific parameters (shared across all images in series)
    if manufacturer is None or model is None or field_strength is None:
        manufacturers = [
            ('SIEMENS', 'Avanto', 1.5),
            ('SIEMENS', 'Skyra', 3.0),
            ('GE MEDICAL SYSTEMS', 'Signa HDxt', 1.5),
            ('GE MEDICAL SYSTEMS', 'Discovery MR750', 3.0),
            ('PHILIPS', 'Achieva', 1.5),
            ('PHILIPS', 'Ingenia', 3.0)
        ]
        manufacturer, model, field_strength = random.choice(manufacturers)

    ds.Manufacturer = manufacturer
    ds.ManufacturerModelName = model
    ds.MagneticFieldStrength = field_strength

    # Calculate imaging frequency based on field strength
    # 1.5T ≈ 63.87 MHz, 3.0T ≈ 127.74 MHz for protons
    ds.ImagingFrequency = field_strength * 42.58  # MHz (gyromagnetic ratio)

    # Sequence parameters (realistic T1-weighted values) - MUST be same for all images in series!
    ds.EchoTime = echo_time if echo_time is not None else random.uniform(10, 30)  # ms
    ds.RepetitionTime = repetition_time if repetition_time is not None else random.uniform(400, 800)  # ms
    ds.FlipAngle = flip_angle if flip_angle is not None else random.uniform(60, 90)  # degrees
    ds.SliceThickness = slice_thickness if slice_thickness is not None else random.uniform(1.0, 5.0)  # mm
    ds.SpacingBetweenSlices = spacing_between_slices if spacing_between_slices is not None else (ds.SliceThickness + random.uniform(0, 0.5))  # mm
    ds.SequenceName = sequence_name if sequence_name is not None else random.choice(['T1_MPRAGE', 'T1_SE', 'T2_FSE', 'T2_FLAIR'])

    # Image parameters (single frame per file)
    ds.SamplesPerPixel = 1
    ds.PhotometricInterpretation = 'MONOCHROME2'
    ds.Rows = height
    ds.Columns = width
    ds.BitsAllocated = 16
    ds.BitsStored = 16
    ds.HighBit = 15
    ds.PixelRepresentation = 0  # unsigned

    # Pixel spacing (typical MRI: 0.5-2mm) - MUST be same for all images in series!
    if pixel_spacing is None:
        pixel_spacing = random.uniform(0.5, 2.0)
    ds.PixelSpacing = [pixel_spacing, pixel_spacing]

    # Image Position and Orientation (for 3D reconstruction)
    if instance_number is not None:
        # Position changes along Z axis for each slice
        slice_position = (instance_number - 1) * ds.SliceThickness
        ds.ImagePositionPatient = [0.0, 0.0, slice_position]
        # Standard axial orientation
        ds.ImageOrientationPatient = [1.0, 0.0, 0.0, 0.0, 1.0, 0.0]
        # Slice Location - critical for medical viewers to create scrollable stack
        ds.SliceLocation = slice_position

    # Window Center and Window Width (for display)
    # These are critical for image visualization in medical viewers
    # For 12-bit data (0-4095), use middle of range
    ds.WindowCenter = "2048"  # Middle of 0-4095 range
    ds.WindowWidth = "4096"   # Full range

    # Can also provide as list for multiple window presets
    ds.WindowCenterWidthExplanation = "Full Range"

    # Rescale intercept and slope (for Hounsfield units in CT, identity for MR)
    ds.RescaleIntercept = "0"
    ds.RescaleSlope = "1"
    ds.RescaleType = "US"  # Unspecified

    return ds


def generate_pixel_data(num_images, width, height, seed=None):
    """
    Generate random pixel data for MRI images.

    Args:
        num_images: Number of frames to generate
        width: Image width
        height: Image height
        seed: Optional random seed for reproducibility

    Returns:
        numpy.ndarray: Array of shape (num_images, height, width) with dtype uint16
    """
    if seed is not None:
        np.random.seed(seed)

    # Generate random noise in 12-bit range (0-4095) - typical for MRI
    # Shape: (num_images, height, width)
    pixel_data = np.random.randint(0, 4096, size=(num_images, height, width), dtype=np.uint16)

    return pixel_data


def generate_single_image(width, height, seed=None, image_number=None, total_images=None, font=None):
    """
    Generate random pixel data for a single MRI image with optional text overlay.

    Args:
        width: Image width
        height: Image height
        seed: Optional random seed for reproducibility
        image_number: Current image number (for text overlay)
        total_images: Total number of images (for text overlay)
        font: Pre-loaded PIL font (to avoid reloading on each image)

    Returns:
        numpy.ndarray: Array of shape (height, width) with dtype uint16
    """
    if seed is not None:
        np.random.seed(seed)

    # Generate random noise in 12-bit range (0-4095) - typical for MRI
    pixel_data = np.random.randint(0, 4096, size=(height, width), dtype=np.uint16)

    # Add text overlay if image number is provided
    if image_number is not None and total_images is not None:
        # Convert to PIL Image for text drawing
        # Scale from 0-4095 to 0-65535 (16-bit) for better contrast
        img_scaled = (pixel_data.astype(np.uint32) * 16).astype(np.uint16)
        img_pil = Image.fromarray(img_scaled, mode='I;16')

        # Convert to RGB for drawing (easier to draw text)
        img_rgb = img_pil.convert('RGB')
        draw = ImageDraw.Draw(img_rgb)

        # Text to draw
        text = f"File {image_number}/{total_images}"

        # Use pre-loaded font if provided, otherwise use default
        if font is None:
            font = ImageFont.load_default()

        # Get text bounding box for centering
        bbox = draw.textbbox((0, 0), text, font=font)
        text_width = bbox[2] - bbox[0]
        text_height = bbox[3] - bbox[1]

        # Calculate position: centered horizontally, near top
        padding_top = int(height * 0.05)  # 5% from top
        x = (width - text_width) // 2  # Center horizontally
        y = padding_top

        # Draw text with white color and thick black outline for visibility
        outline_color = (0, 0, 0)
        text_color = (255, 255, 255)
        outline_thickness = max(3, int(text_height * 0.05))  # 5% of text height

        # Draw thick outline
        for dx in range(-outline_thickness, outline_thickness + 1):
            for dy in range(-outline_thickness, outline_thickness + 1):
                if dx != 0 or dy != 0:  # Skip center
                    draw.text((x + dx, y + dy), text, font=font, fill=outline_color)

        # Draw main text
        draw.text((x, y), text, font=font, fill=text_color)

        # Convert back to grayscale and then to uint16
        img_gray = img_rgb.convert('L')
        pixel_data_with_text = np.array(img_gray, dtype=np.uint16)

        # Scale back to 12-bit range (0-4095)
        pixel_data = (pixel_data_with_text * 16).astype(np.uint16)
        # Clip to ensure we stay in 12-bit range
        pixel_data = np.clip(pixel_data, 0, 4095)

    return pixel_data


def parse_arguments(argv=None):
    """
    Parse command line arguments.

    Args:
        argv: List of arguments (for testing), None uses sys.argv

    Returns:
        argparse.Namespace: Parsed arguments
    """
    parser = argparse.ArgumentParser(
        description='Générer des fichiers DICOM d\'IRM multi-frame pour tests',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Exemples:
  # Générer 120 images dans 1 étude (défaut)
  %(prog)s --num-images 120 --total-size 4.5GB

  # Générer 120 images réparties dans 3 études
  %(prog)s --num-images 120 --total-size 1GB --num-studies 3

  # Avec seed pour reproductibilité
  %(prog)s --num-images 50 --total-size 1GB --output patient-test --seed 42
        """
    )

    parser.add_argument(
        '--num-images',
        type=int,
        required=True,
        help='Nombre d\'images/coupes dans la série'
    )

    parser.add_argument(
        '--total-size',
        type=str,
        required=True,
        help='Taille totale cible (ex: 100MB, 4.5GB)'
    )

    parser.add_argument(
        '--output',
        type=str,
        default='dicom_series',
        help='Nom du dossier de sortie (défaut: dicom_series)'
    )

    parser.add_argument(
        '--seed',
        type=int,
        default=None,
        help='Seed pour la génération aléatoire (reproductibilité)'
    )

    parser.add_argument(
        '--num-studies',
        type=int,
        default=1,
        help='Nombre d\'études (studies) à générer (défaut: 1). Les images seront réparties équitablement.'
    )

    args = parser.parse_args(argv)

    # Validate num_images
    if args.num_images <= 0:
        parser.error("--num-images doit être > 0")

    # Validate num_studies
    if args.num_studies <= 0:
        parser.error("--num-studies doit être > 0")

    if args.num_studies > args.num_images:
        parser.error(f"--num-studies ({args.num_studies}) ne peut pas être supérieur à --num-images ({args.num_images})")

    return args


def format_bytes(bytes_size):
    """Format bytes as human-readable string."""
    for unit in ['B', 'KB', 'MB', 'GB']:
        if bytes_size < 1024.0:
            return f"{bytes_size:.2f} {unit}"
        bytes_size /= 1024.0
    return f"{bytes_size:.2f} TB"


def main():
    """Main entry point."""
    # Parse arguments
    args = parse_arguments()

    try:
        # Parse and validate size
        print("Calcul de la résolution optimale...")
        total_bytes = parse_size(args.total_size)

        if total_bytes <= 0:
            print(f"Erreur: La taille doit être > 0", file=sys.stderr)
            return 1

        # Check disk space
        stat = os.statvfs('.')
        available_space = stat.f_bavail * stat.f_frsize
        if total_bytes > available_space:
            print(f"Erreur: Espace disque insuffisant. Requis: {format_bytes(total_bytes)}, Disponible: {format_bytes(available_space)}", file=sys.stderr)
            return 1

        # Calculate dimensions
        width, height = calculate_dimensions(total_bytes, args.num_images)

        # Estimate actual file size
        pixel_bytes = args.num_images * width * height * 2  # 2 bytes per pixel
        metadata_overhead = 100 * 1024  # 100KB estimate
        estimated_size = pixel_bytes + metadata_overhead

        print(f"Résolution: {width}x{height} pixels par image")
        print(f"Taille estimée: {format_bytes(estimated_size)} ({args.num_images} images)")

        # Create output directory
        output_dir = args.output
        if os.path.exists(output_dir):
            print(f"Attention: Le dossier {output_dir} existe déjà")
        else:
            os.makedirs(output_dir)
            print(f"Création du dossier: {output_dir}")

        # Set seed for reproducibility
        # If no seed specified, generate one from output directory name for consistency
        if args.seed is not None:
            seed = args.seed
            print(f"Utilisation du seed: {seed}")
        else:
            # Generate deterministic seed from output directory name
            # This ensures same directory name = same patient/study IDs
            seed = abs(hash(output_dir)) % (2**31)
            print(f"Génération automatique du seed basé sur '{output_dir}': {seed}")
            print("  (même dossier = mêmes IDs patient/study)")

        np.random.seed(seed)
        random.seed(seed)

        # Generate shared Patient info for all studies (same patient, different exams)
        patient_id = f"PID{random.randint(100000, 999999)}"
        patient_sex = random.choice(['M', 'F'])
        patient_name = generate_patient_name(patient_sex)  # Generate realistic name based on sex
        patient_birth_date = f"{random.randint(1950, 2000):04d}{random.randint(1, 12):02d}{random.randint(1, 28):02d}"

        print(f"Génération de {args.num_images} fichiers DICOM...")
        print(f"Patient: {patient_name} (ID: {patient_id}, né le {patient_birth_date}, sexe {patient_sex})")
        print(f"Nombre d'études: {args.num_studies}")

        # Calculate images per study
        images_per_study = args.num_images // args.num_studies
        remaining_images = args.num_images % args.num_studies

        total_size = 0
        global_image_index = 1

        print("Chargement du font...")
        # Load font once for all images (much faster than loading per image)
        font_size = int(width / 16)  # Large font that's still performant
        script_dir = os.path.dirname(os.path.abspath(__file__))
        font_paths = [
            # Font in project directory (portable solution)
            os.path.join(script_dir, "DejaVuSans-Bold.ttf"),
            # Standard Linux paths
            "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
            "/usr/share/fonts/TTF/DejaVuSans-Bold.ttf",
            "/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
        ]

        font = None
        for font_path in font_paths:
            try:
                if os.path.exists(font_path):
                    font = ImageFont.truetype(font_path, font_size)
                    print(f"✓ Loaded font: {font_path} (size: {font_size}px)")
                    break
            except Exception as e:
                continue

        if font is None:
            print(f"Warning: Could not load TrueType font, using default (text may be small)")
            font = ImageFont.load_default()

        # Generate DICOM files for each study
        for study_num in range(1, args.num_studies + 1):
            # Generate DETERMINISTIC UIDs for this study based on output_dir and study_num
            # This ensures same directory + same study number = same UIDs!
            study_uid = generate_deterministic_uid(f"{output_dir}_study_{study_num}")
            series_uid = generate_deterministic_uid(f"{output_dir}_study_{study_num}_series_1")

            # Generate study-specific info (same for all images in this study)
            study_date = datetime.now().strftime('%Y%m%d')
            study_time = datetime.now().strftime('%H%M%S')
            study_id = f"STD{random.randint(1000, 9999)}"
            study_description = f"Brain MRI - Study {study_num}" if args.num_studies > 1 else "Brain MRI"
            accession_number = f"ACC{random.randint(100000, 999999)}"

            # Generate series-specific acquisition parameters (SAME for all images in series!)
            # These MUST be identical for all images to be grouped together in DICOM viewers
            series_pixel_spacing = random.uniform(0.5, 2.0)
            series_slice_thickness = random.uniform(1.0, 5.0)
            series_spacing_between_slices = series_slice_thickness + random.uniform(0, 0.5)
            series_echo_time = random.uniform(10, 30)
            series_repetition_time = random.uniform(400, 800)
            series_flip_angle = random.uniform(60, 90)
            series_sequence_name = random.choice(['T1_MPRAGE', 'T1_SE', 'T2_FSE', 'T2_FLAIR'])

            # MRI scanner info (same for all images in series)
            manufacturers = [
                ('SIEMENS', 'Avanto', 1.5),
                ('SIEMENS', 'Skyra', 3.0),
                ('GE MEDICAL SYSTEMS', 'Signa HDxt', 1.5),
                ('GE MEDICAL SYSTEMS', 'Discovery MR750', 3.0),
                ('PHILIPS', 'Achieva', 1.5),
                ('PHILIPS', 'Ingenia', 3.0)
            ]
            series_manufacturer, series_model, series_field_strength = random.choice(manufacturers)

            # Calculate how many images for this study
            # Distribute remaining images to first studies
            num_images_this_study = images_per_study + (1 if study_num <= remaining_images else 0)

            print(f"\nÉtude {study_num}/{args.num_studies}: {num_images_this_study} images")
            print(f"  StudyID: {study_id}, Description: {study_description}")
            print(f"  Scanner: {series_manufacturer} {series_model} ({series_field_strength}T)")
            print(f"  Paramètres: PixelSpacing={series_pixel_spacing:.2f}mm, SliceThickness={series_slice_thickness:.2f}mm")

            # Generate each DICOM file for this study
            for instance_in_study in range(1, num_images_this_study + 1):
                # Generate metadata for this instance
                ds = generate_metadata(
                    num_images=num_images_this_study,
                    width=width,
                    height=height,
                    instance_number=instance_in_study,
                    study_uid=study_uid,
                    series_uid=series_uid,
                    patient_id=patient_id,
                    patient_name=patient_name,
                    patient_birth_date=patient_birth_date,
                    patient_sex=patient_sex,
                    study_date=study_date,
                    study_time=study_time,
                    study_id=study_id,
                    study_description=study_description,
                    accession_number=accession_number,
                    series_number=1,
                    # Series acquisition parameters (SAME for all images!)
                    pixel_spacing=series_pixel_spacing,
                    slice_thickness=series_slice_thickness,
                    spacing_between_slices=series_spacing_between_slices,
                    echo_time=series_echo_time,
                    repetition_time=series_repetition_time,
                    flip_angle=series_flip_angle,
                    sequence_name=series_sequence_name,
                    manufacturer=series_manufacturer,
                    model=series_model,
                    field_strength=series_field_strength
                )

                # Generate pixel data for this single image with text overlay
                pixel_data = generate_single_image(
                    width,
                    height,
                    image_number=global_image_index,
                    total_images=args.num_images,
                    font=font
                )

                # Add pixel data to dataset
                ds.PixelData = pixel_data.tobytes()

                # Write DICOM file
                filename = f"IMG{global_image_index:04d}.dcm"
                filepath = os.path.join(output_dir, filename)
                ds.save_as(filepath, write_like_original=False)

                total_size += os.path.getsize(filepath)

                # Progress indicator
                if global_image_index % 10 == 0 or global_image_index == args.num_images:
                    progress = (global_image_index / args.num_images) * 100
                    print(f"  Progression: {global_image_index}/{args.num_images} ({progress:.0f}%)")

                global_image_index += 1

        print(f"\n✓ {args.num_images} fichiers DICOM créés dans: {output_dir}/")
        print(f"  Taille totale: {format_bytes(total_size)}")
        if args.num_studies > 1:
            print(f"  Répartis en {args.num_studies} études (studies)")

        # Create DICOMDIR file
        print("\nCréation du fichier DICOMDIR...")
        try:
            # Create empty FileSet
            fs = FileSet()

            # Add all DICOM files to the fileset (using absolute paths)
            for i in range(1, args.num_images + 1):
                filename = f"IMG{i:04d}.dcm"
                filepath = os.path.join(output_dir, filename)
                fs.add(filepath)

            # Write DICOMDIR to the output directory
            # This will create the DICOMDIR file and the standard hierarchy
            # IMPORTANT: pydicom copies files into PT*/ST*/SE* hierarchy
            fs.write(output_dir)

            print(f"✓ DICOMDIR créé avec structure hiérarchique standard")

            # Remove original IMG*.dcm files from root (they're now in PT*/ST*/SE* hierarchy)
            print(f"\nNettoyage des fichiers temporaires...")
            removed_count = 0
            for i in range(1, args.num_images + 1):
                filename = f"IMG{i:04d}.dcm"
                filepath = os.path.join(output_dir, filename)
                if os.path.exists(filepath):
                    os.remove(filepath)
                    removed_count += 1

            print(f"✓ {removed_count} fichiers temporaires supprimés")
            print(f"\nLa série DICOM est prête à être importée!")
            print(f"Importez le dossier complet: {os.path.abspath(output_dir)}/")
            print(f"\nStructure DICOM standard créée:")
            print(f"  - DICOMDIR (fichier index)")
            print(f"  - PT000000/ST000000/SE000000/ (hiérarchie patient/study/series)")
            print(f"\nPour ouvrir dans un visualiseur DICOM (ex: Weasis):")
            print(f"  1. Ouvrez le dossier principal: {os.path.abspath(output_dir)}")
            print(f"  2. Le visualiseur devrait détecter automatiquement le DICOMDIR")

        except Exception as e:
            print(f"Attention: Erreur lors de la création du DICOMDIR: {e}")
            print(f"Les fichiers DICOM sont valides, mais le DICOMDIR n'a pas pu être créé.")

        return 0

    except ValueError as e:
        print(f"Erreur: {e}", file=sys.stderr)
        return 1
    except OSError as e:
        print(f"Erreur d'écriture: {e}", file=sys.stderr)
        return 1
    except Exception as e:
        print(f"Erreur inattendue: {e}", file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
