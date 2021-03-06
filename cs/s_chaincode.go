// s_chaincode
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {
}

var scrutinIndexStr = "_scrutinindex" //name for the key/value that will store a list of all known marbles
var openScrutinStr = "_openscrutins"  //name for the key/value that will store all open trades
//var voteIndexStr = "_voteindex"       //name for the key/value that will store all votes

type Scrutin struct {
	Name        string  `json:"name"` //the fieldtags are needed to keep case from bouncing around
	Description string  `json:"description"`
	User        string  `json:"user"`
	Votes       []AVote `json:"votes"`
}

type AnOpenScrutin struct {
	Name      string `json:"name"`      //the fieldtags are needed to keep case from bouncing around
	User      string `json:"user"`      //user who created the open trade order
	Timestamp int64  `json:"timestamp"` //utc timestamp of creation
}

type AllScrutinViews struct {
	OpenScrutins []AnOpenScrutin `json:"open_scrutins"`
}

type AVote struct {
	Name      string   `json:"name"`      //the fieldtags are needed to keep case from bouncing around
	Users     []string `json:"users"`     //user who created the open trade order
	Timestamp int64    `json:"timestamp"` //utc timestamp of creation
	Count     int      `json:"count"`
}

/*type AllVotes struct {
	Votes []AVote `json:"votes"`
}*/

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

	/*var votes AllVotes
	jsonAsBytes, _ = json.Marshal(votes) //clear the votes struct
	err = stub.PutState(voteIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}*/

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
	} else if function == "open_scrutin" { //create a new trade order
		return t.open_scrutin(stub, args)
	} else if function == "init_vote" { //create a new marble
		return t.init_vote(stub, args)
	} else if function == "add_vote" { //create a new marble
		return t.add_vote(stub, args)
	} /*else if function == "perform_view" { //forfill an open trade order
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
// Init Scrutin - create a new scrutin, store into chaincode state, update indexes
// ============================================================================================================================
func (t *SimpleChaincode) init_scrutin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	// "nameSccrutin", "descriptionScrutin", "User"
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	//input sanitation
	fmt.Println("- start init scrutin")
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
	var votes []AVote

	//check if scrutin already exists
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
	res.Name = name
	res.Description = description
	res.Votes = votes
	res.User = user

	jsonsAsBytes, _ := json.Marshal(res)
	err = stub.PutState(name, jsonsAsBytes) //rewrite the marble with id as key
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

// ============================================================================================================================
// Init Vote - create a new vote option for a given scrutin, store into chaincode state, update scrutin
// ============================================================================================================================
func (t *SimpleChaincode) init_vote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	// "nameScrutin", "nameVote"
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	//input sanitation
	fmt.Println("- start init vote")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}

	nameScrutin := args[0]
	nameVote := args[1]

	//check if scrutin already exists
	voteAsBytes, err := stub.GetState(nameVote)
	if err != nil {
		return nil, errors.New("Failed to get vote name")
	}
	res := AVote{}
	json.Unmarshal(voteAsBytes, &res)
	if res.Name == nameVote {
		fmt.Println("This Vote arleady exists: " + nameVote)
		fmt.Println(res)
		return nil, errors.New("This vote arleady exists") //all stop a marble by this name exists
	}

	var users []string

	res.Name = nameVote
	res.Users = users
	res.Timestamp = makeTimestamp() //use timestamp as an ID
	res.Count = 0

	jsonsAsBytes, _ := json.Marshal(res)
	err = stub.PutState(nameVote, jsonsAsBytes) //rewrite the marble with id as key
	if err != nil {
		return nil, err
	}
	fmt.Println("Vote added")

	//Get the scrutin and add the vote option
	scrutinAsBytes, err := stub.GetState(nameScrutin)
	if err != nil {
		return nil, errors.New("Failed to get scrutin name")
	}
	scrutin := Scrutin{}
	json.Unmarshal(scrutinAsBytes, &scrutin)
	if scrutin.Name == nameScrutin {
		//Update scrutin by adding vote option
		scrutin.Votes = append(scrutin.Votes, res)
		scrutinUAsBytes, _ := json.Marshal(scrutin)
		err = stub.PutState(nameScrutin, scrutinUAsBytes) //store name of marble
		fmt.Println("scrutin updated")

	}
	fmt.Println("- end init vote")
	return nil, nil

}

// ============================================================================================================================
// Open Trade - create an open trade for a marble you want with marbles you have
// ============================================================================================================================
func (t *SimpleChaincode) open_scrutin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//	0        1      2     3      4      5       6
	//["bob", "blue", "16", "red", "16"] *"blue", "35*
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting like 3")
	}

	open := AnOpenScrutin{}
	open.Name = args[0]
	open.User = args[1]
	open.Timestamp = makeTimestamp() //use timestamp as an ID
	fmt.Println("- start open trade")
	jsonAsBytes, _ := json.Marshal(open)
	err = stub.PutState("_debug1", jsonAsBytes)

	//get the open trade struct
	opensAsBytes, err := stub.GetState(openScrutinStr)
	if err != nil {
		return nil, errors.New("Failed to get openscrutin")
	}
	var views = AllScrutinViews{}
	json.Unmarshal(opensAsBytes, &views) //un stringify it aka JSON.parse()

	views.OpenScrutins = append(views.OpenScrutins, open) //append to open trades
	fmt.Println("! appended open to trades")
	jsonAsBytes, _ = json.Marshal(views)
	err = stub.PutState(openScrutinStr, jsonAsBytes) //rewrite open orders
	if err != nil {
		return nil, err
	}
	fmt.Println("- end open trade")
	return nil, nil
}

// ============================================================================================================================
// Add vote - Add user at a vote and count++
// ============================================================================================================================
func (t *SimpleChaincode) add_vote(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	// "nameVote", "nameUser"
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	//input sanitation
	fmt.Println("- start init vote")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}

	nameVote := args[0]
	nameUser := args[1]

	voteAsBytes, err := stub.GetState(nameVote)
	if err != nil {
		return nil, errors.New("Failed to get vote name")
	}
	vote := AVote{}
	json.Unmarshal(voteAsBytes, &vote)
	if vote.Name == nameVote {
		vote.Users = append(vote.Users, nameUser)
		vote.Count = vote.Count + 1
		voteUAsBytes, _ := json.Marshal(vote)
		err = stub.PutState(nameVote, voteUAsBytes) //store name of marble
		fmt.Println("vote updated")
	}
	fmt.Println("- end updated vote")
	return nil, nil
}
func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}
