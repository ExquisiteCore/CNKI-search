package cnki

// CNKI URL constants. The detail URL of each paper is captured from the
// search results table — it contains session-bound query params and must not
// be reconstructed by hand.
const (
	URLHome          = "https://www.cnki.net"
	URLKNSBase       = "https://kns.cnki.net"
	URLAdvSearch     = "https://kns.cnki.net/kns8s/AdvSearch"
	URLQuickSearch   = "https://kns.cnki.net/kns8s/search"
	URLClientID      = "https://recsys.cnki.net/RCDService/api/UtilityOpenApi/GenerateClientID"
	URLBriefGridPath = "/kns8s/brief/grid"
)
