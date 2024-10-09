package main

import (
    "encoding/json"
    "fmt"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// BloodDonationChaincode implements the smart contract for blood donation management
type BloodDonationChaincode struct {
    contractapi.Contract
}

// Donor structure to hold donor details
type Donor struct {
    DonorID   string `json:"donorID"`
    Name      string `json:"name"`
    BloodType string `json:"bloodType"`
}

// Acceptor structure to hold acceptor (hospital) details, including phone number
type Acceptor struct {
    AcceptorID   string `json:"acceptorID"`
    Name         string `json:"name"`
    Location     string `json:"location"`
    PhoneNumber  string `json:"phoneNumber"` // New field for phone number
}

// BloodUnit structure to hold blood donation details
type BloodUnit struct {
    UnitID      string `json:"unitID"`
    DonorID     string `json:"donorID"`
    BloodType   string `json:"bloodType"`
    Quantity    int    `json:"quantity"`
    Status      string `json:"status"`     // e.g., "Collected", "Tested", "Available", "Partially Used", "Used", "Unsafe"
    TestResult  string `json:"testResult"` // e.g., "Safe", "Unsafe"
    HospitalName string `json:"hospitalName"` // New field for hospital name
}

// Register a new donor
func (s *BloodDonationChaincode) RegisterDonor(ctx contractapi.TransactionContextInterface, donorID string, name string, bloodType string) error {
    donor := Donor{
        DonorID:   donorID,
        Name:      name,
        BloodType: bloodType,
    }
    donorBytes, err := json.Marshal(donor)
    if err != nil {
        return err
    }
    return ctx.GetStub().PutState(donorID, donorBytes)
}

// Register a new acceptor (hospital)
func (s *BloodDonationChaincode) RegisterAcceptor(ctx contractapi.TransactionContextInterface, acceptorID string, name string, location string, phoneNumber string) error {
    acceptor := Acceptor{
        AcceptorID: acceptorID,
        Name:       name,
        Location:   location,
        PhoneNumber: phoneNumber, // Add phone number to acceptor
    }
    acceptorBytes, err := json.Marshal(acceptor)
    if err != nil {
        return err
    }
    return ctx.GetStub().PutState(acceptorID, acceptorBytes)
}

// Record a blood donation
func (s *BloodDonationChaincode) RecordDonation(ctx contractapi.TransactionContextInterface, unitID string, donorID string, bloodType string, quantity int, hospitalName string) error {
    bloodUnit := BloodUnit{
        UnitID:      unitID,
        DonorID:     donorID,
        BloodType:   bloodType,
        Quantity:    quantity,
        Status:      "Collected",
        HospitalName: hospitalName, // Add hospital name to blood unit
    }
    bloodBytes, err := json.Marshal(bloodUnit)
    if err != nil {
        return err
    }
    return ctx.GetStub().PutState(unitID, bloodBytes)
}

// Test blood and update the test result and status
func (s *BloodDonationChaincode) TestBlood(ctx contractapi.TransactionContextInterface, unitID string, testResult string) error {
    bloodBytes, err := ctx.GetStub().GetState(unitID)
    if err != nil {
        return err
    }
    if bloodBytes == nil {
        return fmt.Errorf("Blood unit with ID %s does not exist", unitID)
    }

    var bloodUnit BloodUnit
    err = json.Unmarshal(bloodBytes, &bloodUnit)
    if err != nil {
        return err
    }

    // Update test result and status based on the test result
    bloodUnit.TestResult = testResult
    if testResult == "Safe" {
        bloodUnit.Status = "Tested"
    } else {
        bloodUnit.Status = "Unsafe"
    }

    updatedBloodBytes, err := json.Marshal(bloodUnit)
    if err != nil {
        return err
    }
    return ctx.GetStub().PutState(unitID, updatedBloodBytes)
}

// Query the details of a donor
func (s *BloodDonationChaincode) QueryDonor(ctx contractapi.TransactionContextInterface, donorID string) (*Donor, error) {
    donorBytes, err := ctx.GetStub().GetState(donorID)
    if err != nil {
        return nil, err
    }
    if donorBytes == nil {
        return nil, fmt.Errorf("Donor with ID %s does not exist", donorID)
    }

    var donor Donor
    err = json.Unmarshal(donorBytes, &donor)
    if err != nil {
        return nil, err
    }
    return &donor, nil
}

// Query the details of an acceptor
func (s *BloodDonationChaincode) QueryAcceptor(ctx contractapi.TransactionContextInterface, acceptorID string) (*Acceptor, error) {
    acceptorBytes, err := ctx.GetStub().GetState(acceptorID)
    if err != nil {
        return nil, err
    }
    if acceptorBytes == nil {
        return nil, fmt.Errorf("Acceptor with ID %s does not exist", acceptorID)
    }

    var acceptor Acceptor
    err = json.Unmarshal(acceptorBytes, &acceptor)
    if err != nil {
        return nil, err
    }
    return &acceptor, nil
}

// Query the details of a blood unit
func (s *BloodDonationChaincode) QueryBloodUnit(ctx contractapi.TransactionContextInterface, unitID string) (*BloodUnit, error) {
    bloodBytes, err := ctx.GetStub().GetState(unitID)
    if err != nil {
        return nil, err
    }
    if bloodBytes == nil {
        return nil, fmt.Errorf("Blood unit with ID %s does not exist", unitID)
    }

    var bloodUnit BloodUnit
    err = json.Unmarshal(bloodBytes, &bloodUnit)
    if err != nil {
        return nil, err
    }
    return &bloodUnit, nil
}

// Query blood units by blood type
func (s *BloodDonationChaincode) QueryBloodUnitsByType(ctx contractapi.TransactionContextInterface, bloodType string) ([]*BloodUnit, error) {
    queryString := fmt.Sprintf(`{"selector":{"bloodType":"%s"}}`, bloodType)

    resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var bloodUnits []*BloodUnit
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var bloodUnit BloodUnit
        err = json.Unmarshal(queryResponse.Value, &bloodUnit)
        if err != nil {
            return nil, err
        }
        bloodUnits = append(bloodUnits, &bloodUnit)
    }

    return bloodUnits, nil
}

// AcceptBlood function to update the status of a blood unit when accepted by a hospital
func (s *BloodDonationChaincode) AcceptBlood(ctx contractapi.TransactionContextInterface, unitID string, acceptorID string, quantity int) error {
    bloodBytes, err := ctx.GetStub().GetState(unitID)
    if err != nil {
        return err
    }
    if bloodBytes == nil {
        return fmt.Errorf("Blood unit with ID %s does not exist", unitID)
    }

    var bloodUnit BloodUnit
    err = json.Unmarshal(bloodBytes, &bloodUnit)
    if err != nil {
        return err
    }

    // Check if the quantity requested is available
    if bloodUnit.Quantity < quantity {
        return fmt.Errorf("Insufficient blood quantity available. Available: %d, Requested: %d", bloodUnit.Quantity, quantity)
    }

    // Update the quantity of the blood unit
    bloodUnit.Quantity -= quantity

    // Log the transaction
    transactionRecord := fmt.Sprintf("Blood unit %s accepted by %s, quantity: %d", unitID, acceptorID, quantity)
    fmt.Println(transactionRecord) // Log the acceptance of blood

    // Automatically mark the blood unit as used if quantity is zero
    if bloodUnit.Quantity == 0 {
        bloodUnit.Status = "Used"
        return s.UseBlood(ctx, unitID)
    } else {
        bloodUnit.Status = "Partially Used" // Indicate that some quantity is still available
    }

    updatedBloodBytes, err := json.Marshal(bloodUnit)
    if err != nil {
        return err
    }
    return ctx.GetStub().PutState(unitID, updatedBloodBytes)
}

// Update the status of a blood unit to "Used"
func (s *BloodDonationChaincode) UseBlood(ctx contractapi.TransactionContextInterface, unitID string) error {
    bloodBytes, err := ctx.GetStub().GetState(unitID)
    if err != nil {
        return err
    }
    if bloodBytes == nil {
        return fmt.Errorf("Blood unit with ID %s does not exist", unitID)
    }

    var bloodUnit BloodUnit
    err = json.Unmarshal(bloodBytes, &bloodUnit)
    if err != nil {
        return err
    }

    bloodUnit.Status = "Used" // Update status to used

    updatedBloodBytes, err := json.Marshal(bloodUnit)
    if err != nil {
        return err
    }
    return ctx.GetStub().PutState(unitID, updatedBloodBytes)
}

// Query the donation history for a specific donor
func (s *BloodDonationChaincode) QueryDonationHistory(ctx contractapi.TransactionContextInterface, donorID string) ([]*BloodUnit, error) {
    queryString := fmt.Sprintf(`{"selector":{"donorID":"%s"}}`, donorID)

    resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var donationHistory []*BloodUnit
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var bloodUnit BloodUnit
        err = json.Unmarshal(queryResponse.Value, &bloodUnit)
        if err != nil {
            return nil, err
        }
        donationHistory = append(donationHistory, &bloodUnit)
    }

    return donationHistory, nil
}


// Main function
func main() {
    chaincode, err := contractapi.NewChaincode(new(BloodDonationChaincode))
    if err != nil {
        panic(err)
    }
    if err := chaincode.Start(); err != nil {
        panic(err)
    }
}
