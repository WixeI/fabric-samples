package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Data Structures

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// AgencyMBSPassthrough represents a pool of Agency Mortgage-Backed Securities (MBS) passthrough.
type AgencyMBSPassthrough struct {
	UID                            string  `json:"uid"`                            
	Bond                            string  `json:"bond"`                            // Bond represents the bond associated with the MBS pool.
	Cusip                           string  `json:"cusip"`                           // Cusip represents the CUSIP number of the MBS pool.
	OriginalFace                    int     `json:"originalFace"`                    // The amount of the bond
	OwnerHash                       string  `json:"ownerHash"`                       // Owner of the Bond
	Class1                          string  `json:"class1"`                          // Class1 represents the first class associated with the MBS pool.
}

// The private bond values of an Organization
type PrivateBond struct {
	UID        string  `json:"uid"`
	ReservePrice float64 `json:"reservePrice"`
}

// The direct trade objects.
type DirectTrade struct {
	DirectTradeID string   `json:"directTradeID"`
	Cusip         string   `json:"cusip"`
	OriginalFace  int      `json:"originalFace"`
	BidPrice      float64   `json:"bidPrice"`
	BidderHash    string   `json:"BidderHash"`
	State         string   `json:"state"` //"Open" or "Closed"
	Answers       []Answer `json:"answers"`
	CreatedAt       time.Tiem `json:"createdAt"`

}

// Answer for Direct Trade
type Answer struct {
	SellerIDHash string `json:"sellerIDHash"`

	SellerResponse struct {
		Value     string    `json:"value"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"sellerResponse"`

	BuyerResponse struct {
		Value     string    `json:"value"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"buyerResponse"`
}

// Trade Record
type Transaction struct {
	BuyerID      string    `json:"buyerID"`
	SellerID     string    `json:"sellerID"`
	Cusip        string    `json:"cusip"`
	OriginalFace int       `json:"originalFace"`
	BoughtPrice  string    `json:"boughtPrice"`
	Timestamp    time.Time `json:"timestamp"`
}

// The Open Ledger
type Ledger struct {
	Bonds        []AgencyMBSPassthrough `json:"bonds"`
	DirectTrades []DirectTrade `json:"directTrades"`
	Transactions []Transaction `json:"transactions"`
}

Those are the data structures of this hyperledger fabric project chaincode. The objective is to be able to create bonds, make trades, and etc.

//Functions

Functions:
 - CheckDirectTrades
 - CloseDirectTrade
 - AnswerDirectTrade
 - CreateBond
 - CreateTrade

 - GenerateTransactionObject
 - GenerateOrgHash
 - GenerateEncryptionKey
 - IsOwner

 - GetYourDirectTrades
 - GetAllBonds
 - GetBond
 - GetAllTransactions

The function GenerateEncryptionKey will populate the private collection key "encryption_key" with an encryption key that will be used in the function GenerateOrgHash. The private collection is the implicit private collection that fabric generates for every organization. It can be accessed by writing code in the same format as shown below:
mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, fmt.Errorf("failed to get MSP ID: %v", err)
	}

	privateCollectionBytes, err := ctx.GetStub().GetPrivateData("_implicit_org_"+mspID, "key_here")
	if err != nil {
		return nil, fmt.Errorf("_implicit_org_"+mspID+" - failed to get inventory: %v", err)
	}
	if privateCollectionBytes == nil {
		return nil, nil
	}

The private collection will also have a key "private_bonds_information", which has a list of PrivateBond type.

The function GenerateOrgHash is about encryption. To create the hash, it should use the MSPID of the company, the EncryptionKey they have in their PrivateCollection, and the value of the Timestamp property of the Bond. 

The function CreateBond will take a struct as an argument. This struct should contain the Bond, Cusip, Class1 and OriginalFace values to populate those fields. The OwnerHash should be generated on the spot, by using the function GenerateOrgHash. The uid should be generated on the spot as well by any way you see fit. There should be transient data (anonymous to the ledger) for ReservePrice. The CreateBond function will also create a new item in the private collection key "private_bonds_information", fill the uid field with the same value we generated and just added to the bond, and fill the ReservePrice field with the information got from the transient data. When all of this is done, there should be a new PrivateBond item in the list of the private collection, and a new item in the ledger, in the Bonds list.

There should be a function called IsOwner. It will receive a string as an argument. This string is the value of a Hash. You should get the encryption key from your private collection "encryption_key", and use it in a way that gets you an MSPID from the hash received. If you don't, you're not the Owner.

There should be a GetBond function, that takes a cusip string as argument and returns all bonds from the Bonds list of the ledger that have that cusip. For every bond in the list you got from the ledger, you get the UID and check if it's the same as any of your PrivateBond items of your private collection list. If it is, then you return the data of that PrivateBond as well. The end result should be an array, of arrays, that contain a first item being the public bond from the ledger, and the second item being the PrivateBond with the same UID. If there is no private bond, the second item is left empty.

There should be a GetAllBonds function, that gets all bonds from the ledger

There should be a GetAllTransactions function, that gets all transactions from the ledger

Now, for the most important part: the ability to start trades. Trades work the following way: one company, wanting to buy a bond they saw on the Bonds list in the ledger, will run the function CreateTrade and put a cusip as argument. A new item in the DirectTrades list will be created, of type DirectTrade. Another argument of the function is the OriginalFace, that is gonna fill the field with the same name of the DirectTrade struct. The BidPrice should also be an argument and behave in the same way. The State field should be "Open". The answers field should be an empty array. The DirectTradeID should be generated on the spot. The BidderHash should use the GenerateOrgHash function with the mspid of the org creating the trade and the same timestamp value that's going to fill the CreatedAt field (generate for now).

There should be a function called GetYourDirectTrades, that returns a list of all DirectTrades that you are the owner of

There should be a function called CloseDirectTrade, that receives a DirectTradeID string. It changes the status of a DirectTrade object to "Closed", but only if you are the owner of that DirectTrade object (check by the BidderHash)

-- answers section

There should be a function called AnswerDirectTrade. It receives an argument "answer", that can either be "yes", "no" or "counter". A second argument, being a struct of type {BidPrice string}, can be left empty. Even before doing something with these arguments, you need to validate if the person invoking the AnswerDirectTrade has any bonds with the cusip that the DirectTrade demands. It does this by going through all items in the Bonds list of the ledger, and checking if the OwnerHash field matches the one of the company invoking. If it identifies you have bonds with that cusip, you can proceed. If your answer was "yes", you will verify 