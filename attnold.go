package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"

    "bufio"
    "encoding/base64"

    "image"
    "image/jpeg"
    "strings"
    "io/ioutil"    
    "log"
    "os"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()
 	endata := ImgtoBase64(args[1])

	err := stub.PutState(args[0], []byte(endata))

	// We store the key and the value on the ledger
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}



func ImgtoBase64(img string) string {

    f, _ := os.Open(img)

    // Read entire JPG into byte slice.
    reader := bufio.NewReader(f)
    content, _ := ioutil.ReadAll(reader)

    // Encode as base64.
    encoded := base64.StdEncoding.EncodeToString(content)

    // Print encoded data to console.
    // ... The base64 image can be used as a data URI in a browser.
    fmt.Println("ENCODED: " + encoded)
    return encoded
}

func base64toJpg(data string, key_str string) {

    reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
    m, formatString, err := image.Decode(reader)
    if err != nil {
        log.Fatal(err)
    }
    bounds := m.Bounds()
    fmt.Println("base64toJpg", bounds, formatString)

    //Encode from image format to writer
    pngFilename := "images/"+key_str+".jpg"
    f, err := os.OpenFile(pngFilename, os.O_WRONLY|os.O_CREATE, 0777)
    if err != nil {
        log.Fatal(err)
        return
    }

    err = jpeg.Encode(f, m, &jpeg.Options{Quality: 75})
    if err != nil {
        log.Fatal(err)
        return
    }
    fmt.Println("Jpg file", pngFilename, "created")

}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

        endata := ImgtoBase64(args[1])

	err := stub.PutState(args[0], []byte(endata))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return endata, nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])

	base64toJpg(string(value),args[0])

	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
 
