package vietlott

// Vietlott API endpoints configuration
const (
	// BaseURL is the base URL for Vietlott website
	BaseURL = "https://vietlott.vn"

	// Mega 6/45 endpoints
	Mega645ResultsPath = "/vi/trung-thuong/ket-qua-trung-thuong/645"
	Mega645DetailPath  = "/vi/trung-thuong/ket-qua-trung-thuong/645"

	// Power 6/55 endpoints - Use winning-number page for historical data
	Power655ResultsPath = "/vi/trung-thuong/ket-qua-trung-thuong/winning-number-655"
	Power655DetailPath  = "/vi/trung-thuong/ket-qua-trung-thuong/655"

	// Mega 6/45 historical results
	Mega645HistoryPath = "/vi/trung-thuong/ket-qua-trung-thuong/winning-number-645"

	// Common API parameters
	DefaultPageNumber = 1
	DefaultPageSize   = 100
)

// GameTypePathMap maps our internal game types to result page paths
var GameTypePathMap = map[string]string{
	"mega_6_45":  Mega645ResultsPath,
	"power_6_55": Power655ResultsPath,
}
