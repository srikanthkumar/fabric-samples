/*
 * The smart contract for Customer Loyalty Program
 *
 */

 package main

 /* Imports
  * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
  * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
  */
 import (
	 "bytes"
	 "encoding/json"
	 "fmt"
	 "reflect"
	 "strconv"
	 "time"
 
	 "github.com/hyperledger/fabric/core/chaincode/shim"
	 sc "github.com/hyperledger/fabric/protos/peer"
 )
 
 // Define the Smart Contract structure
 type SmartContract struct {
 }
 
 type Entity interface {
	 getId() string
 }
 
 // Defining structure, with properties.  Struct tags are used by encoding/json library
 type User struct {
	 ObjectType   string `json:"docType"`
	 Id           string `json:"id"`
	 FirstName    string `json:"firstName"`
	 LastName	 string `json:"lastName"`
	 Email        string `json:"email"`
	 PhoneNumber  string `json:"phoneNumber"`
	 UserType     string `json:"userType"`
	 DateOfRegistration string `json:"dateOfRegistration"`
 }
 
 func (x User) getId() string {
	 return x.Id
 }
 
 type Activity struct {
	 ObjectType     string `json:"docType"`
	 Id             string `json:"id"`
	 DateOfActivity string `json:"dateOfActivity"`
 }
 
 func (x Activity) getId() string {
	 return x.Id
 }
 
 // The main function is only relevant in unit test mode. Only included here for completeness.
 func main() {
	 // Create a new Smart Contract
	 err := shim.Start(new(SmartContract))
	 if err != nil {
		 fmt.Printf("Error creating new Smart Contract: %s", err)
	 }
 }
 
 /*
  * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
  * Best practice is to have any Ledger initialization in separate function -- see initLedger()
  */
 func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	 return shim.Success(nil)
 }
 
 /*
  * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
  * The calling application program has also specified the particular smart contract function to be called, with arguments
  */
 func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
 
	 // Retrieve the requested Smart Contract function and arguments
	 function, args := APIstub.GetFunctionAndParameters()
	 // Route to the appropriate handler function to interact with the ledger appropriately
	 inputs := make([]reflect.Value, len(args)+1)
	 inputs[0] = reflect.ValueOf(APIstub)
	 for i, _ := range args {
		 inputs[i+1] = reflect.ValueOf(args[i])
	 }
	 fmt.Println("#####################", inputs, function)
	 if resp, ok := (reflect.ValueOf(s).MethodByName(function).Call(inputs))[0].Interface().(sc.Response); ok {
		 return resp
	 } else {
		 return shim.Error("Invalid Smart Contract function name.")
	 }
 }
 
 //////////////////////////////  Entity   ////////////////////////////////////////////////
 
 // ==================================================================================
 // Create Entity - creates a new entity, store into chaincode state
 // ==================================================================================
 func (s *SmartContract) SaveEntity(APIstub shim.ChaincodeStubInterface, args ...string) sc.Response {
	 if len(args) != 2 {
		 return shim.Error("Incorrect number of arguments. Expecting 2")
	 }
	 var entity interface{}
	 switch args[0] {
	 case "User":
		 entity = new(User)
	 case "Activity":
		 entity = new(Activity)
	 }
	 err := json.Unmarshal([]byte(args[1]), entity)
	 aggAsBytes, err := json.Marshal(entity)
	 if err != nil {
		 return shim.Error(err.Error())
	 }
	 entityConv, _ := entity.(Entity)
	 APIstub.PutState(entityConv.getId(), aggAsBytes)
	 if err != nil {
		 return shim.Error(err.Error())
	 }
	 return shim.Success(nil)
 }
 
 // ============================================================
 // GetEntity - Read an entity from chaincode state
 // ============================================================
 func (s *SmartContract) GetEntity(APIstub shim.ChaincodeStubInterface, args ...string) sc.Response {
	 if len(args) != 1 {
		 return shim.Error("Incorrect number of arguments. Expecting 1")
	 }
	 aggAsBytes, _ := APIstub.GetState(args[0])
	 fmt.Println("Fetching record with ID " + args[0] + " Got response as "+ string(aggAsBytes[:]))
	 return shim.Success(aggAsBytes)
 }
 
 // ========================================================================
 // GetEntityByQuery - Rich Query based Entity read
 // Only available on state databases that support rich query (e.g. CouchDB)
 // ========================================================================
 func (s *SmartContract) GetEntityByQuery(APIstub shim.ChaincodeStubInterface, args ...string) sc.Response {
	 if len(args) != 1 {
		 return shim.Error("Incorrect number of arguments. Expecting 1")
	 }
	 queryString := args[0]
	 queryResults, _ := getQueryResultForQueryString(APIstub, queryString)
	 return shim.Success(queryResults)
 }
 
 func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
	 fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
	 resultsIterator, err := stub.GetQueryResult(queryString)
	 if err != nil {
		 return nil, err
	 }
	 defer resultsIterator.Close()
	 // buffer is a JSON array containing QueryRecords
	 var buffer bytes.Buffer
	 buffer.WriteString("[")
	 bArrayMemberAlreadyWritten := false
	 for resultsIterator.HasNext() {
		 queryResponse,
			 err := resultsIterator.Next()
		 if err != nil {
			 return nil, err
		 }
		 // Add a comma before array members, suppress it for the first array member
		 if bArrayMemberAlreadyWritten == true {
			 buffer.WriteString(",")
		 }
		 buffer.WriteString("{\"Key\":")
		 buffer.WriteString("\"")
		 buffer.WriteString(queryResponse.Key)
		 buffer.WriteString("\"")
		 buffer.WriteString(", \"Record\":")
		 // Record is a JSON object, so we write as-is
		 buffer.WriteString(string(queryResponse.Value))
		 buffer.WriteString("}")
		 bArrayMemberAlreadyWritten = true
	 }
	 buffer.WriteString("]")
	 fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
	 return buffer.Bytes(), nil
 }
 
 // ==================================================================================
 // GetHistoryForEntity - Gets edit-history/audit-info for a given entity
 // ==================================================================================
 func (s *SmartContract) GetHistoryForEntity(stub shim.ChaincodeStubInterface, args ...string) sc.Response {
 
	 if len(args) < 1 {
		 return shim.Error("Incorrect number of arguments. Expecting 1")
	 }
 
	 contractName := args[0]
 
	 fmt.Printf("- start getHistoryForEntity: %s\n", contractName)
 
	 resultsIterator, err := stub.GetHistoryForKey(contractName)
	 if err != nil {
		 return shim.Error(err.Error())
	 }
	 defer resultsIterator.Close()
 
	 // buffer is a JSON array containing historic values for the marble
	 var buffer bytes.Buffer
	 buffer.WriteString("[")
 
	 bArrayMemberAlreadyWritten := false
	 for resultsIterator.HasNext() {
		 response, err := resultsIterator.Next()
		 if err != nil {
			 return shim.Error(err.Error())
		 }
		 // Add a comma before array members, suppress it for the first array member
		 if bArrayMemberAlreadyWritten == true {
			 buffer.WriteString(",")
		 }
		 buffer.WriteString("{\"TxId\":")
		 buffer.WriteString("\"")
		 buffer.WriteString(response.TxId)
		 buffer.WriteString("\"")
 
		 buffer.WriteString(", \"Value\":")
 
		 if response.IsDelete {
			 buffer.WriteString("null")
		 } else {
			 buffer.WriteString(string(response.Value))
		 }
 
		 buffer.WriteString(", \"Timestamp\":")
		 buffer.WriteString("\"")
		 buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		 buffer.WriteString("\"")
 
		 buffer.WriteString(", \"IsDelete\":")
		 buffer.WriteString("\"")
		 buffer.WriteString(strconv.FormatBool(response.IsDelete))
		 buffer.WriteString("\"")
 
		 buffer.WriteString("}")
		 bArrayMemberAlreadyWritten = true
	 }
	 buffer.WriteString("]")
 
	 fmt.Printf("- getHistoryForEntityId returning:\n%s\n", buffer.String())
 
	 return shim.Success(buffer.Bytes())
 }
 
