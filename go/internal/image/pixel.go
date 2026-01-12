package image

import "math/rand/v2"

// GenerateSingleImage generates random pixel data for a single MRI image.
//
// Returns a slice of uint16 values in 12-bit range (0-4095) typical for MRI.
// The seed parameter ensures reproducible generation.
func GenerateSingleImage(width, height int, seed int64) []uint16 {
	// Seed the random number generator for reproducibility
	rng := rand.New(rand.NewPCG(uint64(seed), uint64(seed)))

	// Generate random pixels in 12-bit range (0-4095)
	size := width * height
	pixels := make([]uint16, size)

	for i := 0; i < size; i++ {
		pixels[i] = uint16(rng.IntN(4096))
	}

	return pixels
}
