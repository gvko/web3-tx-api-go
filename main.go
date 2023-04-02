package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Transaction struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
}

type ApiResponse struct {
	Status       string        `json:"status"`
	Message      string        `json:"message"`
	Transactions []Transaction `json:"result"`
}

var etherscanApiKey = ""
var apiUrl = "https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=0x9355372396e3F6daF13359B7b607a3374cc638e0&page=1&offset=10&sort=asc&apikey="
var transactions []Transaction = make([]Transaction, 0)
var txsByFromAddr map[string][]int = make(map[string][]int)
var txsByToAddr map[string][]int = make(map[string][]int)
var txsByValue map[string][]int = make(map[string][]int)

func getEtherscanData() []Transaction {
	log.Println("Get and populate Etherscan data...")
	resp, err := http.Get(apiUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var response ApiResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		panic(err)
	}

	return response.Transactions
}

func populateDbs(txs []Transaction) {
	// reset the DB, so it only holds 100 records
	transactions = transactions[:0]

	for index, tx := range txs {
		transactions = append(transactions, tx)

		txsByFromAddr[tx.From] = append(txsByFromAddr[tx.From], index)
		txsByToAddr[tx.To] = append(txsByToAddr[tx.To], index)
		txsByValue[tx.Value] = append(txsByValue[tx.Value], index)
	}

	//fmt.Printf("%+v\n\n", transactions)
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	router.GET("/etherscan-data", func(ctx *gin.Context) {
		txs := getEtherscanData()
		populateDbs(txs)
	})

	router.GET("/transactions", func(ctx *gin.Context) {
		from := ctx.Query("from")
		to := ctx.Query("to")
		//limit := ctx.Query("limit")
		//offset := ctx.Query("offset")
		aboveValueStr := ctx.Query("aboveValue")
		// TODO: input validation for the query params

		var txs []Transaction = []Transaction{}

		if from != "" {
			for _, txIndex := range txsByFromAddr[from] {
				txs = append(txs, transactions[txIndex])
			}
			ctx.JSON(http.StatusOK, txs)
			return
		}
		if to != "" {
			for _, txIndex := range txsByToAddr[to] {
				txs = append(txs, transactions[txIndex])
			}
			ctx.JSON(http.StatusOK, txs)
			return
		}
		if aboveValueStr != "" {
			aboveValue, err := strconv.ParseInt(aboveValueStr, 10, 64)
			if err != nil {
				// TODO: handle error
				panic("'aboveValue' must be a valid integer")
			}

			for key, txIndexes := range txsByValue {
				keyAsInt, err := strconv.ParseInt(key, 10, 64)
				if err != nil {
					// TODO: handle error
					fmt.Println("Key is not a valid integer", key)
					continue
				}

				if keyAsInt > aboveValue {
					for _, txIndex := range txIndexes {
						txs = append(txs, transactions[txIndex])
					}
				}
			}

			ctx.JSON(http.StatusOK, txs)
			return
		}

		// no query params were provided, so return ALL txs
		ctx.JSON(http.StatusOK, transactions)
	})

	return router
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Could not load env vars from .env file")
	}
	etherscanApiKey = os.Getenv("ETHERSCAN_API_KEY")
	apiUrl += etherscanApiKey

	r := setupRouter()
	r.Run(":8080")
}
