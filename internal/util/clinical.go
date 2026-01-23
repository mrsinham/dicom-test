// internal/util/clinical.go
package util

import "math/rand/v2"

// BodyPartsByModality maps modalities to appropriate body parts
var BodyPartsByModality = map[string][]string{
	"MR": {"HEAD", "BRAIN", "CSPINE", "TSPINE", "LSPINE", "KNEE", "SHOULDER", "HIP", "ANKLE", "WRIST", "PELVIS", "ABDOMEN", "CHEST"},
	"CT": {"HEAD", "CHEST", "ABDOMEN", "PELVIS", "CSPINE", "TSPINE", "LSPINE", "EXTREMITY"},
	"CR": {"CHEST", "HAND", "FOOT", "KNEE", "SHOULDER", "SKULL", "SPINE", "PELVIS", "RIBS"},
	"DX": {"CHEST", "HAND", "FOOT", "KNEE", "SHOULDER", "SKULL", "SPINE", "PELVIS", "RIBS"},
	"US": {"ABDOMEN", "PELVIS", "BREAST", "THYROID", "HEART", "LIVER", "KIDNEY", "UTERUS"},
	"MG": {"BREAST"},
}

// DefaultBodyParts is used when modality is unknown
var DefaultBodyParts = []string{"HEAD", "CHEST", "ABDOMEN", "EXTREMITY"}

// ProtocolsByModalityAndBodyPart maps modality+bodypart to protocol names
var ProtocolsByModalityAndBodyPart = map[string]map[string][]string{
	"MR": {
		"HEAD":     {"BRAIN_ROUTINE", "BRAIN_WITH_CONTRAST", "BRAIN_STROKE", "BRAIN_TUMOR", "BRAIN_MS"},
		"BRAIN":    {"BRAIN_ROUTINE", "BRAIN_WITH_CONTRAST", "BRAIN_STROKE", "BRAIN_TUMOR", "BRAIN_MS"},
		"CSPINE":   {"CSPINE_ROUTINE", "CSPINE_WITH_CONTRAST"},
		"TSPINE":   {"TSPINE_ROUTINE", "TSPINE_WITH_CONTRAST"},
		"LSPINE":   {"LSPINE_ROUTINE", "LSPINE_WITH_CONTRAST", "LSPINE_DISC"},
		"KNEE":     {"KNEE_ROUTINE", "KNEE_ACL", "KNEE_MENISCUS"},
		"SHOULDER": {"SHOULDER_ROUTINE", "SHOULDER_ARTHROGRAM"},
		"HIP":      {"HIP_ROUTINE", "HIP_ARTHROGRAM"},
		"ABDOMEN":  {"ABDOMEN_ROUTINE", "MRCP", "LIVER_DYNAMIC"},
		"PELVIS":   {"PELVIS_ROUTINE", "PROSTATE_MP"},
	},
	"CT": {
		"HEAD":    {"HEAD_ROUTINE", "HEAD_TRAUMA", "HEAD_STROKE", "HEAD_SINUS"},
		"CHEST":   {"CHEST_ROUTINE", "CHEST_PE", "CHEST_HRCT", "CHEST_TRAUMA"},
		"ABDOMEN": {"ABDOMEN_ROUTINE", "ABDOMEN_TRIPLE_PHASE", "ABDOMEN_TRAUMA"},
		"PELVIS":  {"PELVIS_ROUTINE", "PELVIS_WITH_CONTRAST"},
		"CSPINE":  {"CSPINE_TRAUMA", "CSPINE_ROUTINE"},
	},
}

// ClinicalIndications maps body parts to common clinical indications
var ClinicalIndications = map[string][]string{
	"HEAD":     {"Cephalees persistantes", "Vertiges", "Trouble de la vision", "Suspicion AVC", "Bilan tumoral"},
	"BRAIN":    {"Cephalees persistantes", "Vertiges", "Trouble de la vision", "Suspicion AVC", "Bilan tumoral"},
	"CHEST":    {"Toux chronique", "Dyspnee", "Douleur thoracique", "Bilan infectieux", "Suspicion EP"},
	"ABDOMEN":  {"Douleur abdominale", "Bilan hepatique", "Masse abdominale", "Occlusion"},
	"KNEE":     {"Douleur genou", "Traumatisme", "Suspicion rupture LCA", "Blocage articulaire"},
	"SHOULDER": {"Douleur epaule", "Limitation mobilite", "Traumatisme"},
	"LSPINE":   {"Lombalgie", "Sciatique", "Bilan hernie discale"},
	"PELVIS":   {"Douleur pelvienne", "Bilan oncologique", "Trouble urinaire"},
	"BREAST":   {"Depistage", "Masse palpable", "Bilan extension"},
}

// DefaultIndications is used when body part has no specific indications
var DefaultIndications = []string{"Bilan diagnostique", "Controle", "Suivi"}

// GetBodyPartsForModality returns the list of body parts for a modality
func GetBodyPartsForModality(modality string) []string {
	if parts, ok := BodyPartsByModality[modality]; ok {
		return parts
	}
	return DefaultBodyParts
}

// GenerateBodyPart returns a random body part appropriate for the modality
func GenerateBodyPart(modality string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}
	parts := GetBodyPartsForModality(modality)
	return parts[rng.IntN(len(parts))]
}

// GenerateProtocolName generates a protocol name for the given modality and body part
func GenerateProtocolName(modality, bodyPart string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	if modalityProtocols, ok := ProtocolsByModalityAndBodyPart[modality]; ok {
		if protocols, ok := modalityProtocols[bodyPart]; ok {
			return protocols[rng.IntN(len(protocols))]
		}
	}

	// Default protocol name
	return modality + "_" + bodyPart + "_ROUTINE"
}

// GenerateClinicalIndication generates a clinical indication for the body part
func GenerateClinicalIndication(modality, bodyPart string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	if indications, ok := ClinicalIndications[bodyPart]; ok {
		return indications[rng.IntN(len(indications))]
	}
	return DefaultIndications[rng.IntN(len(DefaultIndications))]
}
