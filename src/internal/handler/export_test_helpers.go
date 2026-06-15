package handler

// ExportNormalizeGender exposes normalizeGender for external handler tests.
func ExportNormalizeGender(g string) string {
	return normalizeGender(g)
}
