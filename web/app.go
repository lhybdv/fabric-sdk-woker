package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/trias-lab/fabric-sdk-demo/blockchain"
	"net/http"
	"regexp"
)

type OrgChainCodeInfo struct {
	Org string
	Name string
	Version string
}

func Serve(util *blockchain.FabricUtil) {
	r := gin.Default()

	r.Delims("{[{","}]}")
	r.LoadHTMLGlob("web/templates/*")

	//r.Static("/", "web/templates/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/chaincode/QueryInstantiated", func(c *gin.Context){
		util.ClearErr()
		orgCCInfos := util.QueryInstantiatedChainCode(blockchain.GetOrgs()...)
		var repOrgCCInfos []OrgChainCodeInfo
		for _, orgCCInfo := range orgCCInfos {
			for _, ccInfo := range orgCCInfo.ChainCodeInfos {
				repOrgCCInfos = append(repOrgCCInfos, OrgChainCodeInfo{Org: orgCCInfo.Org, Name: ccInfo.Name, Version: ccInfo.Version})
			}
		}

		c.JSON(http.StatusOK, repOrgCCInfos)
	})

	r.GET("/chaincode/QueryInstalled", func(c *gin.Context){
		util.ClearErr()
		orgCCInfos := util.QueryInstalledChainCode(blockchain.GetOrgs()...)
		var repOrgCCInfos []OrgChainCodeInfo
		for _, orgCCInfo := range orgCCInfos {
			for _, ccInfo := range orgCCInfo.ChainCodeInfos {
				repOrgCCInfos = append(repOrgCCInfos, OrgChainCodeInfo{Org: orgCCInfo.Org, Name: ccInfo.Name, Version: ccInfo.Version})
			}
		}

		c.JSON(http.StatusOK, repOrgCCInfos)
	})

	r.GET("/tx/query", func(c *gin.Context){
		util.ClearErr()
		account := c.Query("account")
		rep := util.Query(account)
		if util.Err != nil {
			c.JSON(http.StatusOK, fmt.Sprintf("query failed: %v", util.Err))
		} else if rep == "" {
			c.JSON(http.StatusOK, "found nothing")
		} else {
			c.JSON(http.StatusOK, rep)
		}
	})

	r.GET("/tx/create", func(c *gin.Context){
		util.ClearErr()
		account := c.Query("account")
		amount := c.Query("amount")
		rep := util.Create(account, amount)
		if util.Err != nil {
			c.JSON(http.StatusOK, fmt.Sprintf("create failed: %v", util.Err))
		} else {
			c.JSON(http.StatusOK, rep)
		}
	})

	r.GET("/tx/transfer", func(c *gin.Context){
		util.ClearErr()
		from := c.Query("fromAccount")
		to := c.Query("toAccount")
		amount := c.Query("transferAmount")
		rep := util.Invoke(from, to, amount)
		if util.Err != nil {
			c.JSON(http.StatusOK, fmt.Sprintf("query failed: %v", util.Err))
		} else if rep == "" {
			c.JSON(http.StatusOK, "found nothing")
		} else {
			c.JSON(http.StatusOK, rep)
		}
	})

	r.GET("/tx/queryById", func(c *gin.Context){
		util.ClearErr()
		txID := c.Query("txId")
		tx := util.QueryTransaction(txID)
		if util.Err != nil {
			c.JSON(http.StatusOK, fmt.Sprintf("query failed: %v", util.Err))
		} else {
			c.JSON(http.StatusOK, tx)
		}
	})


	r.POST("/policy/upgrade", func(c *gin.Context) {
		str := c.PostForm("nodes")
		reg := regexp.MustCompile(`'([\d|.]+)'`)
		captures := reg.FindAllStringSubmatch(str, -1)

		var ips []string
		for _, capture := range captures {
			ips = append(ips, capture[1])
		}

		util.ClearErr()
		lastVerNum := util.GetLatestVersion()
		lastVerNum++
		version := fmt.Sprintf("%d.0", lastVerNum)

		policyStr, orgs := blockchain.GetPolicyByIPs(ips...)

		util.InstallCC(version, orgs...)
		util.UpgradeCC(version, policyStr)
		if util.Err != nil {
			c.JSON(200, gin.H{
				"status": "failed",
				"error": fmt.Sprintf("Upgrad failed: %v", util.Err),
			})
			return
		}

		blockchain.CurrentOrgs = orgs

		c.JSON(200, gin.H {
			"status": "success",
			"version": version,
			"policy": policyStr,
		})
	})

	panic(r.Run("127.0.0.1:12345"))
}
