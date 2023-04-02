package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"net/http"
	"os"
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
var apiUrl = "https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=0x9355372396e3F6daF13359B7b607a3374cc638e0&page=1&offset=2&sort=asc&apikey="
var transactions []Transaction = make([]Transaction, 0)
var txsFromAddr map[string][]int = make(map[string][]int)
var txsToAddr map[string][]int = make(map[string][]int)
var txsValue map[string][]int = make(map[string][]int)

func getEtherscanData() []Transaction {
	fmt.Println("Getting Etherscan data...")
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

		txsFromAddr[tx.From] = append(txsFromAddr[tx.From], index)
		txsToAddr[tx.To] = append(txsToAddr[tx.To], index)
		txsValue[tx.Value] = append(txsValue[tx.Value], index)
	}

	//fmt.Printf("%+v\n\n", transactions)
	//fmt.Printf("%+v\n\n", txsFromAddr)
	//fmt.Printf("%+v\n\n", txsToAddr)
	//fmt.Printf("%+v\n\n", txsValue)
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
		//from := ctx.Query("from")
		//to := ctx.Query("to")
		//limit := ctx.Query("limit")
		//offset := ctx.Query("offset")
		//aboveValue := ctx.Query("aboveValue")

		//value, ok := transactions[from]
		//if ok {
		//	ctx.JSON(http.StatusOK, gin.H{"from": from, "value": value})
		//} else {
		//	ctx.JSON(http.StatusOK, gin.H{"from": from, "status": "no value"})
		//}
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
