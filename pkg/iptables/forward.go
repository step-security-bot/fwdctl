package iptables

import (
	"fmt"
	"strconv"
	"strings"
)

func validateIface(iface string) error {
	if iface == "" {
		return fmt.Errorf("inteface name is empty")
	}
	return nil
}

func validateProto(proto string) error {
	if proto == "" {
		return fmt.Errorf("protocol name is empty")
	}
	if (proto != "tcp") && (proto != "udp") && (proto != "icmp") {
		return fmt.Errorf("protocol name not allowed")
	}
	return nil
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port number not allowed")
	}
	return nil
}

func validateAddress(address string) error {
	return nil
}

func ValidateForward(iface string, proto string, dport int, saddr string, sport int) error {
	err := validateIface(iface)
	if err != nil {
		return fmt.Errorf("interface: '%s' %v", iface, err)
	}

	err = validateProto(proto)
	if err != nil {
		return fmt.Errorf("protocol: '%s' %v", proto, err)
	}

	err = validatePort(dport)
	if err != nil {
		return fmt.Errorf("destination port: '%d' %v", dport, err)
	}

	err = validateAddress(saddr)
	if err != nil {
		return fmt.Errorf("source address: '%s' %v", saddr, err)
	}

	err = validatePort(sport)
	if err != nil {
		return fmt.Errorf("source port: '%d' %v", sport, err)
	}
	return nil
}

func CreateForward(iface string, proto string, dport int, saddr string, sport int) error {
	ipt, err := getIPTablesInstance()
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	// example rule:
	// iptables -t nat -A PREROUTING -i eth0 -p tcp -m tcp --dport 3000 -j DNAT --to-destination 192.168.199.105:80
	ruleSpec := []string{
		"-i", iface,
		"-p", proto,
		"-m", proto,
		"--dport", strconv.Itoa(dport),
		"-j", fwdTarget,
		"--to-destination", saddr + ":" + strconv.Itoa(sport),
	}

	err = ValidateForward(iface, proto, dport, saddr, sport)
	if err != nil {
		return fmt.Errorf("validation error: %v", err)
	}

	// check if input interface exists on the system
	ifaceExits, err := interfaceExists(iface)
	if err != nil {
		return fmt.Errorf("error reading interfaces: %v", err)
	}
	if !ifaceExits {
		return fmt.Errorf("interface %s does not exists", iface)
	}

	// check if provided rule already exists
	ruleExists, err := ipt.Exists(fwdTable, fwdChain, ruleSpec...)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if ruleExists {
		return fmt.Errorf("rule already exists")
	}

	// apply provided rule
	err = ipt.AppendUnique(fwdTable, fwdChain, ruleSpec...)
	if err != nil {
		return fmt.Errorf("rule failed: %v", err)
	}
	return nil
}

func ListForward(outputFormat string) ([]string, error) {
	ipt, err := getIPTablesInstance()
	if err != nil {
		return nil, fmt.Errorf("failed: %v", err)
	}

	//ruleList, err := ipt.ListWithCounters(fwdTable, fwdChain)
	ruleList, err := ipt.List(fwdTable, fwdChain)
	if err != nil {
		return nil, fmt.Errorf("failed: %v", err)
	}

	return ruleList, nil
}

func DeleteForwardById(ruleId int) error {
	ipt, err := getIPTablesInstance()
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	// retrieve rule using Id number
	// (sudo iptables -t nat -L PREROUTING -n --line-numbers)
	rule, err := ipt.ListById(fwdTable, fwdChain, ruleId)
	if err != nil {
		return fmt.Errorf("unable to retrieve rule with ID: %d", ruleId)
	}

	// cleaning rule (removing "-A PREROUTING", "-c 0 0", ...)
	ruleSplit := strings.Split(rule, " ")
	ruleSplit = append(ruleSplit[2:10], ruleSplit[13:]...)

	// delete rule
	err = ipt.Delete(fwdTable, fwdChain, ruleSplit...)
	if err != nil {
		return fmt.Errorf("failed deleting rule #%d\n err: %v", ruleId, err)
	}
	return nil
}

func DeleteForwardByRule(iface string, proto string, dport int, saddr string, sport int) error {
	ipt, err := getIPTablesInstance()
	if err != nil {
		return fmt.Errorf("failed: %v", err)
	}

	// TODO: create function to return []string with packed rule, passing iface, proto, etc as arguments.
	ruleSpec := []string{
		"-i", iface,
		"-p", proto,
		"-m", proto,
		"--dport", strconv.Itoa(dport),
		"-j", fwdTarget,
		"--to-destination", saddr + ":" + strconv.Itoa(sport),
	}

	err = ipt.Delete(fwdTable, fwdChain, ruleSpec...)
	if err != nil {
		return fmt.Errorf("failed deleting rule: '%s'\n err: %v", ruleSpec, err)
	}
	return nil
}
