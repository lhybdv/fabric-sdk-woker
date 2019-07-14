package blockchain

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// QueryHello query the chaincode to get the state of hello
func (util *FabricUtil) Query(name string) string {
	if util.Err != nil {
		return ""
	}

	var args []string
	args = append(args, "query")
	args = append(args, name)

	req := channel.Request{ChaincodeID: util.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}}

	peers, err := util.admins[CurrentOrgs[0]].local.LocalDiscoveryService().GetPeers()
	if err != nil {
		util.Err = err
		return ""
	}
	response, err := util.client.Query(req, channel.WithTargets(peers[0]))
	if err != nil {
		util.Err = fmt.Errorf("failed to query: %v", err)
		return ""
	}

	return string(response.Payload)
}

func (util *FabricUtil) Create(name, amount string) string {
	if util.Err != nil {
		return ""
	}

	var args []string
	args = append(args, "create")
	args = append(args, name)
	args = append(args, amount)

	var allPeers []fab.Peer
	for _, admin := range util.admins {
		if !contains(CurrentOrgs, admin.org) {
			continue
		}
		peers, err := admin.local.LocalDiscoveryService().GetPeers()
		if err != nil {
			util.Err = err
			return ""
		}
		allPeers = append(allPeers, peers[0])
	}
	req := channel.Request{
		ChaincodeID: util.ChainCodeID,
		Fcn:         args[0],
		Args:        [][]byte{[]byte(args[1]), []byte(args[2])},
	}
	response, err := util.client.Execute(req, channel.WithTargets(allPeers...))
	if err != nil {
		util.Err = fmt.Errorf("failed to create: %v", err)
		return ""
	}

	return string(response.Payload)
}
