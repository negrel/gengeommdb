package main

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

func main() {
	dbReader, err := getDatabaseGzip()
	if err != nil {
		respBody := ""
		if dbReader != nil {
			respBytes, _ := io.ReadAll(dbReader)
			respBody = string(respBytes)
		}

		panic(fmt.Errorf("failed to download geo database: %w\nResponse body: %v", err, respBody))
	}

	decompressedDbReader, err := gzip.NewReader(dbReader)
	if err != nil {
		panic(fmt.Errorf("failed to create gzip reader for database: %w", err))
	}

	tsvDbReader := csv.NewReader(decompressedDbReader)
	tsvDbReader.Comma = '\t'
	tsvDbReader.Comment = '#'

	dbWriter, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType: "ip2asn-combined",
		Description: map[string]string{
			"en": "gengeommdb - Open Source IP to country mapping database by Alexandre Negrel",
		},
		DisableIPv4Aliasing:     true,
		IncludeReservedNetworks: true,
		IPVersion:               6,
		Languages:               []string{"en"},
		RecordSize:              32,
		DisableMetadataPointers: false,
		Inserter:                nil,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create mmdb writer: %w", err))
	}

	for {
		record, err := tsvDbReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Errorf("failed to decode records of tsv file: %w", err))
		}
		if len(record) != 5 {
			panic("each record should contains 5 values, please update the go script")
		}

		ipRangeStart := net.ParseIP(record[0])
		ipRangeEnd := net.ParseIP(record[1])
		// asNum := record[2]
		isoCode := record[3]
		// asDesc := record[4]

		err = dbWriter.InsertRange(ipRangeStart, ipRangeEnd, mmdbtype.Map{
			mmdbtype.String("country"): mmdbtype.Map{
				mmdbtype.String("iso_code"): mmdbtype.String(isoCode),
			},
		})
		if err != nil {
			panic(fmt.Errorf("failed to insert range in mmdb: %w", err))
		}
	}

	_, err = dbWriter.WriteTo(os.Stdout)
	if err != nil {
		panic(fmt.Errorf("failed to write to stdout: %w", err))
	}
}

func getDatabaseGzip() (io.Reader, error) {
	resp, err := http.Get("https://iptoasn.com/data/ip2asn-combined.tsv.gz")
	if err != nil {
		return nil, fmt.Errorf("GET request to download ip2asn gzip database failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return resp.Body, fmt.Errorf("server returned non 200 (%v) response", resp.Status)
	}

	return resp.Body, nil
}
