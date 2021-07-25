# Fake Coins service

## How to run
To run as docker-compose bundle `docker-compose up`. Supply needed environment variable values in `docker-compose.yml` if needed

## Implemented features
- Signup: creates 2 wallets for fake coins fBTC and fETH, issues one transaction for each with 100 amount
- JWT token retrieval and authorization
- List wallets

## Approach
Since it is required to implement only 5 endpoints, I have decided to go with only 2 main layers of the software:
1. API layer: handles REST API requests and takes on the responsibility for application logic
2. Active Record: by definition each active record represents an individual row in the DB.
   In addition, it combines data access logic with business logic. All business invariants are protected so if client code holds an instance of an active record, that means the data in that particular instance is logically consistent.

## Testing
I took a boilerplate code I use in my job to write quick tests which look like end to end and API tests combined.

To run the tests:
1. Launch Postgres somewhere
2. Apply migrations
3. Run `TEST_DB_URL=*your postgres connection url* go test ./test/...`

## Transactions considerations
There is no balance state for a wallet. Wallet is completely stateless and its balance is calculated from transactions.

This may lead to write conflicts during concurrent transactions for a single wallet.
It could be solved by using pessimistic lock.
The code could execute the query with a lock: first it locks the table in `ROW EXCLUSIVE` mode and then performs `SELECT ... FOR UPDATE`.
Then the code would check if there is enough balance, then insert new transaction and then commit the DB transaction
