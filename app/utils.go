package main

import (
	"bytes"
	"compress/gzip"
	"log"
	"strings"
)

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func encodeValue(value string, encoding string) (bool, string, string) {
	if len(encoding) > 0 {
		schemas := strings.Split(encoding, ",")
		for i := 0; i < len(schemas); i++ {
			if strings.TrimSpace(schemas[i]) == "gzip" {
				var b bytes.Buffer
				gz := gzip.NewWriter(&b)
				if _, err := gz.Write([]byte(value)); err != nil {
					log.Fatal(err)
				}
				if err := gz.Close(); err != nil {
					log.Fatal(err)
				}

				return true, b.String(), "gzip"
			}
		}
	}

	return false, value, ""
}
