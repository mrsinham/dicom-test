// internal/util/institutions.go
package util

import (
	"fmt"
	"math/rand/v2"
)

// Institution holds generated institution data
type Institution struct {
	Name       string
	Address    string
	Department string
}

var (
	// Hospitals is the list of realistic hospital names
	Hospitals = []struct {
		Name    string
		Address string
	}{
		{"CHU Bordeaux", "Place Amelie Raba-Leon, 33000 Bordeaux"},
		{"Hopital Saint-Louis", "1 Avenue Claude Vellefaux, 75010 Paris"},
		{"CHU Toulouse", "2 Rue Viguerie, 31000 Toulouse"},
		{"Clinique du Parc", "155 Boulevard Stalingrad, 69006 Lyon"},
		{"Hopital Europeen Georges-Pompidou", "20 Rue Leblanc, 75015 Paris"},
		{"CHU Nantes", "1 Place Alexis-Ricordeau, 44000 Nantes"},
		{"CHU Lille", "2 Avenue Oscar Lambret, 59000 Lille"},
		{"Hopital de la Pitie-Salpetriere", "47-83 Boulevard de l'Hopital, 75013 Paris"},
		{"CHU Montpellier", "191 Avenue du Doyen Gaston Giraud, 34090 Montpellier"},
		{"Hopital Cochin", "27 Rue du Faubourg Saint-Jacques, 75014 Paris"},
		{"Massachusetts General Hospital", "55 Fruit Street, Boston, MA 02114"},
		{"Johns Hopkins Hospital", "1800 Orleans Street, Baltimore, MD 21287"},
		{"Cleveland Clinic", "9500 Euclid Avenue, Cleveland, OH 44195"},
		{"Mayo Clinic", "200 First Street SW, Rochester, MN 55905"},
		{"UCLA Medical Center", "757 Westwood Plaza, Los Angeles, CA 90095"},
	}

	// Departments is the list of medical departments
	Departments = []string{
		"Radiologie",
		"Imagerie Medicale",
		"Neuroradiologie",
		"Radiologie Interventionnelle",
		"Urgences",
		"Cardiologie",
		"Neurologie",
		"Oncologie",
		"Pediatrie",
		"Orthopedie",
	}
)

// GenerateInstitution generates a random institution with address and department.
// If rng is nil, uses shared default RNG.
func GenerateInstitution(rng *rand.Rand) Institution {
	if rng == nil {
		rng = defaultRNG
	}

	hospital := Hospitals[rng.IntN(len(Hospitals))]
	department := Departments[rng.IntN(len(Departments))]

	return Institution{
		Name:       hospital.Name,
		Address:    hospital.Address,
		Department: department,
	}
}

// GenerateStationName generates a station name based on modality and body part.
// Format: MODALITY_BODYPART_NN (e.g., "MR_HEAD_01", "CT_CHEST_03")
func GenerateStationName(modality, bodyPart string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	num := rng.IntN(10) + 1
	return modality + "_" + bodyPart + "_" + fmt.Sprintf("%02d", num)
}
