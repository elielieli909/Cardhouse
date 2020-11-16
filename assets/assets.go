package assets

import (
	"exchange/assets/book"
	"math/rand"
	"time"
)

// The assets package will keep track of all assets, their respective order books,
// and be the main entryway in the exchange pipeline for interacting with assets.

// Asset holds all asset metadata (name, ticker, etc.)
type Asset struct {
	id     int
	name   string
	ticker string
	// TODO: Eventually add ALL stats here (market cap, daily Vol, Spotify plays, etc.)
}

// Books is a map keyed off assetID to its respective book
var Books map[int]*book.Book

// Assets is a map keyed off assetID to its metadata
var Assets map[int]*Asset

// IDs is a map keyed off ticker to its ID
var IDs map[string]int

// keep track of previous assigned id so each is unique
var prevID int

// orderQueue maintains a map keyed off assetID to its queue of orders to be fulfilled
// TODO: MAKE THREAD-SAFE, CONSIDER HEAP-BASED PRIORITY QUEUE

// Initialize initializes the empty id=>Book and id=>metadata maps
func Initialize() {
	Books = make(map[int]*book.Book)
	Assets = make(map[int]*Asset)
	IDs = make(map[string]int)
	prevID = 0
}

// CreateAsset adds a new asset with name and ticker to the data structures, and starts concurrently handling orders from queue
func CreateAsset(name string, ticker string) {
	prevID++
	newBook := book.NewBook(prevID)
	// populate book with random limits
	populate(newBook)
	Books[prevID] = newBook

	asset := new(Asset)
	asset.id = prevID
	asset.name = name
	asset.ticker = ticker
	Assets[prevID] = asset

	IDs[ticker] = prevID

	// Begin concurrently handling orders from queue as they're added by the server
	go newBook.MatchOrders()
}

// JUST FOR TESTING: populate book with random limit orders
func populate(b *book.Book) {
	// Let's add a bunch of random orders to the book
	rand.Seed(time.Now().Unix())
	// Buys
	// for i := 0; i < 50; i++ {
	// 	// Random Buys between $10 and $40, of sizes [50, 500]
	// 	b.NewOrder(true, rand.Intn(450)+50, rand.Intn(30)+10)
	// }
	// // Sell
	// for i := 0; i < 50; i++ {
	// 	// Random Buys between $41 and $70, of sizes [50, 500]
	// 	b.NewOrder(false, rand.Intn(450)+50, rand.Intn(30)+41)
	// }
	for i := 0; i < 1000; i++ {
		lim := int(rand.NormFloat64()*10 + 40)
		print("%d\n", lim)
		buyOrSell := true
		if lim > 40 {
			buyOrSell = false
		}
		b.NewOrder(buyOrSell, rand.Intn(450)+50, lim)
	}

	// Let's cancel half of them randomly
	// for i := 0; i < 50; i++ {
	// 	randIndex := rand.Intn(len(b.OrderMap))
	// 	b.Cancel(randIndex)
	// }
}

// GetBookByID is an accessor function to get the pointer to the book for asset with id
func GetBookByID(id int) *book.Book {
	if b, exists := Books[id]; exists {
		return b
	}
	return nil
}

// GetBookBySymbol is an accessor function to get the pointer to the book for asset with symbol symbol
func GetBookBySymbol(symbol string) *book.Book {
	if id, exists := IDs[symbol]; exists {
		if b, exists := Books[id]; exists {
			return b
		}
	}
	return nil
}
