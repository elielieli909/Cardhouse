package main

import (
	"exchange/api"
	"exchange/assets"
	"exchange/users"
)

func main() {
	// Initialize empty maps for books and assets
	assets.Initialize()
	assets.CreateAsset("Travis Scott", "TRAV")
	assets.CreateAsset("24kGolden", "24k")
	assets.CreateAsset("Parallel Doug", "DOUG")
	assets.CreateAsset("Parallel Art", "ART")

	travBook := assets.GetBookByID(1)
	travBook.InOrderTraversal()

	// Initialize an empty ledger
	users.Initialize()

	// Begin concurrently handling orders from queue as they're added by the server
	// go assets.MatchOrders()

	// Start serving requests
	api.StartServer()

	// TODO: Add Orders to Queue, where they're assigned an id and hashed

	// // Example Market Buy for 10 shares
	// fmt.Printf("Bought 1000 shares for %d\n", b.ExecuteMarketBuy(1000))
	// fmt.Printf("New Market Price: %d\n", b.GetMarketPrice())

	// // Example Market Sell for 10 shares
	// fmt.Printf("Sold 1000 shares for %d\n", b.ExecuteMarketSell(1000))
	// fmt.Printf("New Market Price: %d\n", b.GetMarketPrice())

	// TODO: For each order in the queue, if MARKET or LIMIT with a match, execute order
	// If order is normal limit, add to book
}
