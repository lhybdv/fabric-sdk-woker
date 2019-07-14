package blockchain

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

type Peer struct {
	IP        string
	Org       string
	Policy    string
	OrdererID string
}

type PeersConfig struct {
	AnchorOrg string
	Peers     []*Peer
}

var (
	peerCfg     PeersConfig
	AnchorOrg   string
	PeerMap     = map[string]*Peer{}
	CurrentOrgs = []string{"org1"}
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile("/Users/liuhuiyu/.gvm/pkgsets/go1.12.5/global/src/github.com/trias-lab/fabric-sdk-demo/peers.yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Errorf("Read peers.yaml failed: %v", err)
	}

	if err := viper.Unmarshal(&peerCfg); err != nil {
		fmt.Errorf("Unmarshal peers.yaml failed: %v", err)
	}

	AnchorOrg = peerCfg.AnchorOrg
	for _, peer := range peerCfg.Peers {
		PeerMap[peer.IP] = peer
	}
}

func GetOrgs() []string {
	var orgs []string
	for _, peer := range PeerMap {
		orgs = append(orgs, peer.Org)
	}
	return orgs
}

func GetAnchorOrg() string {
	return peerCfg.AnchorOrg
}

func GetPolicyByIPs(ips ...string) (string, []string) {
	var policies []string
	var orgs []string
	for _, ip := range ips {
		peer := PeerMap[ip]
		policies = append(policies, peer.Policy)
		orgs = append(orgs, peer.Org)
	}

	num := len(ips) * 2 / 3
	if len(ips)%3 > 0 {
		num++
	}
	policyStr := fmt.Sprintf("OutOf(%d, '%s')", num, strings.Join(policies, "','"))
	//policyStr := fmt.Sprintf("AND('%s')", strings.Join(policies, "','"))
	return policyStr, orgs
}
