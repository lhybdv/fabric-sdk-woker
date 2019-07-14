package blockchain

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

// InvokeHello
func (util *FabricUtil) Invoke(from, to, value string) string {

	// Prepare arguments
	var args []string
	args = append(args, "invoke")
	args = append(args, from)
	args = append(args, to)
	args = append(args, value)

	//eventID := "eventInvoke"
	// Add data that will be visible in the proposal, like a description of the invoke request
	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in invoke")

	//util.ClearErr()
	//reg, notifier, err := util.event.RegisterChaincodeEvent(util.ChainCodeID, eventID)
	//if err != nil {
	//	util.Err = err
	//	return ""
	//}
	//defer util.event.Unregister(reg)

	// Create a request (proposal) and send it
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
	req := channel.Request{ChaincodeID: util.ChainCodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2]), []byte(args[3])}, TransientMap: transientDataMap}
	response, err := util.client.Execute(req, channel.WithTargets(allPeers...))
	if err != nil {
		util.Err = fmt.Errorf("failed to move funds: %v", err)
		return ""
	}

	if response.TransactionID == "" {
		util.Err = errors.New("TransactionID is empty")
		return ""
	}

	fmt.Println(response.TransactionID)

	//Wait for the result of the submission
	//select {
	//case ccEvent := <-notifier:
	//	fmt.Printf("Received CC event: %v\n", ccEvent)
	//case <-time.After(time.Second * 40):
	//	util.Err = fmt.Errorf("did NOT receive CC event for eventId(%s)", eventID)
	//	return ""
	//}

	return "Transfer success"
}

