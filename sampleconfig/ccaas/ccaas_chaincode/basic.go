package main

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type basic struct {
	contractapi.Contract
}

func (b *basic) Save(ctx contractapi.TransactionContextInterface, key, value string) error {
	return ctx.GetStub().PutState(key, []byte(value))
}

func (b *basic) Query(ctx contractapi.TransactionContextInterface, key string) (string, error) {
	valueBytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", err
	}
	return string(valueBytes), nil
}

func main() {
	ccId := os.Getenv("CHAINCODE_ID")
	address := os.Getenv("CHAINCODE_SERVER_ADDRESS")

	chaincode, err := contractapi.NewChaincode(&basic{})
	if err != nil {
		panic(err)
	}
	s := &shim.ChaincodeServer{
		CCID:     ccId,
		Address:  address,
		CC:       chaincode,
		TLSProps: getTlsProperties(),
	}
	if err := s.Start(); err != nil {
		panic(err)
	}
}

func getTlsProperties() shim.TLSProperties {
	tlsDisabledStr := getEnvOrDefault("CHAINCODE_TLS_DISABLED", "true")
	key := getEnvOrDefault("CHAINCODE_TLS_KEY", "")
	cert := getEnvOrDefault("CHAINCODE_TLS_CERT", "")
	clientCACert := getEnvOrDefault("CHAINCODE_CLIENT_CA_CERT", "")

	// convert tlsDisabledStr to boolean
	tlsDisabled := getBoolOrDefault(tlsDisabledStr, false)
	var keyBytes, certBytes, clientCACertBytes []byte
	var err error

	if !tlsDisabled {
		keyBytes, err = ioutil.ReadFile(key)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
		certBytes, err = ioutil.ReadFile(cert)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
	}
	// Did not request for the peer cert verification
	if clientCACert != "" {
		clientCACertBytes, err = ioutil.ReadFile(clientCACert)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
	}

	return shim.TLSProperties{
		Disabled:      tlsDisabled,
		Key:           keyBytes,
		Cert:          certBytes,
		ClientCACerts: clientCACertBytes,
	}
}

func getEnvOrDefault(env, defaultVal string) string {
	value, ok := os.LookupEnv(env)
	if !ok {
		value = defaultVal
	}
	return value
}

// Note that the method returns default value if the string
// cannot be parsed!
func getBoolOrDefault(value string, defaultVal bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultVal
	}
	return parsed
}
