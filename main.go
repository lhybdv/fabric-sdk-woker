package main

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/trias-lab/fabric-sdk-demo/blockchain"
	"github.com/trias-lab/fabric-sdk-demo/web"
	"os"
	"text/tabwriter"
)

var util *blockchain.FabricUtil

func init() {
	// Definition of the Fabric SDK properties
	util = &blockchain.FabricUtil{
		// Network parameters
		OrdererID: "orderer0.example.com",

		// Channel parameters
		ChannelID:     "mychannel",
		ChannelConfig: os.Getenv("GOPATH") + "/src/github.com/trias-lab/fabric-sdk-demo/fixtures/artifacts/org1.anchors.tx",

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
		panic(fmt.Errorf("Initialize failed for Organization [%s]: %v\n", util.AnchorOrg, util.Err))
	}
}

func main() {
	web.Serve(util)
}

func installAndInstantiateCC(version string) {
	util.ClearErr()
	util.InstallAndInstantiateCC(version)
	if  util.Err != nil {
		fmt.Printf("Unable to install and instantiate the chaincode: %v\n", util.Err)
		return
	}

	//webServe()
}

//func upgradeCC(version string) {
//	util.ClearErr()
//	util.InstallCC(version)
//	util.UpgradeCC(version)
//	err := setups[0].UpgradeCC(version)
//	if err != nil {
//		fmt.Printf("Unable to upgrade the chaincode: %v\n", err)
//	}
//}

func queryAndInvoke() {
	util.ClearErr()
	response := util.Query("aaa")
	if util.Err != nil {
		fmt.Printf("Unable to query hello on the chaincode: %v\n", util.Err)
	} else {
		fmt.Printf("Response from the query hello: %s\n", response)
	}
}

func create() {
	util.ClearErr()
	rep := util.Create("lhy", "1000")
	if util.Err!= nil {
		fmt.Println(util.Err)
	}

	fmt.Println("Response: ", rep)
	response := util.Query("lhy")
	if util.Err != nil {
		fmt.Printf("Unable to query hello on the chaincode: %v\n", util.Err)
	} else {
		fmt.Printf("Response from the query hello: %s\n", response)
	}
}

func queryInstalledChainCode() {
	util.ClearErr()
	orgCCInfos := util.QueryInstalledChainCode("org1", "org2", "org3", "org4")
	if util.Err != nil {
		fmt.Printf("Query Installed ChainCode failed: %v\n", util.Err)
		return
	}
	for _, orgCCInfo := range orgCCInfos {
		fmt.Printf("----- Installed chainCode for Org [%s] -----\n", orgCCInfo.Org)
		writeCodeInfos(orgCCInfo.ChainCodeInfos)
	}
}

func queryInstantiatedChainCode() {
	util.ClearErr()
	orgCCInfos := util.QueryInstantiatedChainCode("org1", "org2", "org3", "org4")
	if util.Err != nil {
		fmt.Printf("Query Instantiated ChainCode failed: %v\n", util.Err)
		return
	}
	for _, orgCCInfo := range orgCCInfos {
		fmt.Printf("----- Instantiated chainCode for Org [%s] -----\n", orgCCInfo.Org)
		writeCodeInfos(orgCCInfo.ChainCodeInfos)
	}
}

func writeCodeInfos(codeInfos []*peer.ChaincodeInfo) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	for _, codeInfo := range codeInfos {
		fmt.Fprintf(w, "\t%s\t%s\n", codeInfo.Name, codeInfo.Version)
	}
}
