package dataplane

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Azure/azure-container-networking/npm/cache"
	"github.com/Azure/azure-container-networking/npm/http/api"
	NPMIPtable "github.com/Azure/azure-container-networking/npm/pkg/dataplane/iptables"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/parse"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/pb"
	"github.com/Azure/azure-container-networking/npm/util"
	"google.golang.org/protobuf/encoding/protojson"
)

// Converter struct
type Converter struct {
	ListMap        map[string]string // key: hash(value), value: one of namespace, label of namespace, multiple values
	SetMap         map[string]string // key: hash(value), value: one of label of pods, cidr, namedport
	AzureNPMChains map[string]bool
	NPMCache       *cache.NPMCache
}

// NpmCacheFromFile initialize NPM cache from file.
func (c *Converter) NpmCacheFromFile(npmCacheJSONFile string) error {
	file, err := os.Open(npmCacheJSONFile)
	if err != nil {
		return fmt.Errorf("failed to open file : %w", err)
	}

	defer file.Close()
	c.NPMCache, err = cache.Decode(bufio.NewReader(file))
	if err != nil {
		return fmt.Errorf("failed to decode npm cache due to : %w", err)
	}
	return nil
}

// NpmCache initialize NPM cache from node.
func (c *Converter) NpmCache() error {
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("http://localhost:%v%v", api.DefaultHttpPort, api.NPMMgrPath),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create http request : %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request NPM Cache : %w", err)
	}
	defer resp.Body.Close()

	c.NPMCache, err = cache.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot decode NPM Cache : %w", err)
	}

	return nil
}

// Initialize converter from file.
func (c *Converter) initConverterFile(npmCacheJSONFile string) error {
	err := c.NpmCacheFromFile(npmCacheJSONFile)
	if err != nil {
		return fmt.Errorf("error occurred during initialize converter : %w", err)
	}
	c.initConverterMaps()
	return nil
}

// Initialize converter from node.
func (c *Converter) initConverter() error {
	err := c.NpmCache()
	if err != nil {
		return fmt.Errorf("error occurred during initialize converter : %w", err)
	}
	c.initConverterMaps()

	return nil
}

// Initialize all converter's maps.
func (c *Converter) initConverterMaps() {
	c.AzureNPMChains = make(map[string]bool)
	for _, chain := range AzureNPMChains {
		c.AzureNPMChains[chain] = true
	}
	c.ListMap = make(map[string]string)
	c.SetMap = make(map[string]string)

	for k := range c.NPMCache.ListMap {
		hashedName := util.GetHashedName(k)
		c.ListMap[hashedName] = k
	}
	for k := range c.NPMCache.SetMap {
		hashedName := util.GetHashedName(k)
		c.SetMap[hashedName] = k
	}
}

// GetJSONRulesFromIptableFile returns a list of json rules from npmCache and iptable-save files.
func (c *Converter) GetJSONRulesFromIptableFile(
	tableName string,
	npmCacheFile string,
	iptableSaveFile string,
) ([][]byte, error) {

	pbRule, err := c.GetProtobufRulesFromIptableFile(tableName, npmCacheFile, iptableSaveFile)
	if err != nil {
		return nil, fmt.Errorf("error occurred during getting JSON rules from iptables : %w", err)
	}
	return c.jsonRuleList(pbRule)
}

// GetJSONRulesFromIptables returns a list of json rules from node
func (c *Converter) GetJSONRulesFromIptables(tableName string) ([][]byte, error) {
	pbRule, err := c.GetProtobufRulesFromIptable(tableName)
	if err != nil {
		return nil, fmt.Errorf("error occurred during getting JSON rules from iptables : %w", err)
	}
	return c.jsonRuleList(pbRule)
}

// Convert list of protobuf rules to list of JSON rules.
func (c *Converter) jsonRuleList(pbRules []*pb.RuleResponse) ([][]byte, error) {
	ruleResListJSON := make([][]byte, 0)
	m := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: true,
	}
	for _, rule := range pbRules {
		ruleJSON, err := m.Marshal(rule) // pretty print
		if err != nil {
			return nil, fmt.Errorf("error occurred during marshaling : %w", err)
		}
		ruleResListJSON = append(ruleResListJSON, ruleJSON)
	}
	return ruleResListJSON, nil
}

// GetProtobufRulesFromIptableFile returns a list of protobuf rules from npmCache and iptable-save files.
func (c *Converter) GetProtobufRulesFromIptableFile(
	tableName string,
	npmCacheFile string,
	iptableSaveFile string,
) ([]*pb.RuleResponse, error) {

	err := c.initConverterFile(npmCacheFile)
	if err != nil {
		return nil, fmt.Errorf("error occurred during getting protobuf rules from iptables : %w", err)
	}

	ipTable, err := parse.IptablesFile(tableName, iptableSaveFile)
	if err != nil {
		return nil, fmt.Errorf("error occurred during parsing iptables : %w", err)
	}
	ruleResList, err := c.pbRuleList(ipTable)
	if err != nil {
		return nil, fmt.Errorf("error occurred during getting protobuf rules from iptables : %w", err)
	}

	return ruleResList, nil
}

// GetProtobufRulesFromIptable returns a list of protobuf rules from node.
func (c *Converter) GetProtobufRulesFromIptable(tableName string) ([]*pb.RuleResponse, error) {
	err := c.initConverter()
	if err != nil {
		return nil, fmt.Errorf("error occurred during getting protobuf rules from iptables : %w", err)
	}

	ipTable, err := parse.Iptables(tableName)
	if err != nil {
		return nil, fmt.Errorf("error occurred during parsing iptables : %w", err)
	}
	ruleResList, err := c.pbRuleList(ipTable)
	if err != nil {
		return nil, fmt.Errorf("error occurred during getting protobuf rules from iptables : %w", err)
	}

	return ruleResList, nil
}

// Create a list of protobuf rules from iptable.
func (c *Converter) pbRuleList(ipTable *NPMIPtable.Table) ([]*pb.RuleResponse, error) {
	ruleResList := make([]*pb.RuleResponse, 0)
	for _, v := range ipTable.Chains {
		chainRules, err := c.getRulesFromChain(v)
		if err != nil {
			return nil, fmt.Errorf("error occurred during getting protobuf rule list : %w", err)
		}
		ruleResList = append(ruleResList, chainRules...)
	}

	return ruleResList, nil
}

func (c *Converter) getRulesFromChain(iptableChain *NPMIPtable.Chain) ([]*pb.RuleResponse, error) {
	rules := make([]*pb.RuleResponse, 0)
	for _, v := range iptableChain.Rules {
		rule := &pb.RuleResponse{}
		rule.Chain = iptableChain.Name
		// filter other chains except for Azure NPM specific chains.
		if _, ok := c.AzureNPMChains[rule.Chain]; !ok {
			continue
		}
		rule.Protocol = v.Protocol
		switch v.Target.Name {
		case util.IptablesMark:
			rule.Allowed = true
		case util.IptablesDrop:
			rule.Allowed = false
		default:
			// ignore other targets
			continue
		}
		rule.Direction = c.getRuleDirection(iptableChain.Name)

		err := c.getModulesFromRule(v.Modules, rule)
		if err != nil {
			return nil, fmt.Errorf("error occurred during getting rules from chain : %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (c *Converter) getRuleDirection(iptableChainName string) pb.Direction {
	if strings.Contains(iptableChainName, "EGRESS") {
		return pb.Direction_EGRESS
	} else if strings.Contains(iptableChainName, "INGRESS") {
		return pb.Direction_INGRESS
	}
	return pb.Direction_UNDEFINED
}

func (c *Converter) getSetType(name string, m string) pb.SetType {
	if m == "ListMap" { // labels of namespace
		if strings.Contains(name, util.IpsetLabelDelimter) {
			if strings.Count(name, util.IpsetLabelDelimter) > 1 {
				return pb.SetType_NESTEDLABELOFPOD
			}
			return pb.SetType_KEYVALUELABELOFNAMESPACE
		}
		return pb.SetType_KEYLABELOFNAMESPACE
	}
	if strings.HasPrefix(name, util.NamespacePrefix) {
		return pb.SetType_NAMESPACE
	}
	if strings.HasPrefix(name, util.NamedPortIPSetPrefix) {
		return pb.SetType_NAMEDPORTS
	}
	if strings.Contains(name, util.IpsetLabelDelimter) {
		return pb.SetType_KEYVALUELABELOFPOD
	}
	matcher.Match([]byte(name))
	if matched := matcher.Match([]byte(name)); matched {
		return pb.SetType_CIDRBLOCKS
	}
	return pb.SetType_KEYLABELOFPOD
}

func (c *Converter) getModulesFromRule(moduleList []*NPMIPtable.Module, ruleRes *pb.RuleResponse) error {
	ruleRes.SrcList = make([]*pb.RuleResponse_SetInfo, 0)
	ruleRes.DstList = make([]*pb.RuleResponse_SetInfo, 0)
	ruleRes.UnsortedIpset = make(map[string]string)
	for _, module := range moduleList {
		switch module.Verb {
		case "set":
			// set module
			OptionValueMap := module.OptionValueMap
			for option, values := range OptionValueMap {
				switch option {
				case "match-set":
					setInfo := &pb.RuleResponse_SetInfo{}

					err := c.populateSetInfo(setInfo, values, ruleRes)
					if err != nil {
						return fmt.Errorf("error occurred during getting modules from rules : %w", err)
					}
					setInfo.Included = true

				case "not-match-set":
					setInfo := &pb.RuleResponse_SetInfo{}
					err := c.populateSetInfo(setInfo, values, ruleRes)
					if err != nil {
						return fmt.Errorf("error occurred during getting modules from rules : %w", err)
					}
					setInfo.Included = false
				default:
					// todo add warning log
					log.Printf("%v option have not been implemented\n", option)
					continue
				}
			}

		case "tcp", "udp":
			OptionValueMap := module.OptionValueMap
			for k, v := range OptionValueMap {
				if k == "dport" {
					portNum, _ := strconv.ParseInt(v[0], Base, Bitsize)
					ruleRes.DPort = int32(portNum)
				} else {
					portNum, _ := strconv.ParseInt(v[0], Base, Bitsize)
					ruleRes.SPort = int32(portNum)
				}
			}
		default:
			continue
		}
	}
	return nil
}

func (c *Converter) populateSetInfo(
	setInfo *pb.RuleResponse_SetInfo,
	values []string,
	ruleRes *pb.RuleResponse,
) error {

	ipsetHashedName := values[0]
	ipsetOrigin := values[1]
	setInfo.HashedSetName = ipsetHashedName
	if v, ok := c.ListMap[ipsetHashedName]; ok {
		setInfo.Name = v
		setInfo.Type = c.getSetType(v, "ListMap")
	} else if v, ok := c.SetMap[ipsetHashedName]; ok {
		setInfo.Name = v
		setInfo.Type = c.getSetType(v, "SetMap")
		if setInfo.Type == pb.SetType_CIDRBLOCKS {
			populateCIDRBlockSet(setInfo)
		}
	} else {
		return fmt.Errorf("%w : %v", errSetNotExist, ipsetHashedName)
	}

	if len(ipsetOrigin) > MinUnsortedIPSetLength {
		ruleRes.UnsortedIpset[ipsetHashedName] = ipsetOrigin
	}
	if strings.Contains(ipsetOrigin, "src") {
		ruleRes.SrcList = append(ruleRes.SrcList, setInfo)
	} else {
		ruleRes.DstList = append(ruleRes.DstList, setInfo)
	}
	return nil
}

// populate CIDRBlock set's content with ip addresses
func populateCIDRBlockSet(setInfo *pb.RuleResponse_SetInfo) {
	ipsetBuffer := bytes.NewBuffer(nil)
	cmdArgs := []string{"list", setInfo.HashedSetName}
	cmd := exec.Command(util.Ipset, cmdArgs...) //nolint:gosec

	cmd.Stdout = ipsetBuffer
	stderrBuffer := bytes.NewBuffer(nil)
	cmd.Stderr = stderrBuffer

	err := cmd.Run()
	if err != nil {
		_, err = stderrBuffer.WriteTo(ipsetBuffer)
		if err != nil {
			panic(err)
		}
	}
	curReadIndex := 0

	// finding the members field
	for curReadIndex < len(ipsetBuffer.Bytes()) {
		line, nextReadIndex := parse.Line(curReadIndex, ipsetBuffer.Bytes())
		curReadIndex = nextReadIndex
		if bytes.HasPrefix(line, MembersBytes) {
			break
		}
	}
	for curReadIndex < len(ipsetBuffer.Bytes()) {
		member, nextReadIndex := parse.Line(curReadIndex, ipsetBuffer.Bytes())
		setInfo.Contents = append(setInfo.Contents, string(member))
		curReadIndex = nextReadIndex
	}
}
