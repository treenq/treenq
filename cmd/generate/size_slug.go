package main

import (
	_ "embed"
	"encoding/json"
	"io"
	"text/template"
	"unicode"
)

//go:embed inputs/sizes.json
var sizeGenInput []byte

//go:embed templates/sizes.txt
var sizesTemplate string

type Sizes struct {
	Package string
	Sizes   []struct {
		Slug         string `json:"slug"`
		Memory       int32  `json:"memory"`
		MemoryGB     float32
		Vcpus        int      `json:"vcpus"`
		Disk         int      `json:"disk"`
		PriceMonthly int      `json:"price_monthly"`
		PriceHourly  float64  `json:"price_hourly"`
		Regions      []string `json:"regions"`
		Available    bool     `json:"available"`
		Transfer     float64  `json:"transfer"`
		Description  string   `json:"description"`
		SlugCamel    string
	} `json:"sizes"`
	RetrievedAt string `json:"retrieved_at"`
}

func GenerateSizeSlugs(w io.Writer) error {
	var sizes Sizes
	err := json.Unmarshal(sizeGenInput, &sizes)
	if err != nil {
		return err
	}
	sizes.Package = "tqsdk"
	for i := range sizes.Sizes {
		sizes.Sizes[i].SlugCamel = kebabToCamel(sizes.Sizes[i].Slug)
		sizes.Sizes[i].MemoryGB = float32(sizes.Sizes[i].Memory) / 1024
	}

	tmpl, err := template.New("goTemplate").Parse(sizesTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(w, sizes)
}

func kebabToCamel(s string) string {
	var res string
	makeUpper := false
	for i := 0; i < len(s); i++ {
		if s[i] == '-' {
			makeUpper = true
			continue
		} else {
			if makeUpper {
				res += string(unicode.ToUpper(rune(s[i])))
				makeUpper = false
			} else {
				res += string(s[i])
			}
		}
	}

	return res
}
