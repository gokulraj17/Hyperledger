/*
Copyright 2017 IBM, Infosys Ltd.

Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type orderStatusValue int
type tradeStatusValue int

const (
	orderPlaced orderStatusValue = iota + 1
	orderConfirmed
)

const (
	toBeSettled tradeStatusValue = iota + 1
	settledCleared
)

//FIOrder is created for trade requests received by FI
type FIOrder struct {
	FIOrderID        string           `json:"fiOrderID"`       // auto-generated unique ID for the FI Order
	ConfirmedOrderID string           `json:"conOrderID"`      // auto-generated unique ID counterConfirmedIDonfirmed Order
	FIID             string           `json:"fiID"`            // Unique ID of the FI
	CustodianBankID  string           `json:"custodianBankID"` // Unique ID of the Custodian Bank
	BrokerID         string           `json:"brokerID"`        // Unique ID of the broker
	AccountID        string           `json:"accountID"`       // Account ID of the FI
	Product          string           `json:"product"`         // name of the Product
	Status           orderStatusValue `json:"status"`          // status of the Product
	CreationDate     time.Time        `json:"creationDate"`    // date of creation of FIOrder
	StockID          string           `json:"stockID"`         // name of the Stock
	Quantity         int              `json:"quantity"`        // quantity of stock to be bought/sold
	Exchange         string           `json:"exchange"`        // name of exchange
	OrderValidity    string           `json:"orderValidity"`   // validity of the order
	OrderType        string           `json:"orderType"`       // type of Order
	LimitPrice       float32          `json:"limitPrice"`      // limit price
	TradeObjectID   string           `json:"tradeObjectID"`    // auto-generated unique ID for the Trade TradeObject
}

// TradeObject Details
type TradeObject struct {
	TradeObjectID    string           `json:"tradeObjectID"`    // auto-generated unique ID for the Trade TradeObject
	SettlementStatus tradeStatusValue `json:"settlementStatus"` // status of the settlement
	OderTradeNumber  string           `json:"oderTradeNumber"`  // OderTradeNumber
	SettlementDate   time.Time        `json:"creationDate"`     // date of settlement
}

// Transaction details
type Transaction struct {
	FIID             string     `json:"fiID"`          // Unique ID of the FI
	TransactionID    string    `json:"transactionID"` // auto-generated unique ID for the Transaction
	AccountID        string    `json:"accountID"`     // account id of the FI
	Product          string   	`json:"product"`         // name of the Product
	StockID          string    `json:"stockID"`       // id of the stock
	Quantity          int       `json:"quantity"`      // quantity of stocks traded
	TransactionDate  time.Time `json:"txnDate"`       // date of Transaction
	TransactionType  string    `json:"txnType"`       // type of txn - debit/credit
	EffectiveBalance int       `json:"balance"`       // effective balance of stocks post transaction
}

type counter struct {
	counterOrderID, counterConfirmedID, counterTradeOrderID, counterTradeObjectID, counterTransactionID int
}

var counterID counter

// AllFIOrders has a list of all orders ==> AllFIOrders[FIOrderID] = FIOrder
var AllFIOrders map[string]FIOrder

// AllOrdersForFI stores the list of all orders for a FI ==> AllOrdersForFI[FIID] = []FIOrderID
var AllOrdersForFI map[string][]string

// AllOrdersForBroker has a list of all orders for a Broker ==> AllOrdersForBroker[BrokerID] = []FIOrderID
var AllOrdersForBroker map[string][]string

// AllTradeObjects has a list of trade objects ==> TradeObject[TradeObjectID] = TradeObject
var AllTradeObjects map[string]TradeObject

// ConfirmedToFIOrder ==> ConfirmedToFIOrder[ConfirmedOrdererdId] = FIOrderID
var ConfirmedToFIOrder map[string]string

// TradeSettlementMap has a list  ==>  TradeSettlementMap[TradeObjectID]=[]ConfirmedOrdererdId *** TO CHECK ****
var TradeSettlementMap map[string][]string

// ListOfTransactions ==> Transaction[TransactionID]=Transaction
var ListOfTransactions map[string]Transaction

// ListOfTransactionsForFI ==> ListOfTransactionsForFI[FIID]=(ListOfStocks[StockID]=[]TransactionID)  *** TO CONFIRM ***
var ListOfTransactionsForFI map[string]map[string][]string //or[]TransactionID

// CapitalMarketChainCode defined the chaincode for global mobile wallet
type CapitalMarketChainCode struct {
}

var err error
var bytesArray []byte

func generateOrderID(stub shim.ChaincodeStubInterface) (string, error) {
	var id int
	_, err = getCounters(stub)
	id = counterID.counterOrderID
	if err != nil {
		fmt.Printf("Failed to get the ID for block chain :%v\n", err)
		return "", err
	}
	counterID.counterOrderID = counterID.counterOrderID + 1
	_, err = setCounters(stub)
	if err != nil {
		fmt.Printf("Failed to set the ID for block chain :%v\n", err)
		return "", err
	}
	return (strconv.Itoa(id)), nil
}

func generateConfirmedID(stub shim.ChaincodeStubInterface) (string, error) {
	var id int
	_, err = getCounters(stub)
	id = counterID.counterConfirmedID
	if err != nil {
		fmt.Printf("Failed to get the ID for block chain :%v\n", err)
		return "", err
	}
	counterID.counterConfirmedID = counterID.counterConfirmedID + 1
	_, err = setCounters(stub)
	if err != nil {
		fmt.Printf("Failed to set the ID for block chain :%v\n", err)
		return "", err
	}
	return (strconv.Itoa(id)), nil
}

func generateTradeOrderID(stub shim.ChaincodeStubInterface) (string, error) {
	var id int
	_, err = getCounters(stub)
	id = counterID.counterTradeOrderID
	if err != nil {
		fmt.Printf("Failed to get the ID for block chain :%v\n", err)
		return "", err
	}
	counterID.counterTradeOrderID = counterID.counterTradeOrderID + 1
	_, err = setCounters(stub)
	if err != nil {
		fmt.Printf("Failed to set the ID for block chain :%v\n", err)
		return "", err
	}
	return (strconv.Itoa(id)), nil
}

func generateTradeObjectID(stub shim.ChaincodeStubInterface) (string, error) {
	var id int
	_, err = getCounters(stub)
	id = counterID.counterTradeObjectID
	if err != nil {
		fmt.Printf("Failed to get the ID for block chain :%v\n", err)
		return "", err
	}
	counterID.counterTradeObjectID = counterID.counterTradeObjectID + 1
	_, err = setCounters(stub)
	if err != nil {
		fmt.Printf("Failed to set the ID for block chain :%v\n", err)
		return "", err
	}
	return (strconv.Itoa(id)), nil
}

func generateTransactionID(stub shim.ChaincodeStubInterface) (string, error) {
	var id int
	_, err = getCounters(stub)
	id = counterID.counterTransactionID
	if err != nil {
		fmt.Printf("Failed to get the ID for block chain :%v\n", err)
		return "", err
	}
	counterID.counterTransactionID = counterID.counterTransactionID + 1
	_, err = setCounters(stub)
	if err != nil {
		fmt.Printf("Failed to set the ID for block chain :%v\n", err)
		return "", err
	}
	return (strconv.Itoa(id)), nil
}

// get the Counter
func getCounters(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("counterID")
	if err != nil {
		fmt.Printf("Failed to initialize the counterID for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("counterID exists.\n")
		err = json.Unmarshal(bytesArray, &counterID)
		if err != nil {
			fmt.Printf("Failed to initialize the counterID for block chain :%v\n", err)
			return nil, err
		}
	} else { // create counterID
		fmt.Printf("counterID does not exist. To be created\n")
		counterID.counterOrderID = 10000
		counterID.counterConfirmedID = 20000
		counterID.counterTradeObjectID = 30000
		counterID.counterTradeOrderID = 40000
		counterID.counterTransactionID = 50000
		bytesArray, err = json.Marshal(&counterID)
		if err != nil {
			fmt.Printf("Failed to initialize the counterID for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("counterID", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the counterID for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("Initiliazed counterID : %v\n", counterID)
	return nil, err
}

// set the Counter
func setCounters(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&counterID)
	if err != nil {
		fmt.Printf("Failed to set the AllFIOrders for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("counterID", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the Counters for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

// get the ListOfTransactionsForFI
func getListOfTransactionsForFI(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("ListOfTransactionsForFI")
	if err != nil {
		fmt.Printf("Failed to initialize the ListOfTransactionsForFI for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("ListOfTransactionsForFI map exists.\n")
		err = json.Unmarshal(bytesArray, &ListOfTransactionsForFI)
		if err != nil {
			fmt.Printf("Failed to initialize the ListOfTransactionsForFI for block chain :%v\n", err)
			return nil, err
		}
	} else { // create a new map for ListOfTransactionsForFI
		fmt.Printf("ListOfTransactionsForFI map does not exist. To be created. \n")
		ListOfTransactionsForFI = make(map[string]map[string][]string)
		bytesArray, err = json.Marshal(&ListOfTransactionsForFI)
		if err != nil {
			fmt.Printf("Failed to initialize the ListOfTransactionsForFI for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("ListOfTransactionsForFI", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the ListOfTransactionsForFI for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("ListOfTransactionsForFI : %v\n", ListOfTransactionsForFI)
	return nil, err
}

// set the ListOfTransactions
func setListOfTransactionsForFI(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&ListOfTransactionsForFI)
	if err != nil {
		fmt.Printf("Failed to set the ListOfTransactionsForFI for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("ListOfTransactionsForFI", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the ListOfTransactionsForFI for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

// get the ListOfTransactions
func getListOfTransactions(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("ListOfTransactions")
	if err != nil {
		fmt.Printf("Failed to initialize the ListOfTransactions for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("ListOfTransactions map exists.\n")
		err = json.Unmarshal(bytesArray, &ListOfTransactions)
		if err != nil {
			fmt.Printf("Failed to initialize the ListOfTransactions for block chain :%v\n", err)
			return nil, err
		}
	} else { // create a new map for ListOfTransactions
		fmt.Printf("ListOfTransactions map does not exist. To be created. \n")
		ListOfTransactions = make(map[string]Transaction)
		bytesArray, err = json.Marshal(&ListOfTransactions)
		if err != nil {
			fmt.Printf("Failed to initialize the ListOfTransactions for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("ListOfTransactions", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the ListOfTransactions for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("ListOfTransactions : %v\n", ListOfTransactions)
	return nil, err
}

// set the ListOfTransactions
func setListOfTransactions(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&ListOfTransactions)
	if err != nil {
		fmt.Printf("Failed to set the ListOfTransactions for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("ListOfTransactions", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the ListOfTransactions for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

// get the TradeSettlementMap
func getTradeSettlementMap(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("TradeSettlementMap")
	if err != nil {
		fmt.Printf("Failed to initialize the TradeSettlementMap for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("TradeSettlementMap map exists.\n")
		err = json.Unmarshal(bytesArray, &TradeSettlementMap)
		if err != nil {
			fmt.Printf("Failed to initialize the TradeSettlementMap for block chain :%v\n", err)
			return nil, err
		}
	} else { // create a new map for TradeSettlementMap
		fmt.Printf("TradeSettlementMap map does not exist. To be created. \n")
		TradeSettlementMap = make(map[string][]string)
		bytesArray, err = json.Marshal(&TradeSettlementMap)
		if err != nil {
			fmt.Printf("Failed to initialize the TradeSettlementMap for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("TradeSettlementMap", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the TradeSettlementMap for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("TradeSettlementMap : %v\n", TradeSettlementMap)
	return nil, err
}

// set the TradeSettlementMap
func setTradeSettlementMap(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&TradeSettlementMap)
	if err != nil {
		fmt.Printf("Failed to set the TradeSettlementMap for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("TradeSettlementMap", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the TradeSettlementMap for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

// get the ConfirmedToFIOrder
func getConfirmedToFIOrder(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("ConfirmedToFIOrder")
	if err != nil {
		fmt.Printf("Failed to initialize the ConfirmedToFIOrder for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("ConfirmedToFIOrder map exists.\n")
		err = json.Unmarshal(bytesArray, &ConfirmedToFIOrder)
		if err != nil {
			fmt.Printf("Failed to initialize the ConfirmedToFIOrder for block chain :%v\n", err)
			return nil, err
		}
	} else { // create a new map for AllTradeObjects
		fmt.Printf("ConfirmedToFIOrder map does not exist. To be created. \n")
		ConfirmedToFIOrder = make(map[string]string)
		bytesArray, err = json.Marshal(&ConfirmedToFIOrder)
		if err != nil {
			fmt.Printf("Failed to initialize the ConfirmedToFIOrder for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("ConfirmedToFIOrder", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the ConfirmedToFIOrder for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("Initiliazed ConfirmedToFIOrder : %v\n", ConfirmedToFIOrder)
	return nil, err
}

// set the TradeSettlementMap
func setConfirmedToFIOrder(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&ConfirmedToFIOrder)
	if err != nil {
		fmt.Printf("Failed to set the ConfirmedToFIOrder for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("ConfirmedToFIOrder", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the ConfirmedToFIOrder for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

// get the Trade Object Map
func getAllTradeObjects(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("AllTradeObjects")
	if err != nil {
		fmt.Printf("Failed to initialize the AllTradeObjects for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("All Trade Objects map exists.\n")
		err = json.Unmarshal(bytesArray, &AllTradeObjects)
		if err != nil {
			fmt.Printf("Failed to initialize the AllTradeObjects for block chain :%v\n", err)
			return nil, err
		}
	} else { // create a new map for AllTradeObjects
		fmt.Printf("All Trade Objects map does not exist. To be created. \n")
		AllTradeObjects = make(map[string]TradeObject)
		bytesArray, err = json.Marshal(&AllTradeObjects)
		if err != nil {
			fmt.Printf("Failed to initialize the AllTradeObjects for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("AllTradeObjects", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the AllTradeObjects for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("Initiliazed AllTradeObjects : %v\n", AllTradeObjects)
	return nil, err
}

// set the TradeSettlementMap
func setAllTradeObjects(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&AllTradeObjects)
	if err != nil {
		fmt.Printf("Failed to set the AllTradeObjects for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("AllTradeObjects", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the AllTradeObjects for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

// get the AllOrdersForBroker
func getAllOrdersForBroker(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("AllOrdersForBroker")

	if err != nil {
		fmt.Printf("Failed to initialize the AllOrdersForBroker for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("AllOrdersForBroker map exists.\n")
		err = json.Unmarshal(bytesArray, &AllOrdersForBroker)
		if err != nil {
			fmt.Printf("Failed to initialize the AllOrdersForBroker for block chain :%v\n", err)
			return nil, err
		}
	} else { // create a new map for AllOrdersForBroker
		fmt.Printf("AllOrdersForBroker map does not exist. To be created.\n")
		AllOrdersForBroker = make(map[string][]string)
		bytesArray, err = json.Marshal(&AllOrdersForBroker)
		if err != nil {
			fmt.Printf("Failed to initialize the AllOrdersForBroker for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("AllOrdersForBroker", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the AllOrdersForBroker for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("Initiliazed AllOrdersForBroker : %v\n", AllOrdersForBroker)
	return nil, err
}

// set the AllOrdersForBroker
func setAllOrdersForBroker(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&AllOrdersForBroker)
	if err != nil {
		fmt.Printf("Failed to set the AllOrdersForBroker for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("AllOrdersForBroker", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the AllOrdersForBroker for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

// get the AllOrdersForFI Map
func getAllOrdersForFI(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("AllOrdersForFI")
	if err != nil {
		fmt.Printf("Failed to initialize the AllOrdersForFI for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("AllOrdersForFI map exists.\n")
		err = json.Unmarshal(bytesArray, &AllOrdersForFI)
		if err != nil {
			fmt.Printf("Failed to initialize the AllOrdersForFI for block chain :%v\n", err)
			return nil, err
		}
	} else { // create a new map for AllOrdersForFI
		fmt.Printf("AllOrdersForFI map does not exist. To be created")
		AllOrdersForFI = make(map[string][]string)
		bytesArray, err = json.Marshal(&AllOrdersForFI)
		if err != nil {
			fmt.Printf("Failed to initialize the AllOrdersForFI for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("AllOrdersForFI", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the AllOrdersForFI for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("Initiliazed AllOrdersForFI : %v\n", AllOrdersForFI)
	return nil, err
}

// set the AllOrdersForFI
func setAllOrdersForFI(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&AllOrdersForFI)
	if err != nil {
		fmt.Printf("Failed to set the AllOrdersForFI for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("AllOrdersForFI", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the AllOrdersForFI for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

// get the AllFIOrders Map
func getAllFIOrders(stub shim.ChaincodeStubInterface) ([]byte, error) {

	bytesArray, err = stub.GetState("AllFIOrders")
	if err != nil {
		fmt.Printf("Failed to initialize the AllFIOrders for block chain :%v\n", err)
		return nil, err
	}
	if len(bytesArray) != 0 {
		fmt.Printf("AllFIOrders map exists.\n")
		err = json.Unmarshal(bytesArray, &AllFIOrders)
		if err != nil {
			fmt.Printf("Failed to initialize the AllFIOrders for block chain :%v\n", err)
			return nil, err
		}
	} else { // create a new map for AllFIOrders
		fmt.Printf("AllFIOrders map does not exist. To be created\n")
		AllFIOrders = make(map[string]FIOrder)
		bytesArray, err = json.Marshal(&AllFIOrders)
		if err != nil {
			fmt.Printf("Failed to initialize the AllFIOrders for block chain :%v\n", err)
			return nil, err
		}
		err = stub.PutState("AllFIOrders", bytesArray)
		if err != nil {
			fmt.Printf("Failed to initialize the AllFIOrders for block chain :%v\n", err)
			return nil, err
		}
	}
	fmt.Printf("Initiliazed AllFIOrders : %v\n", AllFIOrders)
	return nil, err
}

// set the AllFIOrders
func setAllFIOrders(stub shim.ChaincodeStubInterface) ([]byte, error) {
	bytesArray, err = json.Marshal(&AllFIOrders)
	if err != nil {
		fmt.Printf("Failed to set the AllFIOrders for block chain :%v\n", err)
		return nil, err
	}
	err = stub.PutState("AllFIOrders", bytesArray)
	if err != nil {
		fmt.Printf("Failed to set the AllFIOrders for block chain :%v\n", err)
		return nil, err
	}
	return nil, err
}

//setting up initial transactions
func(t *CapitalMarketChainCode) setAllInitialTransactions(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("Setting all initial transactions\n")
	fmt.Printf("len args: %d\n", len(args))
	if len(args) != 2 {
		fmt.Printf("Incorrect number of arguments.\n")
		return nil, errors.New("Incorrect number of arguments")
	}
	fmt.Printf("args[0]: %v\n", args[0])
	fmt.Printf("args[1]: %v\n", args[1])
	
	txTime, err := time.Parse(time.RFC3339, args[1])
	fmt.Printf("initial transactions time is %s\n", txTime)
	if err != nil {
		fmt.Println("Error in parsing time while setting the initial transactions %s", err)
		return nil, errors.New("Failed to set initial transactions")
	}
	
	getListOfTransactions(stub)
	getListOfTransactionsForFI(stub)
	
	if len(ListOfTransactions) < 1 {
	
		var trObj1 = Transaction{FIID : "FI1", TransactionID : "12341", AccountID : "Acct001", Product : "TCS", StockID : "1111", Quantity : 1000, TransactionDate : txTime, TransactionType : "CREDIT", EffectiveBalance : 1000}
		ListOfTransactions["12341"] = trObj1
		
		var trObj2 = Transaction{FIID : "FI2", TransactionID : "12342", AccountID : "Acct002", Product : "IBM", StockID : "2222", Quantity : 1000, TransactionDate : txTime, TransactionType : "CREDIT", EffectiveBalance : 1000}
		ListOfTransactions["12342"] = trObj2
		
		var trObj3 = Transaction{FIID : "FI3", TransactionID : "12343", AccountID : "Acct003", Product : "INFY", StockID : "3333", Quantity : 1000, TransactionDate : txTime, TransactionType : "CREDIT", EffectiveBalance : 1000}
		ListOfTransactions["12343"] = trObj3
		
		var trObj4 = Transaction{FIID : "FI4", TransactionID : "12344", AccountID : "Acct004", Product : "GOOGLE", StockID : "4444", Quantity : 1000, TransactionDate : txTime, TransactionType : "CREDIT", EffectiveBalance : 1000}
		ListOfTransactions["12344"] = trObj4
		
		var trObj5 = Transaction{FIID : "FI1", TransactionID : "12345", AccountID : "Acct001", Product : "IBM", StockID : "2222", Quantity : 500, TransactionDate : txTime, TransactionType : "CREDIT", EffectiveBalance : 500}
		ListOfTransactions["12345"] = trObj5
		
		var trObj6 = Transaction{FIID : "FI2", TransactionID : "12346", AccountID : "Acct002", Product : "TCS", StockID : "1111", Quantity : 500, TransactionDate : txTime, TransactionType : "CREDIT", EffectiveBalance : 500}
		ListOfTransactions["12346"] = trObj6
		
		var trObj7 = Transaction{FIID : "FI3", TransactionID : "12347", AccountID : "Acct003", Product : "GOOGLE", StockID : "4444", Quantity : 500, TransactionDate : txTime, TransactionType : "CREDIT", EffectiveBalance : 500}
		ListOfTransactions["12347"] = trObj7
		
		var trObj8 = Transaction{FIID : "FI4", TransactionID : "12348", AccountID : "Acct004", Product : "INFY", StockID : "3333", Quantity : 500, TransactionDate : txTime, TransactionType : "CREDIT", EffectiveBalance : 500}
		ListOfTransactions["12348"] = trObj8
	
		_, err = setListOfTransactions(stub)
		if err != nil {
			// remove the value from ListOfTransactions
			// remove the value from ListOfTransactions
			fmt.Printf("Error to set map ListOfTransactions : %v\n", err)
			return nil, errors.New("Failed to create list of transactions")
		}
		
		getListOfTransactions(stub)
	}
	
	if len(ListOfTransactionsForFI) < 1 {
	
		var ListOfStocks11 map[string][]string
		var ListOfStocks12 map[string][]string
		var ListOfStocks21 map[string][]string
		var ListOfStocks22 map[string][]string
		ListOfStocks11 = make(map[string][]string)
		ListOfStocks12 = make(map[string][]string)
		ListOfStocks21 = make(map[string][]string)
		ListOfStocks22 = make(map[string][]string)
		
		ListOfStocks11["TCS"] = append(ListOfStocks11["TCS"], "12341")
		ListOfStocks11["IBM"] = append(ListOfStocks11["IBM"], "12345")
		ListOfTransactionsForFI["FI1"] = ListOfStocks11
		
		ListOfStocks12["TCS"] = append(ListOfStocks12["TCS"], "12346")
		ListOfStocks12["IBM"] = append(ListOfStocks12["IBM"], "12342")
		ListOfTransactionsForFI["FI2"] = ListOfStocks12
		
		ListOfStocks21["INFY"] = append(ListOfStocks21["INFY"], "12343")
		ListOfStocks21["GOOGLE"] = append(ListOfStocks21["GOOGLE"], "12347")
		ListOfTransactionsForFI["FI3"] = ListOfStocks21
		
		ListOfStocks22["INFY"] = append(ListOfStocks22["INFY"], "12348")
		ListOfStocks22["GOOGLE"] = append(ListOfStocks22["GOOGLE"], "12344")
		ListOfTransactionsForFI["FI4"] = ListOfStocks22
		
		_, err = setListOfTransactionsForFI(stub)
		if err != nil {
			// remove the value from ListOfTransactionsForFI
			// remove the value from ListOfTransactionsForFI
			fmt.Printf("Error to set map ListOfTransactionsForFI : %v\n", err)
			return nil, errors.New("Failed to create list of transactions for fi")
		}
		
		getListOfTransactionsForFI(stub)
	
	}
	
	fmt.Println("ListOfTransactions are :")
	
	for k,v := range ListOfTransactions {
		fmt.Println("key is %s", k)
		fmt.Printf("%+v\n", v)
	}
	
	return nil, err
}

// Init function
func (t *CapitalMarketChainCode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	getAllFIOrders(stub)
	getAllOrdersForFI(stub)
	getAllOrdersForBroker(stub)
	getAllTradeObjects(stub)
	getConfirmedToFIOrder(stub)
	getTradeSettlementMap(stub)
	getCounters(stub)
	getListOfTransactions(stub)
	getListOfTransactionsForFI(stub)
	
	fmt.Println("Initialization complete")

	return nil, err
}

// add orders created by the FI
func (t *CapitalMarketChainCode) createOrdersByFI(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Creating all orders by FI")
	fmt.Printf("len args: %d\n", len(args))
	if len(args) != 2 {
		fmt.Printf("Incorrect number of arguments.\n")
		return nil, errors.New("Incorrect number of arguments")
	}
	fmt.Printf("args[0]: %v\n", args[0])
	fmt.Printf("args[1]: %v\n", args[1])

	var fiOrders []FIOrder
	var err error

	err = json.Unmarshal([]byte(args[1]), &fiOrders)
	if err != nil {
		fmt.Printf("Error unmarshalling fi orders data : %v\n", err)
		return nil, errors.New("Failed to create fi orders")
	}
	fmt.Printf("fi orders after unmarshal: %v\n", fiOrders)

	if len(fiOrders) > 0 {
		// read the maps
		_, err = getAllFIOrders(stub)

		if err != nil {
			fmt.Printf("Error to read map AllFIOrders : %v\n", err)
			return nil, errors.New("Failed to create fi orders")
		}
		_, err = getAllOrdersForBroker(stub)

		if err != nil {
			fmt.Printf("Error to read map AllFIOrders : %v\n", err)
			return nil, errors.New("Failed to create fi orders")
		}
		_, err = getAllOrdersForFI(stub)

		if err != nil {
			fmt.Printf("Error to read map AllFIOrders : %v\n", err)
			return nil, errors.New("Failed to create fi orders")
		}

		for _, fiOrder := range fiOrders {
			fiOrder.FIOrderID, _ = generateOrderID(stub)
			fiOrder.Status = orderPlaced
			AllFIOrders[fiOrder.FIOrderID] = fiOrder
			AllOrdersForBroker[fiOrder.BrokerID] = append(AllOrdersForBroker[fiOrder.BrokerID], fiOrder.FIOrderID)
			AllOrdersForFI[fiOrder.FIID] = append(AllOrdersForFI[fiOrder.FIID], fiOrder.FIOrderID)
		}

		// set the maps
		_, err = setAllFIOrders(stub)
		if err != nil {
			fmt.Printf("Error to set map AllFIOrders : %v\n", err)
			return nil, errors.New("Failed to create fi orders")
		}
		_, err = setAllOrdersForBroker(stub)
		if err != nil {
			// remove the value from AllFIOrders
			fmt.Printf("Error to set map AllOrdersForBroker : %v\n", err)
			return nil, errors.New("Failed to create fi orders")
		}
		_, err = setAllOrdersForFI(stub)
		if err != nil {
			// remove the value from AllFIOrders
			// remove the value from AllOrdersForBroker
			fmt.Printf("Error to set map AllOrdersForBroker : %v\n", err)
			return nil, errors.New("Failed to create fi orders")
		}
		fmt.Printf("Orders created successfully \n")
		return nil, nil
	}
	return nil, errors.New("There are no orders available for the FI")
}

//create matched orders
func findMatchingOrders(stub shim.ChaincodeStubInterface, fiOrders []string) (map[string][]string, error) {
	//var fiMatchedOrders []string
	fmt.Printf("Matching the orders : %v\n", fiOrders)
	var fiMatchedOrders map[string][]string
	var fiOrder FIOrder
	var ok bool
	var effectiveSum map[string]int
	
	// read the maps

	_, err = getAllFIOrders(stub)
	if err != nil {
		fmt.Printf("Error to read map AllFIOrders : %v\n", err)
		return nil, errors.New("Failed to get FIOrder details")
	}
	_, err = getConfirmedToFIOrder(stub)
	if err != nil {
		fmt.Printf("Error to read map ConfirmedToFIOrder : %v\n", err)
		return nil, errors.New("Failed to get FIOrder details")
	}
	effectiveSum = make(map[string]int)
	fiMatchedOrders = make(map[string][]string)
	for _, orderid := range fiOrders {
		
		if fiOrder, ok = AllFIOrders[orderid]; ok {
			if _, ok := effectiveSum[fiOrder.Product]; ok {
				if fiOrder.OrderType == "BUY" {
					effectiveSum[fiOrder.Product] = effectiveSum[fiOrder.Product] + fiOrder.Quantity
				} else if fiOrder.OrderType == "SELL" {
					effectiveSum[fiOrder.Product] = effectiveSum[fiOrder.Product] - fiOrder.Quantity
				}
				
			} else {
				//effectiveSum[fiOrder.Product] = 0
				if fiOrder.OrderType == "BUY" {
					effectiveSum[fiOrder.Product] = effectiveSum[fiOrder.Product] + fiOrder.Quantity
				} else if fiOrder.OrderType == "SELL" {
					effectiveSum[fiOrder.Product] = effectiveSum[fiOrder.Product] - fiOrder.Quantity
				}
			}
			fiMatchedOrders[fiOrder.Product] = append(fiMatchedOrders[fiOrder.Product], fiOrder.FIOrderID)
		}	
	}
	
	for _, v := range effectiveSum {
		fmt.Printf("Effective sum is : %v\n", v)
		if v != 0 {
			fmt.Printf("Error to validate matching orders : %v\n", err)
			return nil, errors.New("Failed to validate matching orders")
		}	
	}

	// generate the confirmed id
	for _, orderid := range fiOrders {
		if fiOrder, ok = AllFIOrders[orderid]; ok {
			fiOrder.Status = orderConfirmed
			fiOrder.ConfirmedOrderID, err = generateConfirmedID(stub)
			AllFIOrders[orderid] = fiOrder
			ConfirmedToFIOrder[fiOrder.ConfirmedOrderID] = fiOrder.FIOrderID
			fmt.Printf("Confirmed orderid : %v\n", fiOrder.ConfirmedOrderID)
			//fiMatchedOrders = append(fiMatchedOrders, fiOrder.ConfirmedOrderID)
		}
	}
	fmt.Printf("Matching the orders : %v\n", fiOrders)
	_, err = setAllFIOrders(stub)
	if err != nil {
		fmt.Printf("Error to set map AllFIOrders : %v\n", err)
		return nil, errors.New("Failed to get FIOrder details")
	}
	_, err = setConfirmedToFIOrder(stub)
	if err != nil {
		// TO DO - undo all the earlier PutState
		fmt.Printf("Error to set map ConfirmedToFIOrder : %v\n", err)
		return nil, errors.New("Failed to get FIOrder details")
	}

	return fiMatchedOrders, err
}

//Create the confirmed Orders
func (t *CapitalMarketChainCode) processFIOrdersForConfirmationBySE(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Creating confirmed orders by Broker")
	fmt.Printf("len args: %d\n", len(args))
	if len(args) != 2 {
		fmt.Printf("Incorrect number of arguments.\n")
		return nil, errors.New("Incorrect number of arguments")
	}
	fmt.Printf("args[0]: %v\n", args[0])
	fmt.Printf("args[1]: %v\n", args[1])

	var fiOrders []string 
	var mappedOrder map[string][]string //mappedOrder []string
	var err error
	var fiOrder FIOrder
	var ok bool

	err = json.Unmarshal([]byte(args[1]), &fiOrders)
	if err != nil {
		fmt.Printf("Error unmarshalling confirmed orders data : %v\n", err)
		return nil, errors.New("Failed to create Trade Object")
	}

	fmt.Printf("fiOrders****** : %v\n", fiOrders)

	// create confirmed order ids
	if len(fiOrders) > 0 {

		mappedOrder, _ = findMatchingOrders(stub, fiOrders)
		if len(mappedOrder) < 1 {
			return nil, errors.New("Matching has not found!!!")
		}
		fmt.Printf("Mapped order is : ");
		for k,v := range mappedOrder{
			fmt.Printf("Product is : %v\n", k)
			fmt.Printf("orderid is : %v\n", v)
		}
		
		// get the maps
		_, err = getAllTradeObjects(stub)
		if err != nil {
			fmt.Printf("Error to read map AllTradeObjects : %v\n", err)
			return nil, errors.New("Failed to get Trade Object")
		}
		_, err = getTradeSettlementMap(stub)
		if err != nil {
			fmt.Printf("Error to read map TradeSettlementMap : %v\n", err)
			return nil, errors.New("Failed to get TradeSettlementMap")
		}
		_, err = getAllFIOrders(stub)
		if err != nil {
			fmt.Printf("Error to read map AllFIOrders : %v\n", err)
			return nil, errors.New("Failed to get AllFIOrders")
		}

		for _, v := range mappedOrder {
			// created TradeObject
			var confOrder []string
			var tradeObject TradeObject
			tradeObject.TradeObjectID, _ = generateTradeObjectID(stub)
			tradeObject.OderTradeNumber, _ = generateTradeOrderID(stub)
			tradeObject.SettlementStatus = toBeSettled
			AllTradeObjects[tradeObject.TradeObjectID] = tradeObject
			for _, id := range v {
				if fiOrder, ok = AllFIOrders[id]; ok {
					fiOrder.TradeObjectID = tradeObject.TradeObjectID
					AllFIOrders[id] = fiOrder
					confOrder = append(confOrder, fiOrder.ConfirmedOrderID)
				}	
			}
			TradeSettlementMap[tradeObject.TradeObjectID] = confOrder
		}

		// set the maps
		
		_, err = setAllFIOrders(stub)
		if err != nil {
			// TO DO - undo all the earlier PutState
			fmt.Printf("Error to set map AllTradeObjects : %v\n", err)
			return nil, errors.New("Failed to create Trade Object")
		}
					
		_, err = setAllTradeObjects(stub)
		if err != nil {
			// TO DO - undo all the earlier PutState
			fmt.Printf("Error to set map AllTradeObjects : %v\n", err)
			return nil, errors.New("Failed to create Trade Object")
		}
		_, err = setTradeSettlementMap(stub)
		if err != nil {
			// TO DO - undo all the earlier PutState
			fmt.Printf("Error to set map TradeSettlementMap : %v\n", err)
			return nil, errors.New("Failed to create Trade Object")
		}
		fmt.Printf("TradeSettlementMap. %v\n", TradeSettlementMap)
		fmt.Printf("TradeSettlementMap updated.\n")
		return nil, nil
	}
	return nil, errors.New("There are no confirmed orders for settlement")
}

//getOrdersReadyForSettlement
func getFIOrdersReadyForSettlement(stub shim.ChaincodeStubInterface, mappedOrderIDs []string) ([]FIOrder, error) {
	var fiorder FIOrder
	var fiOrders []FIOrder
	var orderid string
	var ok bool

	fmt.Printf("****Mapped order is : ");
		for k,v := range mappedOrderIDs{
			fmt.Printf("Product is : %v\n", k)
			fmt.Printf("orderid is : %v\n", v)
		}
	
	// TO DO - add all the error conditions
	_, err = getAllFIOrders(stub)
	if err != nil {
		fmt.Printf("Error to read map AllFIOrders : %v\n", err)
		return nil, errors.New("Failed to get AllFIOrders")
	}
	_, err = getConfirmedToFIOrder(stub)
	if err != nil {
		fmt.Printf("Error to read map ConfirmedToFIOrder : %v\n", err)
		return nil, errors.New("Failed to get ConfirmedToFIOrder")
	}

	for _, corderid := range mappedOrderIDs { // walk through confirmed order array
		if orderid, ok = ConfirmedToFIOrder[corderid]; ok { // get the matching order id
			if fiorder, ok = AllFIOrders[orderid]; ok { // get the order details for the order id
				fiOrders = append(fiOrders, fiorder) // create an array of orders
			}
		}
	}
	return fiOrders, err

}

//clearAndSettleTrades
func (t *CapitalMarketChainCode) clearAndSettleTrades(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Printf("CLearing and settling the trades")
	fmt.Printf("len args: %d\n", len(args))
	if len(args) != 2 {
		fmt.Printf("Incorrect number of arguments.\n")
		return nil, errors.New("Incorrect number of arguments")
	}
	fmt.Printf("args[0]: %v\n", args[0])
	fmt.Printf("args[1]: %v\n", args[1])
	//fmt.Printf("args[1]: %v\n", args[2])
	var err error
	
	txTime, err := time.Parse(time.RFC3339, args[1])
	fmt.Printf("time is %s\n", txTime)
	if err != nil {
		fmt.Println("Error in parsing time while clearing the trade %s", err)
		return nil, errors.New("Failed to clear and settle trade")
	}
	
	var tradeObjectIDs, matchedOrderIDs, txnObjects []string
	var tObject TradeObject
	var orderids []FIOrder
	var ok bool
	var txnObject, lasttxnObject Transaction
	var lasttxnObjectID string

	// TODO add the error handling
	_, err = getTradeSettlementMap(stub)
	_, err = getListOfTransactions(stub)
	_, err = getListOfTransactionsForFI(stub)

	tradeObjectIDs, _ = getAllTradesForSettlement(stub) // get a list of trades ready for settlement

	for _, tObjectid := range tradeObjectIDs { // walk through list of trades ready for settlement
		fmt.Printf("Trade Object IDs are %v\n ",tObjectid);
		if tObject, ok = AllTradeObjects[tObjectid]; ok { // get the trade Object
			if matchedOrderIDs, ok = TradeSettlementMap[tObjectid]; ok { // get the associated mapped Orders
				orderids, _ = getFIOrdersReadyForSettlement(stub, matchedOrderIDs) // get the array of matched FIOrders
				fmt.Printf("check for orderids : %v\n", orderids)
				
				tObject.SettlementStatus = settledCleared
				fmt.Printf("************************************** %v\n",settledCleared)
				tObject.SettlementDate = txTime // TO FIX, the value has to be passed from outside
				fmt.Printf("**********************************666666 %++v\n",AllTradeObjects[tObjectid])
				AllTradeObjects[tObjectid] = tObject
				fmt.Printf("**********************************777777 %++v\n",AllTradeObjects[tObjectid])
				for _, fiOrder := range orderids {

					fmt.Printf("##FI things : %v, %v\n", fiOrder.FIID, fiOrder.StockID)
					fmt.Printf("##FI ListOfTransactions : %v\n", ListOfTransactionsForFI[fiOrder.FIID])
					fmt.Printf("##FI fiOrder : %v\n", fiOrder)
					// getting the list of txns for FIID and Product
					txnObjects = (ListOfTransactionsForFI[fiOrder.FIID])[fiOrder.Product]
					// getting the last txn for the FIID and Product
					
					fmt.Printf("-----Len of Transactions : %v\n", len(txnObjects))
					
					lasttxnObjectID = txnObjects[len(txnObjects)-1]
					// accessing the last txn object from the ListOfTransactions
					lasttxnObject = ListOfTransactions[lasttxnObjectID]
					fmt.Printf("lasttxnObject : %v\n", lasttxnObject)

					// create the new transaction details
					txnObject.FIID = fiOrder.FIID
					txnObject.AccountID = fiOrder.AccountID
					txnObject.StockID = fiOrder.StockID
					txnObject.Product = fiOrder.Product
					txnObject.TransactionID, _ = generateTransactionID(stub)
					txnObject.Quantity = fiOrder.Quantity
					txnObject.TransactionDate = txTime
					
					if fiOrder.OrderType == "BUY" {
						txnObject.EffectiveBalance = lasttxnObject.EffectiveBalance + fiOrder.Quantity
						txnObject.TransactionType = "Credit"
					} else if fiOrder.OrderType == "SELL" {
						txnObject.EffectiveBalance = lasttxnObject.EffectiveBalance - fiOrder.Quantity
						txnObject.TransactionType = "Debit"
					}
					
					// adding the new txn object to the list of transactions
					ListOfTransactions[txnObject.TransactionID] = txnObject
					//adding the new transaction id for list of txns for FIID and Product
					(ListOfTransactionsForFI[fiOrder.FIID])[fiOrder.Product] = append((ListOfTransactionsForFI[fiOrder.FIID])[fiOrder.Product], txnObject.TransactionID)
					fmt.Printf("-----ListOfTransactionsForFI : %v\n", ListOfTransactionsForFI[fiOrder.FIID])
				}
			}
		}
	}
	// set the maps
	// TODO - Add error handling
	_, err = setAllTradeObjects(stub)
	if err != nil {
		// remove the value from AllTradeObjects
		fmt.Printf("Error to set All Trade Objects : %v\n", err)
		return nil, errors.New("Failed to set AllTradeObjects")
	}
	
	_, err = setTradeSettlementMap(stub)
	if err != nil {
		// remove the value from TradeSettlementMap
		fmt.Printf("Error to set map TradeSettlement : %v\n", err)
		return nil, errors.New("Failed to set TradeSettlementMap")
	}
	
	_, err = setListOfTransactions(stub)
	if err != nil {
		// remove the value from ListOfTransactions
		// remove the value from ListOfTransactions
		fmt.Printf("Error to set map ListOfTransactions : %v\n", err)
		return nil, errors.New("Failed to set ListOfTransactions")
	}
	
	_, err = setListOfTransactionsForFI(stub)
	if err != nil {
		// remove the value from ListOfTransactionsForFI
		// remove the value from ListOfTransactionsForFI
		fmt.Printf("Error to set map ListOfTransactionsForFI : %v\n", err)
		return nil, errors.New("Failed to set ListOfTransactionsForFI")
	}
	
	return nil, err
}

/*
	Returns the list of FIOrders for a FI based on status
*/
func getAllOrdersForFIMap(stub shim.ChaincodeStubInterface) ([]FIOrder, error) {
	var fiOrders []FIOrder

	if len(AllFIOrders) > 0 {
		for _, order := range AllFIOrders {
			// get details of each FI Orders
			fmt.Printf("fiOrders are : %v\n", order)
			fiOrders = append(fiOrders, order)
		}
		fmt.Printf("List Of Orders by FI : %v \n", fiOrders)
		return fiOrders, nil
	}
	return nil, errors.New("Unable to find any orders for FI")

}

/*
	Returns the list of FIOrders for a FI based on status
*/
func getAllOrdersForFIBasedOnStatus(FIID string, Status string, stub shim.ChaincodeStubInterface) ([]FIOrder, error) {
	var fiOrderIDs []string
	var fiOrder FIOrder
	var ok bool
	var fiOrdersByStatus []FIOrder
	var oStatus orderStatusValue

	if Status == "orderPlaced" {
		oStatus = 1
	} else if Status == "orderConfirmed" {
		oStatus = 2
	}

	if fiOrderIDs, ok = AllOrdersForFI[FIID]; ok {
		fmt.Printf("fiOrders : %v\n", fiOrderIDs)
		for _, id := range fiOrderIDs {
			// get details of each FI Orders
			fmt.Printf("fiOrders ids : %v\n", id)
			if fiOrder, ok = AllFIOrders[id]; ok {
				if len(Status) > 0 {
					if fiOrder.Status == oStatus {
						fiOrdersByStatus = append(fiOrdersByStatus, fiOrder)
					}
				} else {
					fiOrdersByStatus = append(fiOrdersByStatus, fiOrder)
				}
			}
		}
		fmt.Printf("List Of Orders by FI %s : %v \n", FIID, fiOrdersByStatus)
		return fiOrdersByStatus, nil
	}
	return nil, errors.New("Unable to find any orders for FI")

}

/*
	Returns the list of FIOrders for a Broker
*/
func getAllOrdersForBrokerBasedOnStatus(BrokerID string, Status string, stub shim.ChaincodeStubInterface) ([]FIOrder, error) {
	var fiOrderIDs []string
	var fiOrdersByStatus []FIOrder
	var fiOrder FIOrder
	var ok bool
	var oStatus orderStatusValue

	if Status == "orderPlaced" {
		oStatus = 1
	} else if Status == "orderConfirmed" {
		oStatus = 2
	}
	if fiOrderIDs, ok = AllOrdersForBroker[BrokerID]; ok {
		fmt.Printf("fiOrders : %v\n", fiOrderIDs)
		for _, id := range fiOrderIDs {
			// get details of each FI Orders
			fmt.Printf("fiOrders ids : %v\n", id)
			if fiOrder, ok = AllFIOrders[id]; ok {
				if len(Status) > 0 {
					if fiOrder.Status == oStatus {
						fiOrdersByStatus = append(fiOrdersByStatus, fiOrder)
					}
				} else {
					fiOrdersByStatus = append(fiOrdersByStatus, fiOrder)
				}
			}
		}
		fmt.Printf("List Of Orders by FI %s : %v \n", BrokerID, fiOrdersByStatus)
		return fiOrdersByStatus, nil
	}
	return nil, errors.New("Unable to find any orders for Broker")
}

/*
	Returns the list of Trade Orders for Stock Exchange
*/
func getAllOrdersForSEBasedOnStatus(Status string, stub shim.ChaincodeStubInterface) ([]FIOrder, error) {
		
	//var tradeIDs TradeObject
	var fiOrdersByStatus []FIOrder
	var confOrder []string
	var fiOrderId string
	var fiAllOrder FIOrder
	var ok bool
	var oStatus tradeStatusValue
	fmt.Printf("Trade Status is %s\n",Status);
	if Status == "toBeSettled" {
		oStatus = 1
	}
	fmt.Printf("oStatus is %s\n",oStatus);
	for _, id := range AllTradeObjects {
		fmt.Printf("Inside the loop %++v\n ",id);
		if id.SettlementStatus == oStatus {
			if confOrder, ok = TradeSettlementMap[id.TradeObjectID]; ok {
				for _, fiid := range confOrder {
					if fiOrderId, ok = ConfirmedToFIOrder[fiid]; ok{
						if fiAllOrder, ok = AllFIOrders[fiOrderId]; ok{
							fiOrdersByStatus = append(fiOrdersByStatus, fiAllOrder)
						}
						
					}
				}
			}
		}		
	}
	fmt.Printf("AllOrdersForSEBasedOnStatus are %v\n ",fiOrdersByStatus);
	return fiOrdersByStatus, nil
	return nil, errors.New("Unable to find any orders for SE")
}


/*
	Returns the list of FIOrders for Custodian
*/

func getAllOrdersForCustodianBasedOnStatus(CustodianBankID string, stub shim.ChaincodeStubInterface) ([]FIOrder, error) {
	var cbOrdersByID []FIOrder

	if len(AllFIOrders) > 0 {
		for _, cbOrder := range AllFIOrders {
			// get details of each FI Orders
			fmt.Printf("fiOrders : %v\n", cbOrder)
			if cbOrder.CustodianBankID == CustodianBankID{
				cbOrdersByID = append(cbOrdersByID, cbOrder)	
			}	
		}
		fmt.Printf("List Of Orders by CUSTODIAN %s : %v \n", CustodianBankID, cbOrdersByID)
		return cbOrdersByID, nil
	}
	return nil, errors.New("Unable to find any orders for custodian")
}


/*
	Returns the list of Holdings for a FI
*/
func getAllHoldingsForFI(FIID string, stub shim.ChaincodeStubInterface) (map[string]int, error) {
	var transactions map[string][]string
	var fiHoldingsList map[string]int
	fiHoldingsList = make(map[string]int)
	var ok bool
	
	if transactions, ok = ListOfTransactionsForFI[FIID]; ok {
		fmt.Printf("**********List of transactions for FI %s : %v\n", FIID, transactions)
		for key, value := range transactions {
			// get details of each FI transactions
			l :=  len(value)
			fiHoldingsList[key] = ListOfTransactions[string(value[l-1])].EffectiveBalance
		}
		fmt.Printf("**********List Of Current Holdings by FI %s : %v \n", FIID, fiHoldingsList)
		return fiHoldingsList, nil
	}
	return nil, errors.New("Unable to find any holding for FI")
}

/*
	Returns the list of Holdings for a FI
*/
func getAllTransactionsForFIBasedOnStockId(FIID string, StockID string, stub shim.ChaincodeStubInterface) ([]Transaction, error) {
	var transactions map[string][]string
	var txArray []string
	var tx Transaction
	var fiTxList []Transaction
	var ok bool

	if transactions, ok = ListOfTransactionsForFI[FIID]; ok {
		fmt.Printf("-->Transactions : %v\n", transactions)
		txArray = transactions[StockID]
		for _, id := range txArray {
			// get details of each FI Transaction
			fmt.Printf("-->Transactions ids : %v\n", id)
			if tx, ok = ListOfTransactions[id]; ok {
				fiTxList = append(fiTxList, tx)
			}
		}
		fmt.Printf("-->List Of Holdings by FI %s : %v \n", FIID, fiTxList)
		return fiTxList, nil
	}
	return nil, errors.New("Unable to find any transaction for FI")
}

/*
getAllTradesForSettlement
*/

func getAllTradesForSettlement(stub shim.ChaincodeStubInterface) ([]string, error) {
	var allTrades []string

	// get the maps
	_, err = getAllTradeObjects(stub)
	if err != nil {
		fmt.Printf("Error to read map AllTradeObjects : %v\n", err)
		return nil, errors.New("Failed to create Trade Object")
	}

	for key, value := range AllTradeObjects {
		if value.SettlementStatus == toBeSettled {
			allTrades = append(allTrades, key)
		}
	}

	return allTrades, err
}

// Query function
func (t *CapitalMarketChainCode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var allOrders []FIOrder
	var err error
	var allBytes []byte
	var allHoldings map[string]int
	var allTransactions []Transaction
	var allFIOrders []FIOrder
	if function == "getAllOrdersForFIBasedOnStatus" {
		if len(args) != 2 {
			fmt.Printf("Incorrect number of arguments to call getAllOrdersForFIBasedOnStatus.\n")
			return nil, errors.New("Incorrect number of arguments")
		}
		allOrders, err = getAllOrdersForFIBasedOnStatus(args[0], args[1], stub)
		if err != nil {
			fmt.Printf("Error getting All Orders for FI %s : %v\n", args[0], err)
			return nil, err
		}
		allBytes, err := json.Marshal(&allOrders)
		if err != nil {
			fmt.Printf("Error unmarshalling all orders : %v\n", err)
			return nil, err
		}
		fmt.Printf("All orders for FI %s successfully read\n", args[0])
		return allBytes, nil
	} else if function == "getAllOrdersForBrokerBasedOnStatus" {
		if len(args) != 2 {
			fmt.Printf("Incorrect number of arguments.\n")
			return nil, errors.New("Incorrect number of arguments to call getAllOrdersForBrokerBasedOnStatus ")
		}
		allOrders, err = getAllOrdersForBrokerBasedOnStatus(args[0], args[1], stub)
		if err != nil {
			fmt.Printf("Error getting All Orders for Broker %s : %v", args[0], err)
			return nil, err
		}
		allBytes, err = json.Marshal(&allOrders)
		if err != nil {
			fmt.Printf("Error unmarshalling all orders : %v\n", err)
			return nil, err
		}
		fmt.Printf("All orders for Broker %s successfully read\n", args[0])
		return allBytes, nil

	} else if function == "getAllOrdersForSEBasedOnStatus" {
		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments.\n")
			return nil, errors.New("Incorrect number of arguments to call getAllOrdersForSEBasedOnStatus ")
		}
		allOrders, err = getAllOrdersForSEBasedOnStatus(args[0], stub)
		if err != nil {
			fmt.Printf("Error getting All Orders for SE %s : %v", args[0], err)
			return nil, err
		}
		allBytes, err = json.Marshal(&allOrders)
		if err != nil {
			fmt.Printf("Error unmarshalling all orders : %v\n", err)
			return nil, err
		}
		fmt.Printf("All orders for SE %s successfully read\n", args[0])
		return allBytes, nil

	} else if function == "getAllOrdersForCustodianBasedOnStatus" {
		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments.\n")
			return nil, errors.New("Incorrect number of arguments to call getAllOrdersForCustodianBasedOnStatus ")
		}
		allOrders, err = getAllOrdersForCustodianBasedOnStatus(args[0], stub)
		if err != nil {
			fmt.Printf("Error getting All Orders for Custodian %s : %v", args[0], err)
			return nil, err
		}
		allBytes, err = json.Marshal(&allOrders)
		if err != nil {
			fmt.Printf("Error unmarshalling all orders : %v\n", err)
			return nil, err
		}
		fmt.Printf("All orders for Custodian %s successfully read\n", args[0])
		return allBytes, nil

	} else if function == "getAllHoldingsForFI" {
		if len(args) != 1 {
			fmt.Printf("Incorrect number of arguments.\n")
			return nil, errors.New("Incorrect number of arguments to call getAllHoldingsForFI ")
		}
		allHoldings, err = getAllHoldingsForFI(args[0], stub)
		if err != nil {
			fmt.Printf("Error getting All holdings for FI %s : %v", args[0], err)
			return nil, err
		}
		allBytes, err = json.Marshal(&allHoldings)
		if err != nil {
			fmt.Printf("Error unmarshalling all holdings : %v\n", err)
			return nil, err
		}
		fmt.Printf("All holdings for FI %s successfully read\n", args[0])
		return allBytes, nil

	} else if function == "getAllTransactionsForFIBasedOnStockId" {
		if len(args) != 2 {
			fmt.Printf("Incorrect number of arguments.\n")
			return nil, errors.New("Incorrect number of arguments to call getAllTransactionsForFIBasedOnStockId ")
		}
		allTransactions, err = getAllTransactionsForFIBasedOnStockId(args[0], args[1], stub)
		if err != nil {
			fmt.Printf("Error getting All Transactions for FI %s : %v", args[0], err)
			return nil, err
		}
		allBytes, err = json.Marshal(&allTransactions)
		if err != nil {
			fmt.Printf("Error unmarshalling all transactions : %v\n", err)
			return nil, err
		}
		fmt.Printf("All transactions for FI %s successfully read\n", args[0])
		return allBytes, nil

	} else if function == "getAllOrdersForFIMap" {
		if len(args) != 0 {
			fmt.Printf("Incorrect number of arguments.\n")
			return nil, errors.New("Incorrect number of arguments to call getAllOrdersForFIMap ")
		}
		allFIOrders, err = getAllOrdersForFIMap(stub)
		if err != nil {
			fmt.Printf("Error getting All FI orders %v", err)
			return nil, err
		}
		allBytes, err = json.Marshal(&allFIOrders)
		if err != nil {
			fmt.Printf("Error unmarshalling all FI orders : %v\n", err)
			return nil, err
		}
		fmt.Printf("All FI orders successfully read\n")
		return allBytes, nil

	} //else {
	fmt.Println("received unknown function call: ", function)
	//}
	return nil, nil
}

// Invoke function
func (t *CapitalMarketChainCode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("Invoke running. Function: " + function)
	fmt.Printf("args: %s\n", args)

	if function == "createOrdersByFI" {
		return t.createOrdersByFI(stub, args)
	} else if function == "processFIOrdersForConfirmationBySE" {
		return t.processFIOrdersForConfirmationBySE(stub, args)
	} else if function == "clearAndSettleTrades" {
		return t.clearAndSettleTrades(stub, args)
	} else if function == "setAllInitialTransactions" {
		return t.setAllInitialTransactions(stub, args)
	}

	return nil, errors.New("Received unknown function invocation: " + function)
}

func main() {
	err := shim.Start(new(CapitalMarketChainCode))
	if err != nil {
		fmt.Printf("Error starting Capital Market chaincode: %s\n", err)
	}
}
