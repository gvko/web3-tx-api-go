package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Transaction struct {
	BlockNumber string `json:"blockNumber"`
	TimeStamp   string `json:"timeStamp"`
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
}

//ContractAddress   string `json:"contractAddress"`
//Nonce             string `json:"nonce"`
//BlockHash         string `json:"blockHash"`
//TokenName         string `json:"tokenName"`
//TokenSymbol       string `json:"tokenSymbol"`
//TokenDecimal      string `json:"tokenDecimal"`
//TransactionIndex  string `json:"transactionIndex"`
//Gas               string `json:"gas"`
//GasPrice          string `json:"gasPrice"`
//GasUsed           string `json:"gasUsed"`
//CumulativeGasUsed string `json:"cumulativeGasUsed"`
//Input             string `json:"input"`
//Confirmations     string `json:"confirmations"`
//}

type ApiResponse struct {
	Status       string        `json:"status"`
	Message      string        `json:"message"`
	Transactions []Transaction `json:"result"`
}

var etherscanApiKey = ""
var apiUrl = "https://api.etherscan.io/api?module=account&action=tokentx&contractaddress=0x9355372396e3F6daF13359B7b607a3374cc638e0&sort=asc"
var transactions = make([]Transaction, 0)
var txsByFromAddr = make(map[string][]int)
var txsByToAddr = make(map[string][]int)
var txsByValue = make(map[string][]int)

func getEtherscanData(apiUrl string) []Transaction {
	log.Println("Get and populate Etherscan data...")
	res, err := http.Get(apiUrl)
	if err != nil {
		// TODO: recover from error
		panic(err)
	}
	defer res.Body.Close()

	var response ApiResponse

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		// TODO: recover from error
		panic(err)
	}

	return response.Transactions
}

func populateDbs(txs []Transaction) {
	// reset the DBs, so the main one only holds 100 records
	// and the index DBs are re-populated, to ensure correct data
	transactions = transactions[:0]
	txsByFromAddr = make(map[string][]int)
	txsByToAddr = make(map[string][]int)
	txsByValue = make(map[string][]int)

	for index, tx := range txs {
		transactions = append(transactions, tx)

		txsByFromAddr[tx.From] = append(txsByFromAddr[tx.From], index)
		txsByToAddr[tx.To] = append(txsByToAddr[tx.To], index)
		txsByValue[tx.Value] = append(txsByValue[tx.Value], index)
	}
}

func getTxsByFromAddr(from string) []Transaction {
	var txs []Transaction = []Transaction{}
	for _, txIndex := range txsByFromAddr[from] {
		txs = append(txs, transactions[txIndex])
	}

	return txs
}

func getTxsByToAddr(to string) []Transaction {
	var txs []Transaction = []Transaction{}
	for _, txIndex := range txsByToAddr[to] {
		txs = append(txs, transactions[txIndex])
	}

	return txs
}

func getTxsByFromAndToAddr(from string, to string) []Transaction {
	txsFrom := getTxsByFromAddr(from)
	txsTo := getTxsByToAddr(to)
	lenTxsFrom := len(txsFrom)
	lenTxsTo := len(txsTo)
	var txs []Transaction = []Transaction{}

	// For optimization purposes only loop thru the shorter list
	if lenTxsFrom < lenTxsTo {
		for _, tx := range txsFrom {
			if tx.To == to {
				txs = append(txs, tx)
			}
		}
	} else {
		for _, tx := range txsTo {
			if tx.From == from {
				txs = append(txs, tx)
			}
		}
	}

	return txs
}

func getTxsByValue(aboveValueStr string) []Transaction {
	var txs []Transaction = []Transaction{}
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

	return txs
}

func computePagination(offsetStr string, limitStr string, txs []Transaction) (int, int) {
	txsCount := len(txs)

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = txsCount // Default to returning all txs
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0 // Default to starting at the first item
	}

	end := offset + limit
	if end > txsCount {
		end = txsCount
	}

	return offset, end
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	router.GET("/etherscan-data", func(ctx *gin.Context) {
		page := ctx.Query("page")
		offset := ctx.Query("offset")

		urlBuild, err := url.Parse(apiUrl)
		if err != nil {
			// TODO: handle error
			panic("API base url is not a valid url")
		}

		query := urlBuild.Query()
		if page != "" {
			query.Set("page", page)
		}
		if offset == "" {
			// Etherscan's API uses offset as limit. Set default limit to 100
			offset = "100"
		}
		query.Set("offset", offset)
		query.Set("apiKey", etherscanApiKey)
		urlBuild.RawQuery = query.Encode()

		urlStr := urlBuild.String()
		txs := getEtherscanData(urlStr)
		fmt.Println(len(txs))
		populateDbs(txs)
	})

	router.GET("/transactions", func(ctx *gin.Context) {
		from := ctx.Query("from")
		to := ctx.Query("to")
		aboveValueStr := ctx.Query("aboveValue")
		offsetStr := ctx.Query("offset")
		limitStr := ctx.Query("limit")
		// TODO: input validation for the query params

		var txs []Transaction = []Transaction{}

		if from != "" && to != "" {
			txs = getTxsByFromAndToAddr(from, to)
		} else if from != "" {
			txs = getTxsByFromAddr(from)
		} else if to != "" {
			txs = getTxsByToAddr(to)
		} else if aboveValueStr != "" {
			txs = getTxsByValue(aboveValueStr)
		} else {
			// no query params were provided, so return ALL txs
			txs = transactions
		}

		offset, end := computePagination(offsetStr, limitStr, txs)
		ctx.JSON(http.StatusOK, txs[offset:end])
	})

	return router
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Could not load env vars from .env file")
	}
	etherscanApiKey = os.Getenv("ETHERSCAN_API_KEY")

	r := setupRouter()
	r.Run(":8080")
}
