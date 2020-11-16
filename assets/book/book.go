package book

// The Order Book will maintain a set of limit offers for both the Buy and Sell side.
// The data structures look like this:
// 			    Book
// 		BuyTree       SellTree
// 	Limit1, Limit2  Limit3, Limit4
//
// limitMap = {price1: Limit1, price2: Limit2, price3: Limit3 ... etc.}
// orderMap = {id1: Order1, id2: Order2, ... etc. }
//
// Each Limit maintains a linked list of Orders.
// If a limit order comes in for a Sell but the limit is in the BuyTree:
// 	The order is executed until either all Buy orders are fulfilled
// 	or all the Sell ones are, in which case the Limit either stays or
// 	switches trees.
//
// Book operates under the assumption that orders are added to orderMap before added to Book

// TODO: MOVE MUTEX LOCK INTO BOOK STRUCT AND MOVE ORDER QUEUE OUT SO LIMITS AND MARKETS CAN BE
// HANDLED CONCURRENTLY
import (
	"fmt"
	"log"
	"time"

	"github.com/HuKeping/rbtree"

	"exchange/users"
)

var curID int

// Order is the basic order, added to linked list of Limit
type Order struct {
	idNumber    int
	buyOrSell   bool // true: Buy, false: Sell
	shares      int
	limit       int
	entryTime   int64 // Time received by API
	eventTime   int64 // Time matched
	parentLimit *Limit
}

// Limit holds a doubly linked list of Orders at specified limit price
type Limit struct {
	LimitPrice  int      `json:"price"`
	Size        int      `json:"size"`   // I'm thinking this is length of linked list
	TotalVolume int      `json:"volume"` // This is sum of shares of linked list
	orders      []*Order // Switched from Linked List to Slices
}

// Order RB-Tree by limitPrice
func (x *Limit) Less(than rbtree.Item) bool {
	return x.LimitPrice < than.(*Limit).LimitPrice
}

// Book holds two BST's of Limit prices, each holding linked lists of orders
type Book struct {
	assetID int

	BuyTree    *rbtree.Rbtree
	sellTree   *rbtree.Rbtree
	lowestSell *Limit
	highestBuy *Limit

	// TODO: Maybe flush OrderMap to database at end of trade day
	OrderMap map[int]*Order // Map keyed off orderID -> Order
	limitMap map[int]*Limit // Map keyed off limitPrice -> Limit

	// TODO: Maybe move this somewhere
	marketPrice int // set whenever a market order is satisfied

	//orderQueue []*OrderSchema // Queue of orders for this asset's book
	OrderQueue chan *OrderSchema // Queue of orders but with a channel

	//mu sync.Mutex // Mutex lock, one per book
}

// GetMarketPrice returns the current market price of this asset
func (b *Book) GetMarketPrice() int {
	return b.marketPrice
}

// NewBook is a "constructor" for the Book.  Necessary to initialize the order and limit maps
func NewBook(assetID int) *Book {
	b := new(Book)
	b.BuyTree = rbtree.New()
	b.sellTree = rbtree.New()
	b.OrderMap = make(map[int]*Order)
	b.limitMap = make(map[int]*Limit)
	b.marketPrice = 0
	b.assetID = assetID
	//b.orderQueue = make([]*OrderSchema, 0)
	// Make a buffered queue for orders, right now with length 30
	b.OrderQueue = make(chan *OrderSchema, 30)
	curID = 1
	return b
}

// NewOrder generates a reference to a new Order object and adds it to the book
func (b *Book) NewOrder(buyOrSell bool, shares int, limit int) *Order {
	//	b.mu.Lock()
	o := new(Order)
	curID++
	o.idNumber = curID
	o.buyOrSell = buyOrSell
	o.shares = shares
	o.limit = limit
	o.entryTime = time.Now().Unix()
	b.OrderMap[curID] = o

	// b.mu.Unlock()

	b.Add(curID)
	return o
}

func newLimit(limit int) *Limit {
	l := new(Limit)
	l.LimitPrice = limit
	l.orders = make([]*Order, 0)
	return l
}

// Add order with orderID to book.  Only called if order can't be executed already (needs to be saved)
// O(log(M)) for first order at a limit, O(1) for all else
func (b *Book) Add(orderID int) {
	// b.mu.Lock()
	// defer b.mu.Unlock()
	// Get order
	o := b.OrderMap[orderID]

	// Check if can be O(1)
	if l, exists := b.limitMap[o.limit]; exists {
		// Limit already exists, add to end of linked list of orders
		o.parentLimit = l
		l.orders = append(l.orders, o)

		// Update limit metadata
		l.Size = l.Size + 1
		l.TotalVolume = l.TotalVolume + o.shares
		return
	}

	// Limit doesn't exist yet, insert new limit in tree O(log(M))
	l := newLimit(o.limit)
	l.Size = 1
	l.TotalVolume = o.shares
	l.orders = append(l.orders, o)
	o.parentLimit = l

	// Is the order a buy or sell?
	if o.buyOrSell {
		// Buy
		// Set highest bid if necessary
		if b.GetBestBid() == nil || l.LimitPrice > b.GetBestBid().LimitPrice {
			b.highestBuy = l
		}
		// Insert limit into RB-Tree
		b.BuyTree.Insert(l)
	} else {
		// Sell
		// Set lowest ask if necessary
		if b.GetBestOffer() == nil || l.LimitPrice < b.GetBestOffer().LimitPrice {
			b.lowestSell = l
		}
		// Insert limit into RB-Tree
		b.sellTree.Insert(l)
	}

	// Add limit to map
	b.limitMap[l.LimitPrice] = l

	// Add order to map (COMMENTED BECAUSE SHOULD BE DONE OUTSIDE ADD)
	// b.orderMap[orderID] = o
}

// Cancel order with orderID
func (b *Book) Cancel(orderID int) {
	// b.mu.Lock()
	// defer b.mu.Unlock()
	// Remove order from all structures, limit too if necessary
	if o, exists := b.OrderMap[orderID]; exists {
		// Delete order from linked list
		l := o.parentLimit
		for i := 0; i < l.Size; i++ {
			if l.orders[i] == o {
				l.orders = append(l.orders[:i], l.orders[i+1:]...)
				// Break so as not to have bad access (ex delete first item in list of 2 then access l[1])
				break
			}
		}

		// Check if parent Limit is now empty
		if l.Size == 1 {
			// This was the last order for this limit.  Delete the limit
			if o.buyOrSell {
				// Limit in buyTree
				b.BuyTree.Delete(l)

				// Update highestBuy
				if b.BuyTree.Min() != nil {
					b.highestBuy = b.BuyTree.Max().(*Limit)
				} else {
					b.highestBuy = nil
				}
			} else {
				// Limit in sellTree
				b.sellTree.Delete(l)

				// Update lowestSell
				if b.sellTree.Min() != nil {
					b.lowestSell = b.sellTree.Min().(*Limit)
				} else {
					b.lowestSell = nil
				}
			}
			// Delete from limit map
			delete(b.limitMap, l.LimitPrice)
		} else {
			// Update parent Limit metadata
			l.Size = l.Size - 1
			l.TotalVolume = l.TotalVolume - o.shares
		}

		// Delete the order from orderMap TODO: UNDERSTAND IF THIS IS NECESSARY
		delete(b.OrderMap, orderID)

		// Hopefully since Go garbage collects, this order is now gonzo
	} else {
		// invalid orderID, order not in map
		fmt.Printf("Invalid Order ID: %d\n", orderID)
	}

}

// ExecuteMarketBuy is called when a market buy comes in for numShares
// Returns total cost of transaction
func (b *Book) ExecuteMarketBuy(numShares int) int {
	// b.mu.Lock()
	// defer b.mu.Unlock()
	transactionSum := 0
	for numShares > 0 {
		// Get best offer
		bestLim := b.GetBestOffer()
		// No More bids in the book?
		if bestLim == nil {
			return transactionSum
		}

		oldestOrder := bestLim.orders[0]

		if oldestOrder.shares > numShares {
			// Just the oldest order is enough to fulfill market buy
			b.marketPrice = bestLim.LimitPrice
			oldestOrder.shares -= numShares
			bestLim.TotalVolume -= numShares

			// Record in ledger
			ledger := users.GetLedger()
			ledger.RecordTrade(b.assetID, numShares, bestLim.LimitPrice, 2, 1)

			return transactionSum + (numShares * bestLim.LimitPrice)
		} else {
			// This order will totally fill the oldest
			// TODO: keep track of what orders were filled to fulfill later
			b.marketPrice = bestLim.LimitPrice
			numShares = numShares - oldestOrder.shares
			transactionSum += oldestOrder.shares * oldestOrder.limit
			b.Cancel(oldestOrder.idNumber)

		}
	}
	return transactionSum
}

// ExecuteMarketSell is called when a market sell comes in for numShares
// Returns total cost of transaction
func (b *Book) ExecuteMarketSell(numShares int) int {
	// b.mu.Lock()
	// defer b.mu.Unlock()
	// Get best offer
	transactionSum := 0
	for numShares > 0 {
		bestLim := b.GetBestBid()
		// No More bids in the book?
		if bestLim == nil {
			return transactionSum
		}

		oldestOrder := bestLim.orders[0]

		if oldestOrder.shares > numShares {
			// Record in ledger
			ledger := users.GetLedger()
			// Only execute the trade if its OK with the ledger; i.e buyer doesn't have enough funds, seller doesn't have enough of the asset
			if ledger.RecordTrade(b.assetID, numShares, bestLim.LimitPrice, 3, 1) {
				// The current limit order is enough to fulfill market buy; Fill, Record in Ledger, alert participants.
				b.marketPrice = bestLim.LimitPrice
				oldestOrder.shares -= numShares
				bestLim.TotalVolume -= numShares
			}

			return transactionSum + (numShares * bestLim.LimitPrice)
		} else {
			// This order will totally fill the oldest, and start filling the next order.
			// Fill this order, record it in the ledger, alert the owner of the order
			b.marketPrice = bestLim.LimitPrice
			numShares = numShares - oldestOrder.shares
			transactionSum += oldestOrder.shares * oldestOrder.limit
			b.Cancel(oldestOrder.idNumber)
		}
	}

	return transactionSum
}

// GetVolumeAtLimit returns the total volume of orders at that limit price
func (b *Book) GetVolumeAtLimit(limit int) int {
	// Get volume at limit price if it exists
	if l, exists := b.limitMap[limit]; exists {
		return l.TotalVolume
	}
	return 0
}

// GetBestBid returns price of highest buy limit
func (b *Book) GetBestBid() *Limit {
	if b.highestBuy != nil {
		return b.highestBuy
	}
	return nil
}

// GetBestOffer returns price of lowest sell limit
func (b *Book) GetBestOffer() *Limit {
	if b.lowestSell != nil {
		return b.lowestSell
	}

	return nil
}

// EnqueueOrder pushes an order to the queue to be executed later
func (b *Book) EnqueueOrder(order *OrderSchema) {
	//mu.Lock()
	//b.orderQueue = append(b.orderQueue, order)
	b.OrderQueue <- order
	//mu.Unlock()
}

// MatchOrders will be running constantly as a goroutine alongside the http listener.  This pops orders from the queue one by one, matching them appropriately.
func (b *Book) MatchOrders() {
	for {
		//mu.Lock()
		o := <-b.OrderQueue
		//if len(b.orderQueue) > 0 {
		start := time.Now()
		// Pop order
		//	o := b.orderQueue[0]
		//	b.orderQueue = b.orderQueue[1:]

		if o.OrderType == "market" {
			// Market Order, simply match
			if o.Side == "buy" {
				fmt.Printf("Bought %d worth of %s!", b.ExecuteMarketBuy(o.Qty), o.Symbol)
			} else {
				fmt.Printf("Sold %d worth of %s!", b.ExecuteMarketSell(o.Qty), o.Symbol)
				//b.ExecuteMarketSell(o.Qty)
			}
		} else if o.OrderType == "limit" {
			// Limit Order, add to book
			// TODO: Consider cases which consitute instant matching (limit buy too high)
			if o.Side == "buy" {
				if b.GetBestOffer() == nil || o.LimitPrice < b.GetBestOffer().LimitPrice {
					b.NewOrder(true, o.Qty, o.LimitPrice)
				} else {
					// Limit Price higher than lowest limit sell; execute immediately
					fmt.Printf("Bought %d worth of %s!", b.ExecuteMarketBuy(o.Qty), o.Symbol)
				}
			} else {
				if b.GetBestBid() == nil || o.LimitPrice > b.GetBestBid().LimitPrice {
					b.NewOrder(false, o.Qty, o.LimitPrice)
				} else {
					// Limit Price lower than highest limit buy; execute immediately
					fmt.Printf("Sold %d worth of %s!", b.ExecuteMarketSell(o.Qty), o.Symbol)
				}
			}
			//b.InOrderTraversal()
		}
		elapsed := time.Since(start)
		log.Printf("Order operation took %s", elapsed)
		//	}
		//mu.Unlock()
	}
}

var bids []Limit
var asks []Limit

// InOrderTraversal prints the contents of each limit tree, in order, and returns slices of those limits
func (b *Book) InOrderTraversal() ([]Limit, []Limit) {
	bids = make([]Limit, 0)
	asks = make([]Limit, 0)

	//fmt.Printf("Bid Limits: \n")
	// traverse(b.BuyTree)
	b.BuyTree.Ascend(b.BuyTree.Min(), printBuys)
	//fmt.Printf("Ask Limits: \n")
	// traverse(b.sellTree)
	b.sellTree.Ascend(b.sellTree.Min(), printSells)

	return bids, asks
}

func printBuys(item rbtree.Item) bool {
	i, ok := item.(*Limit)
	if !ok {
		return false
	}
	bids = append(bids, *i)
	//fmt.Printf("Price: %d    Orders: %d    Volume: %d\n", i.LimitPrice, i.Size, i.TotalVolume)
	return true
}

func printSells(item rbtree.Item) bool {
	i, ok := item.(*Limit)
	if !ok {
		return false
	}
	asks = append(asks, *i)
	//fmt.Printf("Price: %d    Orders: %d    Volume: %d\n", i.LimitPrice, i.Size, i.TotalVolume)
	return true
}
