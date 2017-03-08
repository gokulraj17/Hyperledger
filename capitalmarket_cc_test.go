package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func TestSample(t *testing.T) {
	stub := shim.NewMockStub("mock", new(CapitalMarketChainCode))
	if stub == nil {
		fmt.Println("error")
	}
	fmt.Println("ok")
	//var jsonBlob = `{"gmwID": "9845333333", "emailID": "nagaraja@gmw.com"}`
	//var jsonBlob = `[{ "FIID": "user_type1_0", "brokerID" : "user_type1_1", "CustodianBankID":"user_type1_2", "AccountID":"1111"} ,
	//	{ "FIID": "user_type1_0", "brokerID" : "user_type1_1", "CustodianBankID":"user_type1_2", "AccountID":"1111"},
	//	{ "FIID": "user_type1_0", "brokerID" : "user_type1_1","CustodianBankID":"user_type1_2", "AccountID":"1111"},
	//	{ "FIID": "user_type1_0", "brokerID" : "user_type1_1","CustodianBankID":"user_type1_2", "AccountID":"1111"}
	//]`
	
	var jsonBlob = `[{ "FIID": "FI1", "CustodianBankID":"user_type1_2","BrokerID" : "user_type1_1" ,"AccountID":"Acct001","Product":"TCS", "StockID":"1111", "Quantity":100,"Exchange":"BSE","OrderValidity":"1","OrderType":"BUY","LimitPrice":2},
	{ "FIID": "FI2", "CustodianBankID":"user_type1_2","BrokerID" : "user_type1_1" ,"AccountID":"Acct002","Product":"TCS", "StockID":"1111", "Quantity":100,"Exchange":"BSE","OrderValidity":"1","OrderType":"SELL","LimitPrice":2},
	{ "FIID": "FI3", "CustodianBankID":"user_type1_2","BrokerID" : "user_type1_1" ,"AccountID":"Acct003","Product":"INFY", "StockID":"3333", "Quantity":50,"Exchange":"BSE","OrderValidity":"1","OrderType":"BUY","LimitPrice":2},
	{ "FIID": "FI4", "CustodianBankID":"user_type1_2","BrokerID" : "user_type1_1" ,"AccountID":"Acct004","Product":"INFY", "StockID":"3333", "Quantity":50,"Exchange":"BSE","OrderValidity":"1","OrderType":"SELL","LimitPrice":2},
	{ "FIID": "FI1", "CustodianBankID":"user_type1_2","BrokerID" : "user_type1_1" ,"AccountID":"Acct001","Product":"IBM", "StockID":"2222", "Quantity":30,"Exchange":"BSE","OrderValidity":"1","OrderType":"BUY","LimitPrice":2},
	{ "FIID": "FI2", "CustodianBankID":"user_type1_2","BrokerID" : "user_type1_1" ,"AccountID":"Acct002","Product":"IBM", "StockID":"2222", "Quantity":30,"Exchange":"BSE","OrderValidity":"1","OrderType":"SELL","LimitPrice":2}
	]`;
	var fiid = "FI1"
	var status = "orderPlaced"
	var brokerid = "user_type1_1"
	//var params =
	params := []string{}
	checkInit(t, stub, "init", params)
	checkInvoke(t, stub, "setAllInitialTransactions", time.Now().Format(time.RFC3339))
	checkInvoke(t, stub, "createOrdersByFI", jsonBlob)
	
	checkQuery(t, stub, "getAllOrdersForFIMap")
	checkQueryByID(t, stub, "getAllOrdersForFIBasedOnStatus", fiid, status)
	checkQueryByID(t, stub, "getAllOrdersForBrokerBasedOnStatus", brokerid, status)
	jsonBlob = `["10000","10001","10002", "10003"]`
	checkInvoke(t, stub, "processFIOrdersForConfirmationBySE", jsonBlob)
	//checkQueryBySE(t, stub, "getAllOrdersForSEBasedOnStatus", status)
	checkQueryByID(t, stub, "getAllOrdersForFIBasedOnStatus", fiid, "orderConfirmed")
	
	jsonBlob = `["30000","30001"]`
	checkInvoke(t, stub, "clearAndSettleTrades", time.Now().Format(time.RFC3339))
	//checkQueryBySE(t, stub, "getAllOrdersForCustodianBasedOnStatus", "user_type1_2")
	
	checkQueryBySE(t, stub, "getAllOrdersForSEBasedOnStatus", "toBeSettled")
	
	/*jsonBlob = `["30000","30001","30002"]`
	checkInvoke(t, stub, "clearAndSettleTrades", time.Now().Format(time.RFC3339))*/
	checkQueryBySE(t, stub, "getAllOrdersForCustodianBasedOnStatus", "user_type1_2")
	
	checkQueryBySE(t, stub, "getAllHoldingsForFI", fiid)
	checkQueryBySE(t, stub, "getAllHoldingsForFI", "FI2")
	checkQueryBySE(t, stub, "getAllHoldingsForFI", "FI3")
	checkQueryBySE(t, stub, "getAllHoldingsForFI", "FI4")
	checkQueryByID(t, stub, "getAllTransactionsForFIBasedOnStockId", fiid, "TCS")
	checkQueryByID(t, stub, "getAllTransactionsForFIBasedOnStockId", fiid, "IBM")
	checkQueryByID(t, stub, "getAllTransactionsForFIBasedOnStockId", "FI2", "TCS")
	checkQueryByID(t, stub, "getAllTransactionsForFIBasedOnStockId", "FI2", "IBM")
	checkQueryByID(t, stub, "getAllTransactionsForFIBasedOnStockId", "FI3", "INFY")
	checkQueryByID(t, stub, "getAllTransactionsForFIBasedOnStockId", "FI3", "GOOGLE")
	checkQueryByID(t, stub, "getAllTransactionsForFIBasedOnStockId", "FI4", "INFY")
	checkQueryByID(t, stub, "getAllTransactionsForFIBasedOnStockId", "FI4", "GOOGLE")
	
	jsonBlob = `["10004","10005"]`
	checkInvoke(t, stub, "processFIOrdersForConfirmationBySE", jsonBlob)
	checkQueryBySE(t, stub, "getAllOrdersForSEBasedOnStatus", "toBeSettled")

}

func checkInit(t *testing.T, stub *shim.MockStub, function string, args []string) {
	_, err := stub.MockInit("1", function, args)
	if err != nil {
		fmt.Println("Init failed", err)
		t.FailNow()
	}
}

func checkQuery(t *testing.T, stub *shim.MockStub, function string) {
	params := []string{}

	fmt.Printf("params : %v \n", params)
	querybytes, err := stub.MockQuery(function, params)

	if err != nil {
		fmt.Println("Query", function, "failed", err)
		t.FailNow()
	}
	if querybytes == nil {
		fmt.Println("Query", function, "failed to get value")
		t.FailNow()
	}

	fmt.Printf("get query bytes : %v \n", querybytes)

}

func checkQueryByID(t *testing.T, stub *shim.MockStub, function string, id string, args string) {
	params := []string{id, args}

	fmt.Println("params : ", params)
	querybytes, err := stub.MockQuery(function, params)

	if err != nil {
		fmt.Println("Query", function, "failed", err)
		t.FailNow()
	}
	if querybytes == nil {
		fmt.Println("Query", function, "failed to get value")
		t.FailNow()
	}

	fmt.Printf("get query bytes : %v\n", querybytes)

}

func checkQueryBySE(t *testing.T, stub *shim.MockStub, function string, args string) {
	params := []string{args}

	fmt.Println("params : ", params)
	querybytes, err := stub.MockQuery(function, params)

	if err != nil {
		fmt.Println("Query", function, "failed", err)
		t.FailNow()
	}
	if querybytes == nil {
		fmt.Println("Query", function, "failed to get value")
		t.FailNow()
	}

	fmt.Printf("get query bytes : %v\n", querybytes)

}

func checkInvoke(t *testing.T, stub *shim.MockStub, function string, args string) {
	params := []string{function, args}
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(args), &data)
	fmt.Printf("in checkInvoke, data = %v\n", data)
	_, err := stub.MockInvoke("1", function, params)
	if err != nil {
		fmt.Println("Invoke", params, "failed", err)
		t.FailNow()
	}
}

func checkInvokeWithID(t *testing.T, stub *shim.MockStub, function string, args string, ID string) {
	params := []string{function, args, ID}
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(args), &data)
	fmt.Printf("in checkInvokeWithID, data = %v\n", data)
	_, err := stub.MockInvoke("1", function, params)
	if err != nil {
		fmt.Println("Invoke", params, "failed", err)
		t.FailNow()
	}
}

func checkInvokeWithTime(t *testing.T, stub *shim.MockStub, function string, args string, tm string) {
	params := []string{function, args, tm}
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(args), &data)
	fmt.Printf("in checkInvokeWithID, data = %v\n", data)
	_, err := stub.MockInvoke("1", function, params)
	if err != nil {
		fmt.Println("Invoke", params, "failed", err)
		t.FailNow()
	}
}
