package vietlott

// Vietlott API endpoints configuration
const (
	// BaseURL is the base URL for Vietlott website
	BaseURL = "https://vietlott.vn"

	// Mega 6/45 endpoints
	Mega645ResultsPath    = "/vi/trung-thuong/ket-qua-trung-thuong/6-45"
	Mega645APIPath        = "/api/trung-thuong/6-45"
	Mega645DetailPath     = "/vi/trung-thuong/ket-qua-trung-thuong/6-45/detail/"

	// Power 6/55 endpoints
	Power655ResultsPath   = "/vi/trung-thuong/ket-qua-trung-thuong/6-55"
	Power655APIPath       = "/api/trung-thuong/6-55"
	Power655DetailPath    = "/vi/trung-thuong/ket-qua-trung-thuong/6-55/detail/"

	// Common API parameters
	DefaultPageNumber = 1
	DefaultPageSize   = 100
)

// GameTypeMapping maps our internal game types to API paths
var GameTypePathMap = map[string]string{
	"mega_6_45":  Mega645ResultsPath,
	"power_6_55": Power655ResultsPath,
}

// GameTypeAPIPathMap maps our internal game types to API paths
var GameTypeAPIPathMap = map[string]string{
	"mega_6_45":  Mega645APIPath,
	"power_6_55": Power655APIPath,
}
