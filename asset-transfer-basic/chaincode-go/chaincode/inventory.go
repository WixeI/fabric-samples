package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ⭐ Data Structures ⭐

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// AgencyMBSPassthrough represents a pool of Agency Mortgage-Backed Securities (MBS) passthrough.
type AgencyMBSPassthrough struct {
	UID          string `json:"uid"`
	Bond         string `json:"bond"`         // Bond represents the bond associated with the MBS pool.
	Cusip        string `json:"cusip"`        // Cusip represents the CUSIP number of the MBS pool.
	OriginalFace int    `json:"originalFace"` // The amount of the bond
	OwnerHash    string `json:"ownerHash"`    // Owner of the Bond
	Class1       string `json:"class1"`       // Class1 represents the first class associated with the MBS pool.
}

// The private bond values of an Organization
type PrivateBond struct {
	UID          string  `json:"uid"`
	ReservePrice float64 `json:"reservePrice"`
}

// The direct trade objects.
type DirectTrade struct {
	DirectTradeID string    `json:"directTradeID"`
	Cusip         string    `json:"cusip"`
	OriginalFace  int       `json:"originalFace"`
	BidPrice      float64   `json:"bidPrice"`
	BidderHash    string    `json:"BidderHash"`
	State         string    `json:"state"` //"Open" or "Closed"
	Answers       []Answer  `json:"answers"`
	CreatedAt     time.Time `json:"createdAt"`
}

// AnswerResponse represents the response value and timestamp for an answer.
type AnswerResponse struct {
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// Answer for Direct Trade
type Answer struct {
	SellerIDHash   string         `json:"sellerIDHash"`
	SellerResponse AnswerResponse `json:"sellerResponse"`
	BuyerResponse  AnswerResponse `json:"buyerResponse"`
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
	DirectTrades []DirectTrade          `json:"directTrades"`
	Transactions []Transaction          `json:"transactions"`
}

// ⭐ Functions ⭐

// CreateBondPublic creates a new bond and adds it to the ledger as a public bond
func (s *SmartContract) CreateBondPublic(ctx contractapi.TransactionContextInterface, uid, ownerHash, bondID, cusip, class1 string, originalFace int) (string, error) {
	// Generating UID for bond. This part should be done manually and inputed in the args. In the front-end, you can manage this properly
	// uid := generateUID()
	//TODO: Add validation for uid
	//TODO: Add validation for ownerHash. Maybe it's possible to identify who ran the function while still getting the endorsers to work properly.

	// Generating OwnerHash. This part should be done manually and inputed in the args. In the front-end, you can manage this properly
	// ownerHash, err := s.GenerateOrgHash(ctx)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to generate encryption key: %v", err)
	// }

	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return "", err
	}

	// Storing bond in ledger
	bond := AgencyMBSPassthrough{
		UID:          uid,
		Bond:         bondID,
		Cusip:        cusip,
		OriginalFace: originalFace,
		OwnerHash:    ownerHash,
		Class1:       class1,
	}
	ledger.Bonds = append(ledger.Bonds, bond)
	err = s.updateLedger(ctx, ledger)
	if err != nil {
		return "", fmt.Errorf("failed to store bond: %v", err)
	}

	return uid, nil
}

// CreateBondPrivate stores the bond in the private collection with the specified UID and reserve price
func (s *SmartContract) CreateBondPrivate(ctx contractapi.TransactionContextInterface, uid string, reservePrice float64) error {
	// Storing bond in private collection
	privateBond := PrivateBond{
		UID:          uid,
		ReservePrice: reservePrice,
	}
	err := s.storePrivateBond(ctx, privateBond)
	if err != nil {
		return fmt.Errorf("failed to store private bond: %v", err)
	}

	return nil
}

// CheckDirectTrades checks if there are any open direct trades for a given cusip
func (s *SmartContract) CheckDirectTrades(ctx contractapi.TransactionContextInterface, cusip string) ([]DirectTrade, error) {
	var trades []DirectTrade

	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return nil, err
	}

	for _, trade := range ledger.DirectTrades {
		if trade.Cusip == cusip && trade.State == "Open" {
			trades = append(trades, trade)
		}
	}

	return trades, nil
}

// CloseDirectTrade closes a direct trade by DirectTradeID if the caller is the owner
func (s *SmartContract) CloseDirectTrade(ctx contractapi.TransactionContextInterface, tradeID string) error {
	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return err
	}

	for i, trade := range ledger.DirectTrades {
		if trade.DirectTradeID == tradeID {
			if s.IsOwner(ctx, trade.BidderHash) {
				ledger.DirectTrades[i].State = "Closed"
				return s.updateLedger(ctx, ledger)
			}
			return fmt.Errorf("you are not the owner of the trade")
		}
	}

	return fmt.Errorf("direct trade not found")
}

// GenerateTransactionObject creates a new Transaction object
func (s *SmartContract) GenerateTransactionObject(buyerID, sellerID, cusip string, originalFace int, boughtPrice string, timestamp time.Time) Transaction {
	return Transaction{
		BuyerID:      buyerID,
		SellerID:     sellerID,
		Cusip:        cusip,
		OriginalFace: originalFace,
		BoughtPrice:  boughtPrice,
		Timestamp:    timestamp,
	}
}

// GenerateOrgHash retrieves and returns the value of the private collection "encryption_key"
func (s *SmartContract) GenerateOrgHash(ctx contractapi.TransactionContextInterface) (string, error) {
	encryptionKey, err := s.getEncryptionKey(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %v", err)
	}

	return encryptionKey, nil
}

// IsOwner checks if the caller is the owner by comparing with the encryption key
func (s *SmartContract) IsOwner(ctx contractapi.TransactionContextInterface, ownerHash string) bool {
	encryptionKey, err := s.getEncryptionKey(ctx)
	if err != nil {
		return false
	}

	return ownerHash == encryptionKey
}

// GetBond returns all bonds from the ledger that have the given cusip and their corresponding private bonds
func (s *SmartContract) GetBond(ctx contractapi.TransactionContextInterface, cusip string) ([]struct {
	Public  AgencyMBSPassthrough
	Private PrivateBond
}, error) {
	var result []struct {
		Public  AgencyMBSPassthrough
		Private PrivateBond
	}

	// Retrieve bonds from ledger
	bonds, err := s.getAllBonds(ctx)
	if err != nil {
		return nil, err
	}

	for _, bond := range bonds {
		if bond.Cusip == cusip {
			// Get corresponding private bond
			privateBond, err := s.getPrivateBond(ctx, bond.UID)
			if err != nil {
				return nil, err
			}

			result = append(result, struct {
				Public  AgencyMBSPassthrough
				Private PrivateBond
			}{
				Public:  bond,
				Private: privateBond,
			})
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("could not find any bonds with specified Cusip: %v", cusip)
	}

	return result, nil
}

// GetAllBonds returns all bonds from the ledger
func (s *SmartContract) GetAllBonds(ctx contractapi.TransactionContextInterface) ([]AgencyMBSPassthrough, error) {
	return s.getAllBonds(ctx)
}

// GetAllTransactions returns all transactions from the ledger
func (s *SmartContract) GetAllTransactions(ctx contractapi.TransactionContextInterface) ([]Transaction, error) {
	return s.getAllTransactions(ctx)
}

// GetAllYourBonds returns all bonds from the ledger that the caller is the owner of,
// along with their corresponding private bonds.
func (s *SmartContract) GetAllYourBonds(ctx contractapi.TransactionContextInterface) ([][]interface{}, error) {
	// Get bidder hash
	bidderHash, err := s.GenerateOrgHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bidder hash: %v", err)
	}

	// Retrieve all bonds from the ledger
	allBonds, err := s.getAllBonds(ctx)
	if err != nil {
		return nil, err
	}

	var yourBonds [][]interface{}

	// Iterate through all bonds
	for _, bond := range allBonds {
		// Check if the bond owner is the caller
		if bond.OwnerHash == bidderHash {
			// Retrieve the corresponding private bond
			privateBond, err := s.getPrivateBond(ctx, bond.UID)
			if err != nil {
				return nil, err
			}

			// Append the bond and its corresponding private bond to the result
			yourBonds = append(yourBonds, []interface{}{bond, privateBond})
		}
	}

	return yourBonds, nil
}

// GetYourDirectTrades returns all direct trades where the caller is the owner
func (s *SmartContract) GetYourDirectTrades(ctx contractapi.TransactionContextInterface) ([]DirectTrade, error) {
	// Get bidder hash
	bidderHash, err := s.GenerateOrgHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bidder hash: %v", err)
	}

	// Retrieve ledger
	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return nil, err
	}

	// Filter direct trades where the caller is the owner
	var yourTrades []DirectTrade
	for _, trade := range ledger.DirectTrades {
		if trade.BidderHash == bidderHash {
			yourTrades = append(yourTrades, trade)
		}
	}

	return yourTrades, nil
}

// This is temporary. In the future, it should be an actual encryption procedure. SetEncryptionKey stores the MSPID of the organization invoking the function in the private collection
func (s *SmartContract) SetEncryptionKey(ctx contractapi.TransactionContextInterface) error {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	err = ctx.GetStub().PutPrivateData("_implicit_org_"+mspID, "encryption_key", []byte(mspID))
	if err != nil {
		return fmt.Errorf("failed to store encryption key: %v", err)
	}

	return nil
}

func (s *SmartContract) GetLedger(ctx contractapi.TransactionContextInterface) (*Ledger, error) {
	ledgerBytes, err := ctx.GetStub().GetState("ledger")
	if err != nil {
		return nil, fmt.Errorf("failed to read ledger from world state: %v", err)
	}
	if ledgerBytes == nil {
		return &Ledger{
			Bonds:        []AgencyMBSPassthrough{},
			DirectTrades: []DirectTrade{},
			Transactions: []Transaction{},
		}, nil
	}

	var ledger Ledger
	err = json.Unmarshal(ledgerBytes, &ledger)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ledger: %v", err)
	}

	return &ledger, nil
}

// CreateTrade initiates a new direct trade
func (s *SmartContract) CreateTrade(ctx contractapi.TransactionContextInterface, directTradeID, bidderHash, cusip, createdAtString string, originalFace int, bidPrice float64) (string, error) {
	// Generating UID for direct trade. This part should be done manually and inputed in the args. In the front-end, you can manage this properly
	// directTradeID := generateUID()
	// TODO: Add validation here.

	// Define the layout of the time string
	layout := "2006-01-02T15:04:05Z"

	// Parse the time string into a time.Time type
	parsedTime, err := time.Parse(layout, createdAtString)
	if err != nil {
		return "", fmt.Errorf("error parsing time: %v", err)
	}

	// Generating BidderHash
	// bidderHash, err := s.GenerateOrgHash(ctx)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to generate bidder hash: %v", err)
	// }
	// TODO: see if it's possible to get the mspid of the one executing the chaincode, but still get the endorsers to work properly

	// Creating new direct trade object
	trade := DirectTrade{
		DirectTradeID: directTradeID,
		Cusip:         cusip,
		OriginalFace:  originalFace,
		BidPrice:      bidPrice,
		BidderHash:    bidderHash,
		State:         "Open",
		Answers:       []Answer{},
		CreatedAt:     parsedTime,
	}

	// Storing direct trade in ledger
	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return "", err
	}
	ledger.DirectTrades = append(ledger.DirectTrades, trade)
	err = s.updateLedger(ctx, ledger)
	if err != nil {
		return "", fmt.Errorf("failed to store direct trade: %v", err)
	}

	return directTradeID, nil
}

// AnswerTrade updates the answer for a direct trade
func (s *SmartContract) AnswerTrade(ctx contractapi.TransactionContextInterface, directTradeID, sellerIDHash, answerValue string, timestamp time.Time) error {
	// Retrieve ledger
	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return err
	}

	// Find the direct trade
	var foundTrade *DirectTrade
	for i, trade := range ledger.DirectTrades {
		if trade.DirectTradeID == directTradeID {
			foundTrade = &ledger.DirectTrades[i]
			break
		}
	}
	if foundTrade == nil {
		return fmt.Errorf("direct trade not found")
	}

	// Find or create answer object
	var foundAnswer *Answer
	for i, ans := range foundTrade.Answers {
		if ans.SellerIDHash == sellerIDHash {
			foundAnswer = &foundTrade.Answers[i]
			break
		}
	}
	if foundAnswer == nil {
		// Create new answer object
		newAnswer := Answer{
			SellerIDHash: sellerIDHash,
			SellerResponse: AnswerResponse{
				Value:     "",
				Timestamp: time.Time{},
			},
			BuyerResponse: AnswerResponse{
				Value:     "",
				Timestamp: time.Time{},
			},
		}
		foundTrade.Answers = append(foundTrade.Answers, newAnswer)
		foundAnswer = &foundTrade.Answers[len(foundTrade.Answers)-1]
	}

	// Update SellerResponse
	foundAnswer.SellerResponse.Value = answerValue
	foundAnswer.SellerResponse.Timestamp = timestamp

	// Clear BuyerResponse
	foundAnswer.BuyerResponse.Value = ""
	foundAnswer.BuyerResponse.Timestamp = time.Time{}

	// Update ledger
	err = s.updateLedger(ctx, ledger)
	if err != nil {
		return fmt.Errorf("failed to update ledger: %v", err)
	}

	return nil
}

func (s *SmartContract) AnswerTradeAsOwner(ctx contractapi.TransactionContextInterface, directTradeID, sellerIDHash, answerValue string, timestamp time.Time) error {

	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return err
	}

	var foundTrade *DirectTrade
	for i, trade := range ledger.DirectTrades {
		if trade.DirectTradeID == directTradeID {
			foundTrade = &ledger.DirectTrades[i]
			break
		}
	}
	if foundTrade == nil {
		return fmt.Errorf("direct trade not found")
	}
	if foundTrade.State != "Open" {
		return fmt.Errorf("direct trade is closed")
	}
	// Compare MSP ID with BidderHash
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	if foundTrade.BidderHash != mspID {
		return fmt.Errorf("you are not the owner of the trade")
	}

	// Find or create answer object
	var foundAnswer *Answer
	for i, ans := range foundTrade.Answers {
		if ans.SellerIDHash == sellerIDHash {
			foundAnswer = &foundTrade.Answers[i]
			break
		}
	}
	if foundAnswer == nil {
		// Create new answer object
		newAnswer := Answer{
			SellerIDHash: sellerIDHash,
			SellerResponse: AnswerResponse{
				Value:     "",
				Timestamp: time.Time{},
			},
			BuyerResponse: AnswerResponse{
				Value:     "",
				Timestamp: time.Time{},
			},
		}
		foundTrade.Answers = append(foundTrade.Answers, newAnswer)
		foundAnswer = &foundTrade.Answers[len(foundTrade.Answers)-1]
	}

	// Update BuyerResponse
	foundAnswer.BuyerResponse.Value = answerValue
	foundAnswer.BuyerResponse.Timestamp = timestamp

	// Update ledger
	// err = s.updateLedger(ctx, ledger)
	// if err != nil {
	// 	return fmt.Errorf("failed to update ledger: %v", err)
	// }

	// Check if both SellerResponse and BuyerResponse are "yes"
	if foundAnswer.SellerResponse.Value == "yes" && foundAnswer.BuyerResponse.Value == "yes" {
		// Get bond from ledger
		bonds, err := s.getAllBonds(ctx)
		if err != nil {
			return fmt.Errorf("failed to get bond: %v", err)
		}

		// Find the bond owned by the caller
		var ownedBond *AgencyMBSPassthrough
		for _, bond := range bonds {
			if bond.OwnerHash == sellerIDHash {
				ownedBond = &bond
				break
			}
		}
		if ownedBond == nil {
			return fmt.Errorf("the seller does not own any bonds for this trade")
		}

		// Update bond owner
		for i, bond := range ledger.Bonds {
			if bond.OwnerHash == sellerIDHash {
				ownedBond = &bond
				ledger.Bonds[i].OwnerHash = foundTrade.BidderHash
				break
			}
		}

		// Close the Trade
		foundTrade.State = "Closed"

		// Generate transaction
		transaction := s.GenerateTransactionObject(foundTrade.BidderHash, foundAnswer.SellerIDHash, foundTrade.Cusip, foundTrade.OriginalFace, fmt.Sprintf("%.2f", foundTrade.BidPrice), timestamp)

		// Add transaction to ledger
		ledger.Transactions = append(ledger.Transactions, transaction)
	}

	// Update ledger
	err = s.updateLedger(ctx, ledger)
	if err != nil {
		return fmt.Errorf("failed to update ledger: %v", err)
	}

	return nil
}

// CreateTransaction generates a new transaction and adds it to the ledger
func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, buyerID, sellerID, cusip string, originalFace int, boughtPrice float64, timestamp time.Time) error {
	// Create transaction object
	transaction := Transaction{
		BuyerID:      buyerID,
		SellerID:     sellerID,
		Cusip:        cusip,
		OriginalFace: originalFace,
		BoughtPrice:  fmt.Sprintf("%.2f", boughtPrice),
		Timestamp:    timestamp,
	}

	// Retrieve ledger
	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return err
	}

	// Add transaction to ledger
	ledger.Transactions = append(ledger.Transactions, transaction)

	// Update ledger
	err = s.updateLedger(ctx, ledger)
	if err != nil {
		return fmt.Errorf("failed to update ledger: %v", err)
	}

	return nil
}

// ⭐ Helper functions for accessing ledger and private collection ⭐

func (s *SmartContract) updateLedger(ctx contractapi.TransactionContextInterface, ledger *Ledger) error {
	ledgerBytes, err := json.Marshal(ledger)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger: %v", err)
	}

	err = ctx.GetStub().PutState("ledger", ledgerBytes)
	if err != nil {
		return fmt.Errorf("failed to update ledger: %v", err)
	}

	return nil
}

func (s *SmartContract) getEncryptionKey(ctx contractapi.TransactionContextInterface) (string, error) {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get MSP ID: %v", err)
	}

	privateCollectionBytes, err := ctx.GetStub().GetPrivateData("_implicit_org_"+mspID, "encryption_key")
	if err != nil {
		return "", fmt.Errorf("_implicit_org_%s - failed to get encryption key: %v", mspID, err)
	}
	if privateCollectionBytes == nil {
		return "", fmt.Errorf("_implicit_org_%s - encryption key not found", mspID)
	}

	encryptionKey := string(privateCollectionBytes)
	return encryptionKey, nil
}

func (s *SmartContract) storePrivateBond(ctx contractapi.TransactionContextInterface, privateBond PrivateBond) error {
	// Fetching existing private bonds
	privateBonds, err := s.getPrivateBonds(ctx)
	if err != nil {
		return err
	}

	// Adding new private bond
	privateBonds = append(privateBonds, privateBond)

	// Storing updated private bonds
	privateBondsBytes, err := json.Marshal(privateBonds)
	if err != nil {
		return fmt.Errorf("failed to marshal private bonds: %v", err)
	}

	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	err = ctx.GetStub().PutPrivateData("_implicit_org_"+mspID, "private_bonds_information", privateBondsBytes)
	if err != nil {
		return fmt.Errorf("_implicit_org_%s - failed to update private bonds: %v", mspID, err)
	}

	return nil
}

func (s *SmartContract) getPrivateBond(ctx contractapi.TransactionContextInterface, uid string) (PrivateBond, error) {
	privateBonds, err := s.getPrivateBonds(ctx)
	if err != nil {
		return PrivateBond{}, err
	}

	for _, bond := range privateBonds {
		if bond.UID == uid {
			return bond, nil
		}
	}

	return PrivateBond{}, fmt.Errorf("private bond with UID %s not found", uid)
}

func (s *SmartContract) getPrivateBonds(ctx contractapi.TransactionContextInterface) ([]PrivateBond, error) {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, fmt.Errorf("failed to get MSP ID: %v", err)
	}

	privateBondsBytes, err := ctx.GetStub().GetPrivateData("_implicit_org_"+mspID, "private_bonds_information")
	if err != nil {
		return nil, fmt.Errorf("_implicit_org_%s - failed to get private bonds: %v", mspID, err)
	}
	if privateBondsBytes == nil {
		return []PrivateBond{}, nil
	}

	var privateBonds []PrivateBond
	err = json.Unmarshal(privateBondsBytes, &privateBonds)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private bonds: %v", err)
	}

	return privateBonds, nil
}

func (s *SmartContract) getAllBonds(ctx contractapi.TransactionContextInterface) ([]AgencyMBSPassthrough, error) {
	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return nil, err
	}

	return ledger.Bonds, nil
}

func (s *SmartContract) getAllTransactions(ctx contractapi.TransactionContextInterface) ([]Transaction, error) {
	ledger, err := s.GetLedger(ctx)
	if err != nil {
		return nil, err
	}

	return ledger.Transactions, nil
}

func generateUID() string {
	return uuid.New().String()
}
