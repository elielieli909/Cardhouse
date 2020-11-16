package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type routes []route

// NewRouter returns a new router with routes defined in routes.go
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range apiRoutes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}

var apiRoutes = routes{
	// ACCESSORS
	// Route to get current market price of asset with assetID
	route{
		"Current Market Price",
		"GET",
		"/api/{assetID}/data/marketPrice",
		HandleMarketPriceRequest,
	},
	// Route to get snapshot of order book for asset with assetID
	route{
		"Order Book Snapshot",
		"GET",
		"/api/{assetID}/data/LOBSnapshot",
		HandleBookSnapshotRequest,
	},
	// Route to get snapshot of order book for asset with assetID
	route{
		"Transaction Ledger Snapshot (For Asset)",
		"GET",
		"/api/{assetID}/data/LedgerSnapshot",
		HandleAssetsLedgerSnapshotRequest,
	},
	// MODIFIERS
	// Route to post an order for asset with assetID
	// Order details specified in request.body (Market vs. Limit, numShares, etc.)
	route{
		"Order",
		"POST",
		"/api/order",
		HandleOrder,
	},
}
