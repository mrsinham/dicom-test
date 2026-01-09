# DICOM MRI Generator

Outil CLI Python pour générer des fichiers DICOM d'IRM multi-frame valides pour tester des interfaces médicales.

## Installation

```bash
pip install -r requirements.txt
```

## Usage

```bash
python generate_dicom_mri.py --num-images 120 --total-size 4.5GB --output mri_test.dcm
```

### Paramètres

- `--num-images` (requis): Nombre d'images/coupes dans la série
- `--total-size` (requis): Taille totale cible (KB, MB, GB)
- `--output` (optionnel): Nom du fichier de sortie (défaut: `generated_mri.dcm`)
- `--seed` (optionnel): Seed pour reproductibilité

### Exemples

```bash
# Générer 120 images pour 4.5 GB
python generate_dicom_mri.py --num-images 120 --total-size 4.5GB

# Avec nom de fichier personnalisé et seed
python generate_dicom_mri.py --num-images 50 --total-size 1GB --output test.dcm --seed 42
```

## Performance

- Small files (< 100MB): < 5 seconds
- Medium files (100MB - 1GB): 5-20 seconds
- Large files (1GB - 5GB): 20-90 seconds

Performance depends on disk speed. The script generates pixel data progressively to avoid memory issues.

## Testing

Run unit tests:
```bash
pytest tests/test_generate_dicom_mri.py -v
```

Run integration tests:
```bash
pytest tests/test_integration.py -v
```

Run all tests:
```bash
pytest tests/ -v
```
