package users

import (
	"math/rand"
	"time"
)

var curTID int = 0

// Transaction describes the transaction of an asset, which is then stored in the ledger
type Transaction struct {
	id        int
	seller    *User
	buyer     *User
	assetID   int
	NumShares int   `json:"numShares"`
	Price     int   `json:"price"`
	Date      int64 `json:"time"`
}

// Ledger holds all transactions ever made.
type Ledger struct {
	historyAll       []*Transaction
	HistoryByAssetID map[int][]*Transaction
	historyByUserID  map[int][]*Transaction

	// pointer to all Users held here so as to access it
	users *Users
}

// GlobalLedger is the ledger keeping track of all transactions.  Constructed in users.Initialize()
var globalLedger *Ledger

// Initialize is GlobalLedger's constructor; returns the blank ledger.  MUST BE CALLED IN ORDER FOR THE LEDGER TO BEGIN
func Initialize() {
	l := new(Ledger)
	l.historyAll = make([]*Transaction, 0)
	l.HistoryByAssetID = make(map[int][]*Transaction)
	l.historyByUserID = make(map[int][]*Transaction)
	l.users = NewUsers()
	l.populate()

	globalLedger = l
	// TODO: temporary
	//globalLedger.populate()
}

// populates ledger with some random users.
func (l *Ledger) populate() {
	for i := 0; i < 20; i++ {
		u := l.users.NewUser()
		u.DepositCash(rand.Intn(100000))
		u.name = "username"
	}
}

// GetLedger is O(1) access to get the global ledger from another package
func GetLedger() *Ledger {
	return globalLedger
}

// GetAssetHistory exposes the transaction history for the asset with assetID
func (l *Ledger) GetAssetHistory(assetID int) []*Transaction {
	return l.HistoryByAssetID[assetID]
}

// RecordTrade performs the trade operation, recording the transaction and shifting funds and ownership accordingly
func (l *Ledger) RecordTrade(assetID int, numShares int, price int, buyerID int, sellerID int) bool {
	buyer := l.users.users[buyerID]
	seller := l.users.users[sellerID]

	// Check for errors
	// TODO: Implement error checking. Right now just save the transaction
	// if buyer.cash < numShares*price {
	// 	return false
	// }
	// if seller.sharesOwned[assetID] < numShares {
	// 	return false
	// }

	// Create the transaction
	t := new(Transaction)
	t.assetID = assetID
	t.Date = time.Now().UnixNano()
	t.seller = seller
	t.buyer = buyer
	t.NumShares = numShares
	t.Price = price

	// Exchange cash and numShares between users
	seller.cash += numShares * price
	seller.sharesOwned[assetID] -= numShares
	if seller.sharesOwned[assetID] == 0 {
		// Remove assets from User's list of assets
		for index, id := range seller.assets {
			if id == assetID {
				seller.assets[len(seller.assets)-1], seller.assets[index] = seller.assets[index], seller.assets[len(seller.assets)-1]
				seller.assets = seller.assets[:len(seller.assets)-1]
			}
		}
	}

	buyer.cash -= numShares * price
	buyer.sharesOwned[assetID] += numShares
	// TODO: check if buyer already has the asset, then don't add it
	buyer.assets = append(buyer.assets, assetID)

	// Record transaction in Ledger
	l.historyAll = append(l.historyAll, t)
	l.HistoryByAssetID[assetID] = append(l.HistoryByAssetID[assetID], t)
	l.historyByUserID[buyerID] = append(l.historyByUserID[buyerID], t)
	l.historyByUserID[sellerID] = append(l.historyByUserID[sellerID], t)

	return true
}
