package blockchain

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"strconv"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	putils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"time"
)

type fabricAdmin struct {
	org    string
	client *resmgmt.Client
	local  *context.Local
}

// FabricUtil implementation
type FabricUtil struct {
	ConfigFile      string
	OrgID           string
	OrdererID       string
	ChannelID       string
	ChainCodeID     string
	initialized     bool
	ChannelConfig   string
	ChaincodeGoPath string
	ChaincodePath   string
	OrgAdmin        string
	AnchorOrg       string
	UserName        string
	Err             error
	client          *channel.Client
	ledgerClient    *ledger.Client
	admins          map[string]*fabricAdmin
	sdk             *fabsdk.FabricSDK
	event           *event.Client
}

type OrgChainCodeInfo struct {
	Org            string
	ChainCodeInfos []*peer.ChaincodeInfo
}

type TransactionDetail struct {
	TransactionID string
	CreateTime string
	Args []string
}

type queryFn func(options ...resmgmt.RequestOption) (*peer.ChaincodeQueryResponse, error)

func (util *FabricUtil) Initialize() {
	util.InitSdk()
	util.InitAdmin(GetAnchorOrg(), GetOrgs()...)
	util.InitClientAndEvent()
	//util.SetAchors()
}

//// Initialize reads the configuration file and sets up the client, chain and event hub
//func (setup *FabricUtil) Initialize() error {
//	if setup.initialized {
//		return errors.New("sdk already initialized")
//	}
//
//	setup.InitSdk()
//	setup.InitAdmin()
//
//	// The MSP client allow us to retrieve user information from their identity, like its signing identity which we will need to save the channel
//	mspClient, err := mspclient.New(setup.sdk.Context(), mspclient.WithOrg(setup.AnchorOrg))
//	if err != nil {
//		return errors.WithMessage(err, "failed to create MSP client")
//	}
//
//	adminIdentity, err := mspClient.GetSigningIdentity(setup.OrgAdmin)
//	if err != nil {
//		return errors.WithMessage(err, "failed to get admin signing identity")
//	}
//
//	req := resmgmt.SaveChannelRequest{
//		ChannelID:         setup.ChannelID,
//		ChannelConfigPath: setup.ChannelConfig,
//		SigningIdentities: []msp.SigningIdentity{adminIdentity},
//	}
//	txID, err := setup.admin.SaveChannel(req, resmgmt.WithOrdererEndpoint(setup.OrdererID))
//	if err != nil || txID.TransactionID == "" {
//		return errors.WithMessage(err, "failed to save channel")
//	}
//	fmt.Println("Channel created")
//
//	// Make admin user join the previously created channel
//	err = setup.admin.JoinChannel(setup.ChannelID,
//		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
//		resmgmt.WithOrdererEndpoint(setup.OrdererID))
//	if err != nil {
//		return errors.WithMessage(err, "failed to make admin join channel")
//	}
//	fmt.Println("Channel joined")
//
//	fmt.Println("Initialization Successful")
//	setup.initialized = true
//	return nil
//}

func (util *FabricUtil) InstallAndInstantiateCC(version string) {
	util.InstallCC(version)
	if util.Err != nil {
		return
	}

	// Set up chaincode policy
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"org1.hf.chainhero.io"})

	instantiateCCReq := resmgmt.InstantiateCCRequest{
		Name:    util.ChainCodeID,
		Path:    util.ChaincodeGoPath,
		Version: version,
		Args:    [][]byte{[]byte("init")},
		Policy:  ccPolicy,
	}

	resp, err := util.admins[util.AnchorOrg].client.InstantiateCC(util.ChannelID, instantiateCCReq)
	if err != nil || resp.TransactionID == "" {
		util.Err = errors.WithMessage(err, "failed to instantiate the chaincode")
		return
	}
	fmt.Println("Chaincode instantiated")
}

func (util *FabricUtil) InstallCC(version string, orgs ...string) *resource.CCPackage {
	if util.Err != nil {
		return nil
	}

	ccPkg, err := packager.NewCCPackage(util.ChaincodePath, util.ChaincodeGoPath)
	if err != nil {
		util.Err = errors.WithMessage(err, "failed to create chaincode package")
		return nil
	}
	fmt.Println("ccPkg created")

	// InstallCC example cc to org peers
	installCCReq := resmgmt.InstallCCRequest{
		Name:    util.ChainCodeID,
		Path:    util.ChaincodePath,
		Version: version,
		Package: ccPkg,
	}

	for _, admin := range util.admins {
		_, err = admin.client.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
		if err != nil {
			util.Err = errors.WithMessage(err, "failed to install chaincode")
			return nil
		}
		fmt.Println("Chaincode installed")
	}
	return ccPkg
}

func (util *FabricUtil) UpgradeCC(version, policyStr string) {
	if util.Err != nil {
		return
	}

	ccPolicy, err := cauthdsl.FromString(policyStr)
	if err != nil {
		util.Err = errors.WithMessage(err, fmt.Sprintf("invalid policy string [%s]", policyStr))
		return
	}

	upgradeCCReq := resmgmt.UpgradeCCRequest{
		Name:    util.ChainCodeID,
		Path:    util.ChaincodeGoPath,
		Version: version,
		Args:    [][]byte{[]byte("init")},
		Policy:  ccPolicy,
	}

	resp, err := util.admins[util.AnchorOrg].client.UpgradeCC(util.ChannelID, upgradeCCReq)
	if err != nil || resp.TransactionID == "" {
		util.Err = errors.WithMessage(err, "failed to upgrade the chaincode")
		return
	}
	fmt.Println("Chaincode upgraded")
}

func (util *FabricUtil) CloseSDK() {
	util.sdk.Close()
}

func (util *FabricUtil) InitSdk() {
	util.admins = make(map[string]*fabricAdmin)
	// Initialize the SDK with the configuration file
	sdk, err := fabsdk.New(config.FromFile(util.ConfigFile))
	if err != nil {
		util.Err = errors.WithMessage(err, "failed to create SDK")
		return
	}
	util.sdk = sdk
	fmt.Println("SDK created")
}

func (util *FabricUtil) InitAdmin(anchorOrg string, orgs ...string) {
	if util.Err != nil {
		return
	}
	util.AnchorOrg = anchorOrg
	for _, org := range orgs {
		// The resource management client is responsible for managing channels (create/update channel)
		resMgmClientCtx := util.sdk.Context(fabsdk.WithUser(util.OrgAdmin), fabsdk.WithOrg(org))
		local, err := context.NewLocal(resMgmClientCtx)
		if err != nil {
			util.Err = errors.WithMessage(err, "failed to create local")
			return
		}

		resMgmtClient, err := resmgmt.New(resMgmClientCtx)
		if err != nil {
			util.Err = errors.WithMessage(err, "failed to create channel management client from Admin identity")
			return
		}

		admin := fabricAdmin{org: org, client: resMgmtClient, local: local}
		util.admins[org] = &admin
		fmt.Println("Ressource management client created")
	}
}

func (util *FabricUtil) InitClientAndEvent() {
	if util.Err != nil {
		return
	}
	// Channel client is used to query and execute transactions
	//peers, err := util.admins[util.AnchorOrg].local.LocalDiscoveryService().GetPeers()
	clientContext := util.sdk.ChannelContext(util.ChannelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg(CurrentOrgs[0]))
	client, err := channel.New(clientContext)
	if err != nil {
		util.Err = errors.WithMessage(err, "failed to create new channel client")
		return
	}
	util.client = client
	fmt.Println("Channel client created")

	// Creation of the client which will enables access to our channel events
	//util.event, err = event.New(clientContext)
	//if err != nil {
	//	util.Err = errors.WithMessage(err, "failed to create new event client")
	//	return
	//}
	//fmt.Println("Event client created")
	ledgerClient, err := ledger.New(clientContext)
	if err != nil {
		util.Err = errors.WithMessage(err, "failed to create new ledger client")
		return
	}

	util.ledgerClient = ledgerClient
	fmt.Println("Ledger client created")
}

func (util *FabricUtil) JoinChannel() {
	if util.Err != nil {
		return
	}
	mspClient, err := mspclient.New(util.sdk.Context(), mspclient.WithOrg(util.AnchorOrg))
	if err != nil {
		util.Err = errors.WithMessage(err, "failed to create MSP client")
		return
	}
	adminIdentity, err := mspClient.GetSigningIdentity(util.OrgAdmin)
	if err != nil {
		util.Err = errors.WithMessage(err, "failed to get admin signing identity")
		return
	}
	req := resmgmt.SaveChannelRequest{
		ChannelID:         util.ChannelID,
		ChannelConfigPath: util.ChannelConfig,
		SigningIdentities: []msp.SigningIdentity{adminIdentity},
	}
	txID, err := util.admins[util.AnchorOrg].client.SaveChannel(req, resmgmt.WithOrdererEndpoint(util.OrdererID))
	if err != nil || txID.TransactionID == "" {
		util.Err = errors.WithMessage(err, "failed to save channel")
	}

	err = util.admins[util.AnchorOrg].client.JoinChannel(util.ChannelID,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint(util.OrdererID))
	if err != nil {
		util.Err = err
	}
}

func (util *FabricUtil) SetAchors() {
	for i:= 1; i <= 3; i++ {
		org := fmt.Sprintf("org%d", i)
		mspClient, err := mspclient.New(util.sdk.Context(), mspclient.WithOrg(org))
		if err != nil {
			util.Err = errors.WithMessage(err, "failed to create MSP client")
			return
		}
		adminIdentity, err := mspClient.GetSigningIdentity(util.OrgAdmin)
		if err != nil {
			util.Err = errors.WithMessage(err, "failed to get admin signing identity")
			return
		}

		cfgPath := fmt.Sprintf("/Users/liuhuiyu/.gvm/pkgsets/go1.12.5/global/src/github.com/trias-lab/fabric-sdk-demo/fixtures/artifacts/org%d.anchors.tx", i)
		req := resmgmt.SaveChannelRequest{
			ChannelID:         util.ChannelID,
			ChannelConfigPath: cfgPath,
			SigningIdentities: []msp.SigningIdentity{adminIdentity},
		}

		ordererID := fmt.Sprintf("orderer%d.example.com", i-1)
		txID, err := util.admins[org].client.SaveChannel(req, resmgmt.WithOrdererEndpoint(ordererID))
		if err != nil || txID.TransactionID == "" {
			util.Err = errors.WithMessage(err, "failed to save channel")
		}
	}
}

func (util *FabricUtil) QueryInstalledChainCode(orgs ...string) []OrgChainCodeInfo {
	var orgCCInfos []OrgChainCodeInfo
	for _, org := range orgs {
		admin := util.admins[org]
		ccInfos := util.queryChainCode(admin.client.QueryInstalledChaincodes, admin.local)
		orgCCInfo := OrgChainCodeInfo{Org: org, ChainCodeInfos: ccInfos}
		orgCCInfos = append(orgCCInfos, orgCCInfo)
	}
	return orgCCInfos
}

func (util *FabricUtil) GetLatestVersion() int {
	versionNum := 1
	orgCCInfos := util.QueryInstantiatedChainCode(GetOrgs()...)
	for _, orgCCInfo := range orgCCInfos {
		for _, ccInfo := range orgCCInfo.ChainCodeInfos {
			if ccInfo.Name != util.ChainCodeID {
				continue
			}

			num, err := strconv.ParseFloat(ccInfo.Version, 64)
			if err != nil {
				num = 1
			}
			if int(num) > versionNum {
				versionNum = int(num)
			}
		}
	}

	return versionNum
}

func (util *FabricUtil) QueryInstantiatedChainCode(orgs ...string) []OrgChainCodeInfo {
	var orgCCInfos []OrgChainCodeInfo
	for _, org := range orgs {
		admin := util.admins[org]
		ccInfos := util.queryChainCode(func(options ...resmgmt.RequestOption) (*peer.ChaincodeQueryResponse, error) {
			return admin.client.QueryInstantiatedChaincodes(util.ChannelID, options...)
		}, admin.local)
		orgCCInfo := OrgChainCodeInfo{Org: org, ChainCodeInfos: ccInfos}
		orgCCInfos = append(orgCCInfos, orgCCInfo)
	}
	return orgCCInfos
}

func (util *FabricUtil) queryChainCode(fn queryFn, local *context.Local) []*peer.ChaincodeInfo {
	if util.Err != nil {
		return nil
	}
	peers, err := local.LocalDiscoveryService().GetPeers()
	if err != nil {
		util.Err = err
		return nil
	}

	if len(peers) == 0 {
		util.Err = errors.New("no peers found")
		return nil
	}

	rep, err := fn(resmgmt.WithTargets(peers[0]))
	if err != nil {
		util.Err = err
		return nil
	}

	return rep.Chaincodes
}

func (util *FabricUtil) QueryTransaction(txID string) *TransactionDetail {
	if util.Err != nil {
		return nil
	}

	peers, err := util.admins[util.AnchorOrg].local.LocalDiscoveryService().GetPeers()
	if err != nil {
		util.Err = err
		return nil
	}
	env, err := util.ledgerClient.QueryTransaction(fab.TransactionID(txID), ledger.WithTargets(peers[0]))
	if err != nil {
		util.Err = err
		return nil
	}

	payload, err := putils.GetPayload(env.TransactionEnvelope)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	chHeaderBytes := payload.Header.ChannelHeader
	chHeader := &cb.ChannelHeader{}
	err = proto.Unmarshal(chHeaderBytes, chHeader)
	if err != nil {
		util.Err = errors.WithMessage(err, "error extracting channel header")
		return nil
	}

	tx, err := putils.GetTransaction(payload.Data)
	if err != nil {
		util.Err = errors.WithMessage(err, "error extracing transaction")
		return nil
	}

	ccActionPayload, err := putils.GetChaincodeActionPayload(tx.Actions[0].Payload)
	if err != nil {
		util.Err = errors.WithMessage(err, "error extracting chaincode action payload")
		return nil
	}

	propPayload := &pb.ChaincodeProposalPayload{}
	err = proto.Unmarshal(ccActionPayload.ChaincodeProposalPayload, propPayload)
	if err != nil {
		util.Err = errors.WithMessage(err, "error extracting Proposal payload")
		return nil
	}

	invokeSpec := &pb.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(propPayload.Input, invokeSpec)
	if err != nil {
		util.Err = errors.WithMessage(err, "error extracting Invocation Spec")
		return nil
	}

	var args []string

	for _, v := range invokeSpec.ChaincodeSpec.Input.Args {
		args = append(args, string(v))
	}

	return &TransactionDetail{
		TransactionID: chHeader.TxId,
		Args: args,
		CreateTime: time.Unix(chHeader.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05"),
	}
}



func (util *FabricUtil) ClearErr() {
	util.Err = nil
}
