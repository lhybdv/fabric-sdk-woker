package blockchain

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFabricUtil_QueryInstalledChainCode(t *testing.T) {
	util := buildUtil()
	orgCCInfos := util.QueryInstalledChainCode(GetOrgs()...)
	for _, orgCCInfo := range orgCCInfos {
		fmt.Printf("Org: %s\n", orgCCInfo.Org)
		for _, ccInfo := range orgCCInfo.ChainCodeInfos {
			fmt.Printf("CC: %s, Version: %s", ccInfo.Name, ccInfo.Version)
		}
	}
}

func TestFabricUtil_QueryInstantiatedChainCode(t *testing.T) {
	util := buildUtil()
	orgCCInfos := util.QueryInstantiatedChainCode(GetOrgs()...)
	for _, orgCCInfo := range orgCCInfos {
		fmt.Printf("Org: %s\n", orgCCInfo.Org)
		for _, ccInfo := range orgCCInfo.ChainCodeInfos {
			fmt.Printf("CC: %s, Version: %s", ccInfo.Name, ccInfo.Version)
		}
	}
}

func TestFabricUtil_GetLatestVersion(t *testing.T) {
	util := buildUtil()
	latestVerNum := util.GetLatestVersion()
	assert.Equal(t, latestVerNum, 9)
}

type Transaction struct {
	payload string
}

func TestFabricUtil_QueryTransaction(t *testing.T) {
	util := buildUtil()
	if util.Err != nil {
		fmt.Println(util.Err)
		return
	}

	txID := "6ab9bc6e2a68a1841dc0b5ef9de8965dadfef266d01c4bb95ea3dd21c36ab01c"
	txDetail := util.QueryTransaction(txID)
	if util.Err != nil {
		fmt.Println(util.Err)
		return
	}

	fmt.Println(txDetail)
}

func buildUtil() *FabricUtil {
	util := &FabricUtil{
		// Network parameters
		OrdererID: "orderer0.example.com",

		// Channel parameters
		ChannelID:     "mychannel",
		ChannelConfig: os.Getenv("GOPATH") + "/src/github.com/trias-lab/fabric-sdk-demo/fixtures/artifacts/mychannel.tx",

		// Chaincode parameters
		ChainCodeID:     "txcc",
		ChaincodeGoPath: os.Getenv("GOPATH"),
		ChaincodePath:   "github.com/trias-lab/fabric-sdk-demo/chaincode/",
		OrgAdmin:        "Admin",
		//ConfigFile:      "config.yaml",
		ConfigFile: os.Getenv("GOPATH") + "/src/github.com/trias-lab/fabric-sdk-demo/config.yaml",

		// User parameters
		UserName: "User1",
	}

	util.Initialize()
	if util.Err != nil {
		fmt.Printf("Initialize failed for Organization [%s]: %v\n", util.AnchorOrg, util.Err)
	}

	return util
}
