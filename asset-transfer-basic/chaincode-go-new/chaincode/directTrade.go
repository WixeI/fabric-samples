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
	Bond                            string  `json:"bond"`                            // Bond represents the bond associated with the MBS pool.
	Cusip                           string  `json:"cusip"`                           // Cusip represents the CUSIP number of the MBS pool.
	OriginalFace                    int     `json:"originalFace"`                    // The amount of the bond
	OwnerHash                       string  `json:"ownerHash"`                       // Owner of the Bond
	Class1                          string  `json:"class1"`                          // Class1 represents the first class associated with the MBS pool.
	Class2                          string  `json:"class2"`                          // Class2 represents the second class associated with the MBS pool.
	Class3                          string  `json:"class3"`                          // Class3 represents the third class associated with the MBS pool.
	Class4                          string  `json:"class4"`                          // Class4 represents the fourth class associated with the MBS pool.
	Coupon                          float64 `json:"coupon"`                          // Coupon represents the coupon rate of the MBS pool.
	CouponType                      string  `json:"couponType"`                      // CouponType represents the type of coupon (e.g., Fixed or Floating) of the MBS pool.
	IssueYear                       int     `json:"issueYear"`                       // IssueYear represents the year of issuance of the MBS pool.
	IssueDate                       string  `json:"issueDate"`                       // IssueDate represents the date of issuance of the MBS pool.
	OriginationAmount               float64 `json:"originationAmount"`               // OriginationAmount represents the original amount of the MBS pool.
	Factor                          float64 `json:"factor"`                          // Factor represents the factor of the MBS pool.
	FactorDate                      string  `json:"factorDate"`                      // FactorDate represents the date of factor calculation of the MBS pool.
	WeightedAverageCoupon           float64 `json:"weightedAverageCoupon"`           // WeightedAverageCoupon represents the weighted average coupon of the MBS pool.
	WeightedAverageLoanAge          float64 `json:"weightedAverageLoanAge"`          // WeightedAverageLoanAge represents the weighted average loan age of the MBS pool.
	WeightedAverageMaturity         float64 `json:"weightedAverageMaturity"`         // WeightedAverageMaturity represents the weighted average maturity of the MBS pool.
	WeightedAverageOriginalMaturity float64 `json:"weightedAverageOriginalMaturity"` // WeightedAverageOriginalMaturity represents the weighted average original maturity of the MBS pool.
	LoanSize                        float64 `json:"loanSize"`                        // LoanSize represents the loan size of the MBS pool.
	LoanToValue                     float64 `json:"loanToValue"`                     // LoanToValue represents the loan-to-value ratio of the MBS pool.
	Fico                            float64 `json:"fico"`                            // Fico represents the FICO score of the MBS pool.
	Cpr1m                           float64 `json:"cpr1m"`                           // Cpr1m represents the CPR (Constant Prepayment Rate) for 1 month of the MBS pool.
	Cpr3m                           float64 `json:"cpr3m"`                           // Cpr3m represents the CPR for 3 months of the MBS pool.
	Cpr6m                           float64 `json:"cpr6m"`                           // Cpr6m represents the CPR for 6 months of the MBS pool.
	Cpr12m                          float64 `json:"cpr12m"`                          // Cpr12m represents the CPR for 12 months of the MBS pool.
	Servicer                        string  `json:"servicer"`                        // Servicer represents the servicer associated with the MBS pool.
	Geography                       string  `json:"geography"`                       // Geography represents the geographic location of the MBS pool.
	PurchasePercent                 float64 `json:"purchasePercent"`                 // PurchasePercent represents the percentage of purchases in the MBS pool.
	RefinancePercent                float64 `json:"refinancePercent"`                // RefinancePercent represents the percentage of refinances in the MBS pool.
	ThirdpartyOriginationPercent    float64 `json:"thirdpartyOriginationPercent"`    // ThirdpartyOriginationPercent represents the percentage of third-party originations in the MBS pool.
	LoanCount                       int     `json:"loanCount"`                       // LoanCount represents the number of loans in the MBS pool.
}

// The private bond values of an Organization
type PrivateBond struct {
	Cusip        string  `json:"cusip"`
	ReservePrice float64 `json:"reservePrice"`
}

// The direct trade objects.
type DirectTrade struct {
	DirectTradeID string   `json:"directTradeID"`
	Cusip         string   `json:"cusip"`
	OriginalFace  int      `json:"originalFace"`
	BidPrice      string   `json:"bidPrice"`
	BidderHash    string   `json:"BidderHash"`
	State         string   `json:"state"` //"Open" or "Closed"
	Answers       []Answer `json:"answers"`
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
	DirectTrades []AgencyMBSPassthrough `json:"directTrades"`
	Transactions []AgencyMBSPassthrough `json:"transactions"`
}

//Functions

// Initializes the ledger with bsae set of assets
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// Unmarshal JSON content from "data.go" into slice of assets
	var assets []AgencyMBSPassthrough
	err := json.Unmarshal(InitData, &assets)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Put each asset into the ledger
	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.Cusip, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

//Utils

// Returns true when bond asset with the given Cusip exists in world state
func (s *SmartContract) BondExists(ctx contractapi.TransactionContextInterface, cusip string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(cusip)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GenerateMetadata generates metadata for an asset.
func GenerateMetadata(ctx contractapi.TransactionContextInterface) (AssetMetadata, error) {
	// Get the organization ID of the peer executing the function
	orgID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return AssetMetadata{}, err
	}

	// Get the name of the organization
	orgName, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return AssetMetadata{}, err
	}

	// Get the current time
	now := time.Now()

	// Create metadata
	metadata := AssetMetadata{
		Owner:       orgName,
		OwnerId:     orgID,
		DateCreated: now,
	}

	return metadata, nil
}

//Ledger-Related

// Updates an existing bond asset in the world state with provided parameters.
func (s *SmartContract) UpdateBond(ctx contractapi.TransactionContextInterface, bondJSON string) error {
	var bond AgencyMBSPassthrough
	err := json.Unmarshal([]byte(bondJSON), &bond)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bond JSON: %v", err)
	}

	exists, err := s.BondExists(ctx, bond.Cusip)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the bond with Cusip %s does not exist", bond.Cusip)
	}

	newBondJSON, err := json.Marshal(bond)
	if err != nil {
		return fmt.Errorf("failed to marshal bond: %v", err)
	}

	return ctx.GetStub().PutState(bond.Cusip, newBondJSON)
}

// Deletes a given bond asset from the world state.
func (s *SmartContract) DeleteBond(ctx contractapi.TransactionContextInterface, cusip string) error {
	exists, err := s.BondExists(ctx, cusip)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the bond with Cusip %s does not exist", cusip)
	}

	return ctx.GetStub().DelState(cusip)
}

// Returns all bond assets found in world state
func (s *SmartContract) GetAllBonds(ctx contractapi.TransactionContextInterface) ([]*AgencyMBSPassthrough, error) {
	// Range query with empty string for startKey and endKey retrieves all bonds
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var bonds []*AgencyMBSPassthrough
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("error iterating over results: %v", err)
		}

		var bond AgencyMBSPassthrough
		err = json.Unmarshal(queryResponse.Value, &bond)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling bond JSON: %v", err)
		}
		bonds = append(bonds, &bond)
	}

	return bonds, nil
}

// GetBond fetches an AgencyMBSPassthrough from the ledger by its Cusip
func (s *SmartContract) GetBond(ctx contractapi.TransactionContextInterface, cusip string) (*AgencyMBSPassthrough, error) {
	// Retrieve the bond asset from the world state
	assetJSON, err := ctx.GetStub().GetState(cusip)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("bond with Cusip %s does not exist", cusip)
	}

	// Unmarshal the asset JSON into an AgencyMBSPassthrough object
	var bond AgencyMBSPassthrough
	err = json.Unmarshal(assetJSON, &bond)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal bond JSON: %v", err)
	}

	return &bond, nil
}

// GetBondHistoryData

// Inventory-Related

// Creates a new bond asset in the world state with given details and adds it to the organization's inventory
func (s *SmartContract) CreateBond(ctx contractapi.TransactionContextInterface, bondJSON string) error {

	var bond AgencyMBSPassthrough
	err := json.Unmarshal([]byte(bondJSON), &bond)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bond JSON: %v", err)
	}

	exists, err := s.BondExists(ctx, bond.Cusip)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the bond with Cusip %s already exists", bond.Cusip)
	}

	// Add the new bond to the world state
	newBondJSON, err := json.Marshal(bond)
	if err != nil {
		return fmt.Errorf("failed to marshal bond: %v", err)
	}
	err = ctx.GetStub().PutState(bond.Cusip, newBondJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	s.AddToInventory(ctx, bondJSON)

	return nil
}

// Creates a new bond asset in the world state with fixed details and adds it to the organization's inventory
func (s *SmartContract) CreateBondAuto(ctx contractapi.TransactionContextInterface) error {

	// Bond details to be added to the inventory
	bond := AgencyMBSPassthrough{
		Bond:                            "FR RA7777",
		Cusip:                           "Cusip123",
		Class1:                          "passthrough",
		Class2:                          "MBS 30yr",
		Class3:                          "Freddie Mac",
		Class4:                          "LB200",
		Coupon:                          6,
		CouponType:                      "FIXED",
		IssueYear:                       2023,
		IssueDate:                       "2023-01-09T12:00:00Z",
		OriginationAmount:               231480386,
		Factor:                          0.96735693,
		FactorDate:                      "2024-01-02T12:00:00Z",
		WeightedAverageCoupon:           6.895,
		WeightedAverageLoanAge:          7,
		WeightedAverageMaturity:         350,
		WeightedAverageOriginalMaturity: 359,
		LoanSize:                        189432.61,
		LoanToValue:                     74,
		Fico:                            747,
		Cpr1m:                           7.07805522352359,
		Cpr3m:                           6.72923614321762,
		Cpr6m:                           0,
		Cpr12m:                          0,
		Servicer:                        "MULTIPLE",
		Geography:                       "8.1% MI",
		PurchasePercent:                 86.11,
		RefinancePercent:                2.57,
		ThirdpartyOriginationPercent:    22.74,
		LoanCount:                       1202,
	}

	exists, err := s.BondExists(ctx, bond.Cusip)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the bond with Cusip %s already exists", bond.Cusip)
	}

	// Add the new bond to the world state
	newBondJSON, err := json.Marshal(bond)
	if err != nil {
		return fmt.Errorf("failed to marshal bond: %v", err)
	}
	err = ctx.GetStub().PutState(bond.Cusip, newBondJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	return nil
}

// GetInventory returns the inventory for the organization from the private data collection
func (s *SmartContract) GetInventory(ctx contractapi.TransactionContextInterface) (*Inventory, error) {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, fmt.Errorf("failed to get MSP ID: %v", err)
	}

	inventoryBytes, err := ctx.GetStub().GetPrivateData("_implicit_org_"+mspID, "inventory")
	if err != nil {
		return nil, fmt.Errorf("_implicit_org_"+mspID+" - failed to get inventory: %v", err)
	}
	if inventoryBytes == nil {
		return nil, nil
	}

	var inventory Inventory
	err = json.Unmarshal(inventoryBytes, &inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal inventory: %v", inventoryBytes)
	}

	return &inventory, nil
}

// Adds an AgencyMBSPassthrough item to the organization's inventory
func (s *SmartContract) AddToInventoryAuto(ctx contractapi.TransactionContextInterface) error {

	// Bond details to be added to the inventory
	bond := AgencyMBSPassthrough{
		Bond:                            "FR RA8888",
		Cusip:                           "Cusip123",
		Class1:                          "passthrough",
		Class2:                          "MBS 30yr",
		Class3:                          "Freddie Mac",
		Class4:                          "LB200",
		Coupon:                          6,
		CouponType:                      "FIXED",
		IssueYear:                       2023,
		IssueDate:                       "2023-01-09T12:00:00Z",
		OriginationAmount:               231480386,
		Factor:                          0.96735693,
		FactorDate:                      "2024-01-02T12:00:00Z",
		WeightedAverageCoupon:           6.895,
		WeightedAverageLoanAge:          7,
		WeightedAverageMaturity:         350,
		WeightedAverageOriginalMaturity: 359,
		LoanSize:                        189432.61,
		LoanToValue:                     74,
		Fico:                            747,
		Cpr1m:                           7.07805522352359,
		Cpr3m:                           6.72923614321762,
		Cpr6m:                           0,
		Cpr12m:                          0,
		Servicer:                        "MULTIPLE",
		Geography:                       "8.1% MI",
		PurchasePercent:                 86.11,
		RefinancePercent:                2.57,
		ThirdpartyOriginationPercent:    22.74,
		LoanCount:                       1202,
	}

	metadata, err := GenerateMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate metadata: %v", err)
	}

	// Get the inventory for the organization
	inventory, err := s.GetInventory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %v", err)
	}
	if inventory == nil {
		inventory = &Inventory{
			Assets: []*PrivateAgencyMBSPassthrough{},
		}
	}

	privateBond := PrivateAgencyMBSPassthrough{
		Metadata: metadata,
		Content:  &bond,
	}

	// Add the bond to the inventory
	inventory.Assets = append(inventory.Assets, &privateBond)

	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	// Marshal and put the updated inventory into the private data collection
	inventoryBytes, err := json.Marshal(inventory)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %v", err)
	}
	err = ctx.GetStub().PutPrivateData("_implicit_org_"+mspID, "inventory", inventoryBytes)
	if err != nil {
		return fmt.Errorf("failed to put inventory of %s: %v", mspID, err)
	}

	return nil
}

// Adds a fixed AgencyMBSPassthrough item to the organization's inventory
func (s *SmartContract) AddToInventory(ctx contractapi.TransactionContextInterface, bondJSON string) error {
	// Convert bondJSON string to byte slice
	bondBytes := []byte(bondJSON)

	// Unmarshal bondJSON into AgencyMBSPassthrough struct
	var bond AgencyMBSPassthrough
	err := json.Unmarshal(bondBytes, &bond)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bond JSON: %v", err)
	}

	// Get the inventory for the organization
	inventory, err := s.GetInventory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %v", err)
	}
	if inventory == nil {
		inventory = &Inventory{
			Assets: []*PrivateAgencyMBSPassthrough{},
		}
	}

	metadata, err := GenerateMetadata(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate metadata: %v", err)
	}

	privateBond := PrivateAgencyMBSPassthrough{
		Metadata: metadata,
		Content:  &bond,
	}

	// Add the bond to the inventory
	inventory.Assets = append(inventory.Assets, &privateBond)

	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	// Marshal and put the updated inventory into the private data collection
	inventoryBytes, err := json.Marshal(inventory)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %v", err)
	}
	err = ctx.GetStub().PutPrivateData("_implicit_org_"+mspID, "inventory", inventoryBytes)
	if err != nil {
		return fmt.Errorf("failed to put inventory of %s: %v", mspID, err)
	}

	return nil
}

// Adds a fixed AgencyMBSPassthrough item to the organization's inventory
func (s *SmartContract) FromInventoryToLedger(ctx contractapi.TransactionContextInterface, cusip string) error {
	// Get the inventory from the private collection
	inventory, err := s.GetInventory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %v", err)
	}

	// Check if the inventory is empty
	if inventory == nil || len(inventory.Assets) == 0 {
		return fmt.Errorf("inventory is empty")
	}

	// Find the PrivateAgencyMBSPassthrough with the given CUSIP
	var privateBond *PrivateAgencyMBSPassthrough
	for _, asset := range inventory.Assets {
		if asset.Content != nil && asset.Content.Cusip == cusip {
			privateBond = asset
			break
		}
	}

	// Check if the PrivateAgencyMBSPassthrough with the given CUSIP exists
	if privateBond == nil {
		return fmt.Errorf("private MBSPassthrough with CUSIP %s not found", cusip)
	}

	publicBond := privateBond.Content

	// Add the new bond to the world state
	publicBondJSON, err := json.Marshal(publicBond)
	if err != nil {
		return fmt.Errorf("failed to marshal bond: %v", err)
	}
	err = ctx.GetStub().PutState(publicBond.Cusip, publicBondJSON)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	return nil
}

// Removes a bond from the inventory by its CUSIP
func (s *SmartContract) RemoveFromInventory(ctx contractapi.TransactionContextInterface, cusip string) error {
	// Get the inventory for the organization
	inventory, err := s.GetInventory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %v", err)
	}
	if inventory == nil {
		return fmt.Errorf("inventory not found")
	}

	// Find the bond in the inventory by its CUSIP and remove it
	found := false
	for i, privateBond := range inventory.Assets {
		if privateBond.Content.Cusip == cusip {
			inventory.Assets = append(inventory.Assets[:i], inventory.Assets[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("bond with CUSIP %s not found in the inventory", cusip)
	}

	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	// Marshal and put the updated inventory into the private data collection
	inventoryBytes, err := json.Marshal(inventory)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %v", err)
	}
	err = ctx.GetStub().PutPrivateData("_implicit_org_"+mspID, "inventory", inventoryBytes)
	if err != nil {
		return fmt.Errorf("failed to put inventory of %s: %v", mspID, err)
	}

	return nil
}

// Edits a bond in the inventory using provided bond JSON string
func (s *SmartContract) EditBondInInventory(ctx contractapi.TransactionContextInterface, bondJSON string) error {
	// Unmarshal bondJSON directly into AgencyMBSPassthrough struct
	var bond AgencyMBSPassthrough
	err := json.Unmarshal([]byte(bondJSON), &bond)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bond JSON: %v", err)
	}

	// Get the inventory for the organization
	inventory, err := s.GetInventory(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %v", err)
	}
	if inventory == nil {
		return fmt.Errorf("inventory not found")
	}

	// Find the bond in the inventory by its CUSIP and update it
	found := false
	for i, privateBond := range inventory.Assets {
		if privateBond.Content.Cusip == bond.Cusip {
			inventory.Assets[i].Content = &bond
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("bond with CUSIP %s not found in the inventory", bond.Cusip)
	}

	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	// Marshal and put the updated inventory into the private data collection
	inventoryBytes, err := json.Marshal(inventory)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %v", err)
	}
	err = ctx.GetStub().PutPrivateData("_implicit_org_"+mspID, "inventory", inventoryBytes)
	if err != nil {
		return fmt.Errorf("failed to put inventory of %s: %v", mspID, err)
	}

	return nil
}
