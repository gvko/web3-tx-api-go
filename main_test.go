package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetEtherscanData(t *testing.T) {
	// Create a mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send a JSON response similar to the real Etherscan API
		response := ApiResponse{
			Status:  "1",
			Message: "OK",
			Transactions: []Transaction{
				{From: "a", To: "b", Value: "10"},
				{From: "b", To: "c", Value: "20"},
				{From: "c", To: "d", Value: "30"},
			},
		}

		json.NewEncoder(w).Encode(&response)
	}))
	defer ts.Close()

	// Call the function with the mock server URL
	apiUrl := ts.URL + "/api?module=account&action=tokentx&address=0x1234567890123456789012345678901234567890&sort=asc"
	txs := getEtherscanData(apiUrl)

	// Check that the returned transactions match the mock response
	if len(txs) != 3 {
		t.Errorf("Expected 3 transactions, but got %v", len(txs))
	}

	if txs[0].From != "a" {
		t.Errorf("Expected first transaction to be from 'a', but got from '%v'", txs[0].From)
	}

	if txs[1].To != "c" {
		t.Errorf("Expected second transaction to be to 'c', but got to '%v'", txs[1].To)
	}

	if txs[2].Value != "30" {
		t.Errorf("Expected third transaction value to be '30', but got '%v'", txs[2].Value)
	}
}

func TestPopulateDbs(t *testing.T) {
	txs := []Transaction{
		{From: "a", To: "b", Value: "10"},
		{From: "b", To: "c", Value: "20"},
		{From: "c", To: "d", Value: "30"},
	}

	populateDbs(txs)

	if len(transactions) != 3 {
		t.Errorf("Expected transactions length to be 3, but got %v", len(transactions))
	}

	if len(txsByFromAddr["a"]) != 1 {
		t.Errorf("Expected txsByFromAddr['a'] length to be 1, but got %v", len(txsByFromAddr["a"]))
	}

	if len(txsByToAddr["b"]) != 1 {
		t.Errorf("Expected txsByToAddr['b'] length to be 1, but got %v", len(txsByToAddr["b"]))
	}

	if len(txsByValue["10"]) != 1 {
		t.Errorf("Expected txsByValue['10'] length to be 1, but got %v", len(txsByValue["10"]))
	}
}

func TestGetTxsByFromAddr(t *testing.T) {
	txs := []Transaction{
		{From: "a", To: "b", Value: "10"},
		{From: "b", To: "c", Value: "20"},
		{From: "a", To: "d", Value: "30"},
	}

	populateDbs(txs)

	// Test getTxsByFromAddr() for address 'a'
	expectedTxs := []Transaction{
		{From: "a", To: "b", Value: "10"},
		{From: "a", To: "d", Value: "30"},
	}
	gotTxs := getTxsByFromAddr("a")
	if !reflect.DeepEqual(gotTxs, expectedTxs) {
		t.Errorf("getTxsByFromAddr('a') returned unexpected results. Got %v, expected %v", gotTxs, expectedTxs)
	}

	// Test getTxsByFromAddr() for address 'b'
	expectedTxs = []Transaction{
		{From: "b", To: "c", Value: "20"},
	}
	gotTxs = getTxsByFromAddr("b")
	if !reflect.DeepEqual(gotTxs, expectedTxs) {
		t.Errorf("getTxsByFromAddr('b') returned unexpected results. Got %v, expected %v", gotTxs, expectedTxs)
	}

	// Test getTxsByFromAddr() for address 'c' (no transactions)
	expectedTxs = []Transaction{}
	gotTxs = getTxsByFromAddr("c")
	if !reflect.DeepEqual(gotTxs, expectedTxs) {
		t.Errorf("getTxsByFromAddr('c') returned unexpected results. Got %v, expected %v", gotTxs, expectedTxs)
	}
}

func TestGetTxsByToAddr(t *testing.T) {
	txs := []Transaction{
		{From: "a", To: "b", Value: "10"},
		{From: "b", To: "c", Value: "20"},
		{From: "a", To: "d", Value: "30"},
	}

	populateDbs(txs)

	// test case 1: valid "to" address
	expectedTxs1 := []Transaction{{From: "a", To: "b", Value: "10"}}
	actualTxs1 := getTxsByToAddr("b")
	if !reflect.DeepEqual(actualTxs1, expectedTxs1) {
		t.Errorf("expected %v, but got %v", expectedTxs1, actualTxs1)
	}

	// test case 2: invalid "to" address
	expectedTxs2 := []Transaction{}
	actualTxs2 := getTxsByToAddr("e")
	if !reflect.DeepEqual(actualTxs2, expectedTxs2) {
		t.Errorf("expected %v, but got %v", expectedTxs2, actualTxs2)
	}
}

func TestGetTxsByFromAndToAddr(t *testing.T) {
	txs := []Transaction{
		{From: "a", To: "b", Value: "10"},
		{From: "a", To: "c", Value: "20"},
		{From: "b", To: "c", Value: "30"},
		{From: "c", To: "d", Value: "40"},
	}

	populateDbs(txs)

	// Test when txsByFromAddr is shorter
	result := getTxsByFromAndToAddr("a", "c")
	expectedResult := []Transaction{
		{From: "a", To: "c", Value: "20"},
	}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	// Test when txsByToAddr is shorter
	result = getTxsByFromAndToAddr("b", "c")
	expectedResult = []Transaction{
		{From: "b", To: "c", Value: "30"},
	}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	// Test when both txsByFromAddr and txsByToAddr have equal length
	result = getTxsByFromAndToAddr("a", "b")
	expectedResult = []Transaction{
		{From: "a", To: "b", Value: "10"},
	}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	// Test when no transaction exists between txsByFromAddr and txsByToAddr
	result = getTxsByFromAndToAddr("a", "d")
	expectedResult = []Transaction{}
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}
}
