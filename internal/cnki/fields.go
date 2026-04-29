package cnki

// fieldCodeMap translates our external flag values to CNKI's internal search
// field codes used inside QueryJson.
var fieldCodeMap = map[string]string{
	"topic":    "SU",
	"keyword":  "KY",
	"title":    "TI",
	"author":   "AU",
	"abstract": "AB",
	"fulltext": "FT",
	"doi":      "DOI",
}
