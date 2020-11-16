Description of the exchange


Theres gonna be three threads running at all times:

1. HTTP Listener, constantly listening for requests.
    (A description of all the requests will come later.)

    Overview: 
        Each time a request is detected, a goroutine running a handlerFunc will spawn to handle the request.

        A typical handler function should work like this:
            1. Check for errors (if the request could contain any)
            2. Do the operation (if necessary)
            3. Respond with update

        Since most API's will require an operation, like a lookup or order submission, channels will be used to pass the requests along to the allocation process, which will then determine which asset's order matching process should process the request.

Connected to:
Allocation Process via channels

2. Allocation Process

    Overview:
        One, or more, if neccessary (TODO: look into load balancing by dynamically allocating new goroutines) goroutines, listening on a channel (maybe buffered) for orders to come in.  When they do, check their destination and lookup the correct channel to send the order down.  In the case of order requests, that channel will lead to the requested asset's matching process.  In the case of data requests, instead of sending the request down a channel, the allocator will simply lookup the asset's book and grab the data (should be O(1))

Connected to:
Order Matching Processes, each having its own channel

3. Order Matching Process

    Overview: 
        One goroutine per asset.  Each of these goroutines are constantly listening on a specified buffered channel for order requests; once one comes in, lock the asset.  Then four things can happen:
            
            a. Market order request is fully matched
                The market order is matched with one or more previously established limit orders, in which case each transaction is noted by the ledger process
            b. Market order request is partially matched
                Right now, the order is filled as much as possible and then cancelled.  
                TODO: allow different order types so traders can customize what should happen here. Here's a reference: https://money.stackexchange.com/questions/87668/what-happens-if-a-market-order-is-not-fulfilled-completely
            c. Limit order request can be executed immediately
                The limit order is matched with an existing limit order on the other side of the book, as if the limit order is just a market order.
            d. Limit order request can't be executed immediately
                The limit order is added to this asset's book.
                TODO: Allow limit order timeframes, so as to give traders more choice over what happens to their limits

        For each matched order, the order details will be passed to a channel which leads to the ledger process, which will write the transaction in the ledger and facilitate the trade.

Connected to:
Ledger Process.  Each Matching Process will write to the ledger channel, which is listened to by the ledger process.

4. Ledger Process
    Overview: 
        The Ledger Process will be listening on one channel, which is written to by all assets' Matching process, and will add the transaction to the ledger.
        TODO: Move ledger to MySQL DB (to keep persistent)




Types of HTTP Requests:
    NOTES: 
        - /api/v1/assets/{assetID}/ contains all the apis for that asset (Getters, setters, etc.)
        - The * sign means that request requires a body
        - Responses are currently just sent when receieved, but shouldn't be (TODO: send after confirmation of operation)

    Accessors (for getting data, snapshots of the exchange):
    a. Get all Assets
        Asset Schema:
            { 
                assetID,
                Name,
                Ticker,
                TODO: Maybe add more stuff like # outstanding shares, market cap, daily volume, etc.
            }

        response: 200 OK, [Asset Schema1, Asset Schema2, ...]
            Some Error Code

        link: /api/v1/assets/all

    b. TODO: Add more direct ways of getting asset info (like by ID, ticker, etc.)

    c. LOB Snapshot

        TODO: Could add more details, like a list of the actual orders per limit

        LOB Schema:
            [
                {
                    Limit price,
                    Number of Orders,
                    Total Volume
                },
            ]

        response: 200 OK, LOB Schema
            Some Error Code

        link: /api/v1/assets/{assetID}/data/LOBSnapshot

    d. LOB History *

        TODO: Determine if this really should exist

        body: 
            timescale: 'daily', 'weekly', 'monthly'

        response: 200 OK, [LOB Schema1, LOB Schema2, ...]
            Some Error Code

        link: /api/v1/assets/{assetID}/data/LOBHistory

    e. Current Price

        response: 200 OK, integer price
            Some Error Code

        link: /api/v1/assets/{assetID}/data/marketPrice

    f. Price History *

        TODO: Make this, or determine if this should really exist

        body: 
            timescale: 'daily', 'weekly', 'monthly'
        
        response: 200 OK, [integer price1, integer price2, ...]
            Some Error Code

        link: /api/v1/assets/{assetID}/data/priceHistory
        
    g. BA Spread

        TODO: Make this, or determine what this actually should look like (Check other exchanges)

        response: 200 OK, {IDK}

        link: /api/v1/assets/{assetID}/data/BASpread

    Modifiers (for submitting orders):

    a. Send Order *

        body:
            qty: integer value, should be reasonable number of shares
            type: 'market', 'limit', TODO: Maybe add stops
            side: 'buy', 'sell'
            limit: integer value limit price, only looked at if type is 'limit'
            time_in_force: TODO: Figure this out

            api_key: TODO: Assign one of these to each user, and only allow requests from authorized keys

        TODO: Fix link, not actually this in the code

        response: 200 OK, integer OrderID
            Some Error Code (Timeout, Empty Book, Invalid Req Body, etc.)

            IMPORTANT NOTE ABOUT RESPONSE: OrderID should be noted, because it is used to cancel outstanding orders
                *** Thus, the orderID needs to be saved on the client ***

        link: /api/v1/assets/{assetID}/orders/makeOrder

    b. Cancel Order *

        TODO: Make this

        response 200 OK
            Some Error Code (Maybe an apology LOL)

        link: /api/v1/assets/{assetID}/orders/cancelOrder/{orderID}
