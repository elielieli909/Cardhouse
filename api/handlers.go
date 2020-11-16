package api

import (
	"exchange/assets"
	"exchange/assets/book"
	"exchange/users"
	"fmt"
	"io"
	"io/ioutil"

	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// HandleMarketPriceRequest is the handler function for the API requesting current market price
func HandleMarketPriceRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	assetID, e := strconv.Atoi(vars["assetID"])

	if e != nil {
		panic(e)
	} else {
		b := assets.GetBookByID(assetID)
		json.NewEncoder(w).Encode(b.GetMarketPrice())
	}
}

type bookResponseSchema [][]book.Limit

func HandleBookSnapshotRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	assetID, e := strconv.Atoi(vars["assetID"])

	if e != nil {
		panic(e)
	} else {
		b := assets.GetBookByID(assetID)
		bids, asks := b.InOrderTraversal()

		res := bookResponseSchema{
			bids,
			asks,
		}

		json.NewEncoder(w).Encode(res)
	}
}

func HandleAssetsLedgerSnapshotRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	assetID, e := strconv.Atoi(vars["assetID"])

	if e != nil {
		panic(e)
	} else {
		ledger := users.GetLedger()
		hist := ledger.GetAssetHistory(assetID)

		// res := bookResponseSchema{
		// 	bids,
		// 	asks,
		// }

		json.NewEncoder(w).Encode(hist)
	}
}

// HandleOrder is the handler function for handling API order requests
// TODO: send a channel to the Enqueue Order to get response, send back only after that's happened
func HandleOrder(w http.ResponseWriter, r *http.Request) {
	// read body
	body, e := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if e != nil {
		panic(e)
	}

	// Close IO
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	// Fill order with details from req body
	var order book.OrderSchema
	if err := json.Unmarshal(body, &order); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		fmt.Print(err)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	fmt.Printf("Received %s %s Order for %d shares of %s!\n", order.OrderType, order.Side, order.Qty, order.Symbol)

	// TODO: Catch errors first, then execute order (FOR CLEANLINESS)
	// Make sure the symbol actually exists
	b := assets.GetBookBySymbol(order.Symbol)
	if b == nil {
		// ERROR: SYMBOL DOESN'T EXIST
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Symbol Doesn't Exist"); err != nil {
			panic(err)
		}
	}
	if order.Side != "buy" && order.Side != "sell" {
		// Make sure its either a Buy or a Sell
		// ERROR: INVALID ORDER SIDE
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Error: didn't specify order side!"); err != nil {
			panic(err)
		}
	}
	if order.OrderType != "market" && order.OrderType != "limit" {
		// Make sure its either a market or limit order
		// TODO: Add stop and shit
		// ERROR: NOT AN ACCEPTABLE ORDER TYPE
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Invalid order type!"); err != nil {
			panic(err)
		}
	}
	if order.Qty <= 0 {
		// ERROR: INVALID QUANTITY
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode("Quantity must be greater than 0"); err != nil {
			panic(err)
		}
	}
	// Check if limit book is empty for this market order
	if order.OrderType == "market" {
		if order.Side == "buy" {
			if b.GetBestOffer() == nil {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusNotAcceptable)
				if err := json.NewEncoder(w).Encode("You're order has been cancelled due to a lack of liquidity... Try placing a limit order."); err != nil {
					panic(err)
				}
				return
			}
		} else {
			if b.GetBestBid() == nil {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusNotAcceptable)
				if err := json.NewEncoder(w).Encode("You're order has been cancelled due to a lack of liquidity... Try placing a limit order."); err != nil {
					panic(err)
				}
				return
			}
		}
	}

	// Everything's good, add order to queue
	b.EnqueueOrder(&order)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode("Your Order Has Been Processed!"); err != nil {
		panic(err)
	}
}
