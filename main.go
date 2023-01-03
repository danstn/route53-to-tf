package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type RecordSet struct {
	Name            string `json:"Name"`
	Type            string `json:"Type"`
	TTL             int    `json:"TTL"`
	ResourceRecords []struct {
		Value string `json:"Value"`
	} `json:"ResourceRecords"`
	AliasTarget struct {
		DNSName              string `json:"DNSName"`
		EvaluateTargetHealth bool   `json:"EvaluateTargetHealth"`
		HostedZoneId         string `json:"HostedZoneId"`
	} `json:"AliasTarget"`
}

type ListResourceRecordSetsOutput struct {
	RecordSets []RecordSet `json:"ResourceRecordSets"`
}

var showImports = true

const KeyZoneID = "zone_id"

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go ZONE_ID DOMAIN")
		os.Exit(1)
	}

	zoneId := os.Args[1]
	domain := os.Args[2]

	// read from stdio
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println("Failed to read from stdin:", err)
		os.Exit(1)
	}

	// parse json
	var output ListResourceRecordSetsOutput
	err = json.Unmarshal(input, &output)
	if err != nil {
		fmt.Println("Failed to parse JSON:", err)
		os.Exit(1)
	}

	// render locals
	zoneRef := dashify(domain)
	fmt.Printf("resource aws_route53_zone %s {\n", zoneRef)
	fmt.Printf("  name = \"%s\"\n", domain)
	fmt.Print("}\n\n")

	zoneImport := fmt.Sprintf("terraform import aws_route53_zone.%s %s", dashify(domain), zoneId)
	imports := []string{
		zoneImport,
	}

	// render records
	for _, rr := range output.RecordSets {
		isAlias := len(rr.ResourceRecords) == 0

		cleanDomain := removeTrailingDot(rr.Name)
		recordRef := fmt.Sprintf("%s-%s", dashify(cleanDomain), strings.ToLower(rr.Type))
		fmt.Printf("resource \"aws_route53_record\" \"record-%s\" {\n", recordRef)
		fmt.Printf("  name    = \"%s\"\n", cleanDomain)
		fmt.Printf("  type    = \"%s\"\n", rr.Type)
		if !isAlias {
			fmt.Printf("  ttl     = %d\n", rr.TTL)
		}
		fmt.Printf("  zone_id = aws_route53_zone.%s.zone_id\n", zoneRef)
		if isAlias {
			// Handle alias resource record sets
			fmt.Printf("  alias {\n")
			fmt.Printf("    name                   = \"%s\"\n", rr.AliasTarget.DNSName)
			fmt.Printf("    evaluate_target_health = %t\n", rr.AliasTarget.EvaluateTargetHealth)
			fmt.Printf("    zone_id                = \"%s\"\n", rr.AliasTarget.HostedZoneId)
			fmt.Printf("  }")
		} else {
			// Handle regular resource record sets
			fmt.Printf("  records = [")
			for i, r := range rr.ResourceRecords {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("\n    \"%s\"", strings.ReplaceAll(r.Value, "\"", ""))
			}
			fmt.Print("\n  ]")
		}
		fmt.Println()

		fmt.Print("}\n\n")

		r := fmt.Sprintf("terraform import aws_route53_record.record-%s-%s %s_%s_%s",
			dashify(cleanDomain),
			strings.ToLower(rr.Type),
			zoneId,
			dashify(cleanDomain),
			rr.Type)
		imports = append(imports, r)
	}

	if showImports {
		for _, r := range imports {
			fmt.Println(r)
		}
	}
}

func dashify(withDots string) string {
	return strings.ReplaceAll(withDots, ".", "-")
}

func removeTrailingDot(domain string) string {
	return strings.TrimRight(domain, ".")
}
