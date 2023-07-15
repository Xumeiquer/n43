package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/Xumeiquer/n43"
)

var (
	trim            bool           = false
	timeFormat      n43.TimeFormat = n43.ENGLISH_DATE
	filterPositives bool           = false
	filterNegatives bool           = false
	filterLineIn    string         = ""
	filterLineOut   string         = ""
	headerTpl       string         = ".BankCode,.BranchCode,.AccountNumber,.StartDate,.EndDate,.InitialBalance,.Currency,.InformationModeCode,.AccountName"
	lineTpl         string         = ".BranchCode,.TransactionDate,.ValueDate,.Amount,.Balance,.Description,.ExtraInformation"
	footerTpl       string         = ".BankCode,.BranchCode,.AccountNumber,.DebitEntries,.DebitAmount,.CreditEntries,.CreditAmount,.FinalBalance,.Currency"
	sepTpl          string         = " "
	fin             string         = ""
)

func init() {
	flag.BoolVar(&trim, "trim", trim, "Trim spaces surronding lines.")
	flag.Var(&timeFormat, "timeFormat", "Time parse format.")
	flag.BoolVar(&filterPositives, "filterPositive", filterPositives, "Filter positive values.")
	flag.BoolVar(&filterNegatives, "filterNegative", filterNegatives, "Filter negative values.")
	flag.StringVar(&filterLineIn, "filterLineIn", filterLineIn, "Filter (include) lines with extra information. This values will be used as regex.")
	flag.StringVar(&filterLineOut, "filterLineOut", filterLineOut, "Filter (exclude) lines with extra information. This values will be used as regex.")
	flag.StringVar(&headerTpl, "headerTpl", headerTpl, "Output template for the account header")
	flag.StringVar(&footerTpl, "footerTpl", footerTpl, "Output template for the account footer")
	flag.StringVar(&lineTpl, "lineTpl", lineTpl, "Output template for the movement line")
	flag.StringVar(&sepTpl, "sepTpl", sepTpl, "Sparator character")
	flag.StringVar(&fin, "in", fin, "Read from file.")

	flag.Parse()
}

func main() {
	ops := &n43.ParserOptions{
		Trim:           trim,
		TimeFormat:     timeFormat,
		FilterPositive: filterPositives,
		FilterNegative: filterNegatives,
		FilterLineIn:   filterLineIn,
		FilterLineOut:  filterLineOut,
	}

	if fin == "" {
		// read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeNamedPipe) != 0 {
			stdin, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal(err.Error())
			}

			dataLines := strings.Split(string(stdin), "\n")
			parser := n43.NewParser(dataLines, ops)
			res, err := parser.Parse()
			if err != nil {
				log.Fatal(err.Error())
			}

			printOutput(*res)
		}
	} else {
		data, err := os.ReadFile(fin)
		if err != nil {
			log.Fatal(err.Error())
		}

		dataLines := strings.Split(string(data), "\n")
		parser := n43.NewParser(dataLines, ops)
		res, err := parser.Parse()
		if err != nil {
			log.Fatal(err.Error())
		}

		printOutput(*res)
	}
}

func printOutput(res n43.Norma43) {
	tplData := TemplateData{
		Documents: []n43.Norma43{res},
	}

	tplGenerated := generateTeplate(headerTpl, lineTpl, footerTpl, sepTpl)

	tpl := template.Must(template.New("output").Parse(tplGenerated))
	err := tpl.Execute(os.Stdout, tplData)
	if err != nil {
		log.Fatalf("unable to generate the template. %s", err.Error())
	}
}

type TemplateData struct {
	Documents []n43.Norma43
}

func prepareTempaltes(tpl string, prefix string, sep string) string {
	if tpl == "" {
		return ""
	}
	tpls := strings.Split(tpl, ",")
	for i, t := range tpls {
		tpls[i] = "{{" + prefix + t + "}}"
	}
	return strings.Join(tpls, sep)
}

func generateTeplate(header string, line string, footer string, sep string) string {
	tpl := "{{- range $idx, $doc := .Documents }}{{- range $jdx, $account := $doc.Accounts }}"

	if header != "" {
		tpl += "\n" + prepareTempaltes(header, ".Header", sep)
	}

	if line != "" {
		tpl += "{{- range $kdx, $movement := .Movements }}\n" + prepareTempaltes(line, "", sep) + "{{ end }}"
	}

	if footer != "" {
		tpl += "\n" + prepareTempaltes(footer, ".Footer", sep)
	}

	tpl += "{{ end }}{{ end }}\n"
	return tpl
}
