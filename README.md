# DICOM MRI Generator

Outil CLI pour générer des séries DICOM d'IRM valides pour tester des interfaces médicales.

**Génère plusieurs fichiers DICOM** (un par image) dans un dossier, format standard attendu par les plateformes médicales.

## Installation

### Go (recommandé)

```bash
go build ./cmd/generate-dicom-mri/
```

### Python (version legacy)

La version Python originale est disponible dans le répertoire `python/` :

```bash
cd python
pip install -r requirements.txt
python generate_dicom_mri.py --num-images 10 --total-size 100MB
```

## Usage

```bash
./generate-dicom-mri --num-images 120 --total-size 1GB --output mri_series
```

Cela créera un dossier `mri_series/` contenant 120 fichiers DICOM individuels + fichier DICOMDIR:
```
mri_series/
├── DICOMDIR                    # Fichier d'index de la série
├── PT000000/ST000000/SE000000/ # Hiérarchie standard DICOM
│   ├── IM000001
│   ├── IM000002
│   └── ...
```

### Paramètres

| Paramètre | Description | Défaut |
|-----------|-------------|--------|
| `--num-images` | Nombre d'images/coupes (requis) | - |
| `--total-size` | Taille totale cible (requis) | - |
| `--output` | Dossier de sortie | `dicom_series` |
| `--seed` | Seed pour reproductibilité | auto |
| `--num-studies` | Nombre d'études | 1 |
| `--workers` | Workers parallèles | CPU cores |

### Exemples

```bash
# Générer 120 images pour 1 GB total
./generate-dicom-mri --num-images 120 --total-size 1GB

# Avec nom de dossier personnalisé et seed
./generate-dicom-mri --num-images 50 --total-size 500MB --output my_mri --seed 42

# Plusieurs études
./generate-dicom-mri --num-images 30 --total-size 500MB --num-studies 3

# Limiter le parallélisme
./generate-dicom-mri --num-images 100 --total-size 1GB --workers 4
```

## Caractéristiques

- Génère des fichiers DICOM individuels (format standard)
- **Fichier DICOMDIR** automatiquement créé (index de la série)
- **Hiérarchie PT/ST/SE** standard pour compatibilité PACS
- **Texte "File X/Y"** incrusté sur chaque image
- **Génération parallèle** avec worker pool (~4.5x speedup)
- Tous les fichiers partagent le même Study UID et Series UID
- Métadonnées MRI réalistes (SIEMENS, GE, PHILIPS)
- Noms de patients français réalistes
- Tags Window/Level pour affichage dans viewers médicaux
- Reproductible avec seed

## Performance

| Images | Taille | Temps (1 worker) | Temps (24 workers) |
|--------|--------|------------------|-------------------|
| 50 | 100MB | ~3.1s | ~0.7s |
| 120 | 1GB | ~15s | ~3s |

## Testing

```bash
# Tests unitaires
go test ./internal/...

# Tests d'intégration
go test ./tests/...

# Tous les tests
go test ./...
```

## Structure du projet

```
.
├── cmd/generate-dicom-mri/   # Point d'entrée CLI
├── internal/
│   ├── dicom/                # Génération DICOM et DICOMDIR
│   ├── image/                # Génération de pixels
│   └── util/                 # Utilitaires (UID, parsing)
├── tests/                    # Tests d'intégration
├── scripts/                  # Scripts de validation
├── python/                   # Version Python legacy
│   ├── generate_dicom_mri.py
│   ├── requirements.txt
│   └── tests/
└── go.mod
```
