package converter

// ToGrams mengonversi satuan berat apapun ke gram
func ToGrams(amount float64, unit string) float64 {
	switch unit {
	case "mg":
		return amount / 1000
	case "kg":
		return amount * 1000
	default: // asumsikan gram
		return amount
	}
}
