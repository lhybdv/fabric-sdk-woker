package blockchain

import (
	"fmt"
	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/msp"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetOrgs(t *testing.T) {
	expectedOrg := []string{"org1", "org2", "org3", "org4"}
	orgs := GetOrgs()
	assert.ElementsMatch(t, orgs, expectedOrg)
}

func TestGetAnchorOrg(t *testing.T) {
	expectedAhchorOrg := "org1"
	anchorOrg := GetAnchorOrg()
	assert.Equal(t, expectedAhchorOrg, anchorOrg)
}

func TestGetOrgsAndPolicyByIPs(t *testing.T) {
	ips := []string{"54.255.239.58", "13.229.49.131", "54.179.169.147", "52.77.242.99"}
	policyStr := GetPolicyByIPs(ips...)

	num := len(ips)*2/3
	if len(ips)%3 > 0 {
		num++
	}
	_, err := cauthdsl.FromString(policyStr)
	if err != nil {
		assert.Error(t, err)
	}
	expectedPolicyStr := fmt.Sprintf("OutOf (%d, 'Org1MSP.member','Org2MSP.member','Org3MSP.member','Org4MSP.member')", num)
	assert.Equal(t, policyStr, expectedPolicyStr)
}

func TestPolicy(t *testing.T) {
	p1, err := cauthdsl.FromString(GetPolicyByIPs("54.255.239.58", "13.229.49.131", "54.179.169.147", "52.77.242.99"))
	assert.NoError(t, err)

	principals := make([]*msp.MSPPrincipal, 0)

	principals = append(principals, &msp.MSPPrincipal{
		PrincipalClassification: msp.MSPPrincipal_ROLE,
		Principal:               utils.MarshalOrPanic(&msp.MSPRole{Role: msp.MSPRole_MEMBER, MspIdentifier: "Org1MSP"})})

	principals = append(principals, &msp.MSPPrincipal{
		PrincipalClassification: msp.MSPPrincipal_ROLE,
		Principal:               utils.MarshalOrPanic(&msp.MSPRole{Role: msp.MSPRole_MEMBER, MspIdentifier: "Org2MSP"})})

	principals = append(principals, &msp.MSPPrincipal{
		PrincipalClassification: msp.MSPPrincipal_ROLE,
		Principal:               utils.MarshalOrPanic(&msp.MSPRole{Role: msp.MSPRole_MEMBER, MspIdentifier: "Org3MSP"})})

	principals = append(principals, &msp.MSPPrincipal{
		PrincipalClassification: msp.MSPPrincipal_ROLE,
		Principal:               utils.MarshalOrPanic(&msp.MSPRole{Role: msp.MSPRole_MEMBER, MspIdentifier: "Org4MSP"})})

	p2 := &common.SignaturePolicyEnvelope{
		Version:    0,
		Rule:       cauthdsl.And(cauthdsl.And(cauthdsl.And(cauthdsl.SignedBy(0), cauthdsl.SignedBy(1)), cauthdsl.SignedBy(2)), cauthdsl.SignedBy(3)),
		Identities: principals,
	}

	assert.Equal(t, p1, p2)
}
