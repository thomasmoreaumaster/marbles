// s_chaincode
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {
}

var scrutinIndexStr = "_scrutinindex" //name for the key/value that will store a list of all known marbles
var openScrutinStr = "_openscrutins"  //name for the key/value that will store all open trades

type Scrutin struct {
	Name        string `json:"name"` //the fieldtags are needed to keep case from bouncing around
	Description string `json:"description"`
	User        string `json:"user"`
}

type AnOpenScrutin struct {
	User      string `json:"user"`      //user who created the open trade order
	Timestamp int64  `json:"timestamp"` //utc timestamp of creation
}

type AllScrutinViews struct {
	OpenScrutins []AnOpenScrutin `json:"open_scrutins"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval))) //making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty) //marshal an emtpy array of strings to clear the index
	err = stub.PutState(scrutinIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	var views AllScrutinViews
	jsonAsBytes, _ = json.Marshal(views) //clear the open trade struct
	err = stub.PutState(openScrutinStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" { //writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_scrutin" { //create a new marble
		return t.init_scrutin(stub, args)
	} /*else if function == "set_user" { //change owner of a marble
		res, err := t.set_user(stub, args)
		cleanScrutins(stub) //lets make sure all open trades are still valid
		return res, err
	} else if function == "open_scrutin" { //create a new trade order
		return t.open_scrutin(stub, args)
	} else if function == "perform_view" { //forfill an open trade order
		res, err := t.perform_view(stub, args)
		cleanScrutins(stub) //lets clean just in case
		return res, err
	} else if function == "remove_view" { //cancel an open trade order
		return t.remove_view(stub, args)
	}*/
	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil //send it onward
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Marble - create a new marble, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_scrutin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}

	name := args[0]
	description := strings.ToLower(args[1])
	user := strings.ToLower(args[2])

	//check if marble already exists
	scrutinAsBytes, err := stub.GetState(name)
	if err != nil {
		return nil, errors.New("Failed to get scrutin name")
	}
	res := Scrutin{}
	json.Unmarshal(scrutinAsBytes, &res)
	if res.Name == name {
		fmt.Println("This scrutin arleady exists: " + name)
		fmt.Println(res)
		return nil, errors.New("This scrutin arleady exists") //all stop a marble by this name exists
	}

	//build the marble json string manually
	str := `{"name": "` + name + `", "description": "` + description + `", "user": "` + user + `"}`
	err = stub.PutState(name, []byte(str)) //store marble with id as key
	if err != nil {
		return nil, err
	}

	//get the marble index
	scrutinsAsBytes, err := stub.GetState(scrutinIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get scrutin index")
	}
	var scrutinIndex []string
	json.Unmarshal(scrutinsAsBytes, &scrutinIndex) //un stringify it aka JSON.parse()

	//append
	scrutinIndex = append(scrutinIndex, name) //add marble name to index list
	fmt.Println("! scrutin index: ", scrutinIndex)
	jsonAsBytes, _ := json.Marshal(scrutinIndex)
	err = stub.PutState(scrutinIndexStr, jsonAsBytes) //store name of marble

	fmt.Println("- end init scrutin")
	return nil, nil
}