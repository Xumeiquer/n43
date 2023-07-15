package n43

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Norma43 struct {
	Accounts        []*Account
	ReportedEntries int
}

type Account struct {
	Header    *Header
	Movements []*Movement
	Footer    *Footer
}

type Header struct {
	BankCode            string
	BranchCode          string
	AccountNumber       string
	StartDate           time.Time
	EndDate             time.Time
	InitialBalance      float64
	Currency            string
	InformationModeCode string
	AccountName         string
}

type Movement struct {
	BranchCode       string
	TransactionDate  time.Time
	ValueDate        time.Time
	Amount           float64
	Balance          float64
	Description      string
	ExtraInformation []string
}

type Footer struct {
	BankCode      string
	BranchCode    string
	AccountNumber string
	DebitEntries  int
	DebitAmount   float64
	CreditEntries int
	CreditAmount  float64
	FinalBalance  float64
	Currency      string
}

type (
	LineType   int
	TimeFormat string
)

func (tf *TimeFormat) String() string {
	return string(*tf)
}

func (tf *TimeFormat) Set(val string) error {
	if d, ok := timeFormat[val]; ok {
		*tf = d
		return nil
	}
	return errors.New("invalid value for assigment")
}

const (
	HEADER_LINE              LineType = 11
	MOVEMENT_LINE            LineType = 22
	MOVEMENT_EXTRA_INFO_LINE LineType = 23
	FOOTER_LINE              LineType = 33
	END_OF_FILE_LINE         LineType = 88

	SPANISH_DATE TimeFormat = "DMY"
	ENGLISH_DATE TimeFormat = "YMD"
)

var timeFormat map[string]TimeFormat = map[string]TimeFormat{
	"DMY": SPANISH_DATE,
	"YMD": ENGLISH_DATE,
}

type Parser struct {
	lines       []string
	cLine       string
	pos         int
	n43         *Norma43
	parseOption *ParserOptions
}

type ParserOptions struct {
	Trim            bool
	TimeFormat      TimeFormat
	FilterPositive  bool
	FilterNegative  bool
	FilterLineIn    string
	filterLineInRe  *regexp.Regexp
	FilterLineOut   string
	filterLineOutRe *regexp.Regexp
}

func NewParser(lines []string, parserOptions *ParserOptions) *Parser {
	po := new(ParserOptions)

	po.TimeFormat = ENGLISH_DATE

	if parserOptions != nil {
		po.Trim = parserOptions.Trim
		po.TimeFormat = parserOptions.TimeFormat
		po.FilterPositive = parserOptions.FilterPositive
		po.FilterNegative = parserOptions.FilterNegative

		if parserOptions.FilterLineIn != "" {
			po.filterLineInRe = regexp.MustCompile(parserOptions.FilterLineIn)
		}
		if parserOptions.FilterLineOut != "" {
			po.filterLineOutRe = regexp.MustCompile(parserOptions.FilterLineOut)
		}
	}
	return &Parser{
		lines:       lines,
		pos:         -1,
		n43:         &Norma43{},
		parseOption: po,
	}
}

func NewParserReader(r io.Reader, parserOptions *ParserOptions) *Parser {
	fileScanner := bufio.NewScanner(r)

	fileScanner.Split(bufio.ScanLines)

	dataLines := []string{}
	for fileScanner.Scan() {
		dataLines = append(dataLines, fileScanner.Text())
	}

	return NewParser(dataLines, parserOptions)
}

func getLineType(code string) (LineType, error) {
	switch code {
	case "11":
		return HEADER_LINE, nil
	case "22":
		return MOVEMENT_LINE, nil
	case "23":
		return MOVEMENT_EXTRA_INFO_LINE, nil
	case "33":
		return FOOTER_LINE, nil
	case "88":
		return END_OF_FILE_LINE, nil
	}

	return LineType(0), errors.New(code + " is an invalid line code type")
}

func extract_date(date string, format TimeFormat) (time.Time, error) {
	year := ""
	month := ""
	day := ""

	for idx := 0; idx < len(format); idx++ {
		symbol := format[idx]
		extracted := date[idx*2 : idx*2+2]
		if symbol == 'Y' {
			year = "20" + extracted
		} else if symbol == 'D' {
			day = extracted
		} else if symbol == 'M' {
			month = extracted
		}
	}

	yearNumber, yErr := strconv.Atoi(year)
	monthNumber, mErr := strconv.Atoi(month)
	dayNumber, dErr := strconv.Atoi(day)

	if yErr != nil || mErr != nil || dErr != nil {
		return time.Now(), errors.New("wrong date format")
	}

	return time.Date(yearNumber, time.Month(monthNumber), dayNumber, 0, 0, 0, 0, time.UTC), nil
}

func (p *Parser) lineType() (LineType, error) {
	lineCode := p.getLine()[:2]
	return getLineType(lineCode)
}

func (p *Parser) getLine() string {
	return p.cLine
	// if p.parseOption.Trim {
	// 	return strings.TrimSpace(p.lines[p.pos])
	// }
	// return p.lines[p.pos]
}

func (p *Parser) nextLine() (string, error) {
	if p.pos+1 < len(p.lines) {
		p.pos++
		line := p.lines[p.pos]

		if p.parseOption.Trim {
			line = strings.TrimSpace(line)
		}
		p.cLine = line
		return line, nil
	}
	return "", errors.New("end of document")
}

func (p *Parser) next() (LineType, error) {
	_, err := p.nextLine()
	if err != nil {
		return LineType(0), err
	}

	return p.lineType()
}

func (p *Parser) peek() (string, error) {
	if p.pos+1 < len(p.lines) {
		return p.lines[p.pos+1], nil
	}
	return "", errors.New("end of document")
}

func (p *Parser) Parse() (*Norma43, error) {
	lineType, err := p.next()
	if err != nil {
		return p.n43, err
	}

header:

	if lineType == HEADER_LINE {
		// Process header
		h, err := p.parseHeader()
		if err != nil {
			return p.n43, err
		}

		p.n43.Accounts = make([]*Account, 0)
		account := new(Account)
		account.Header = h
		p.n43.Accounts = append(p.n43.Accounts, account)

	}

	lineType, err = p.next()
	if err != nil {
		return p.n43, err
	}

	for lineType != FOOTER_LINE {
		if lineType == MOVEMENT_LINE {
			// Process movements
			l, err := p.parseMovementLine()
			if err != nil {
				return p.n43, err
			}

			filtered := false

			if p.parseOption.FilterNegative && l.Amount < 0 {
				// filter
				peekLine, err := p.peek()
				if err != nil {
					return p.n43, errors.New("malformed document")
				}

				peekLineType, err := getLineType(peekLine[:2])
				if err != nil {
					return p.n43, err
				}

				if peekLineType == MOVEMENT_EXTRA_INFO_LINE {
					p.next()
				}

				filtered = true
			}

			if p.parseOption.FilterPositive && l.Amount > 0 {
				// filter
				peekLine, err := p.peek()
				if err != nil {
					return p.n43, errors.New("malformed document")
				}

				peekLineType, err := getLineType(peekLine[:2])
				if err != nil {
					return p.n43, err
				}

				if peekLineType == MOVEMENT_EXTRA_INFO_LINE {
					p.next()
				}

				filtered = true
			}

			if !filtered {
				p.n43.Accounts[len(p.n43.Accounts)-1].Movements = append(p.n43.Accounts[len(p.n43.Accounts)-1].Movements, l)
			}
		} else if lineType == MOVEMENT_EXTRA_INFO_LINE {
			// Process extra info

		extraInfoLine:
			line := p.getLine()

			if p.parseOption.filterLineInRe != nil {
				res := p.parseOption.filterLineInRe.FindString(line[4:])
				if res != "" { // match
					if p.parseOption.filterLineOutRe != nil {
						res = p.parseOption.filterLineOutRe.FindString(line[4:])
						if res != "" {
							// DO NOT INCLUDE
							p.purgeLastMovemnt()
						} else {
							// INCLUDE
							p.parseMovementLineExtraInfo()
						}
					} else {
						// INCLUDE
						p.parseMovementLineExtraInfo()
					}
				} else {
					// DO NOT INCLUDE
					p.purgeLastMovemnt()
				}
			}

			if p.parseOption.filterLineOutRe != nil {
				res := p.parseOption.filterLineOutRe.FindString(line[4:])
				if res != "" {
					// DO NOT INCLUDE
					p.purgeLastMovemnt()
				} else {
					// INCLUDE
					p.parseMovementLineExtraInfo()
				}
			}

			if p.parseOption.filterLineInRe == nil && p.parseOption.filterLineOutRe == nil {
				p.parseMovementLineExtraInfo()
			}

			peekLine, err := p.peek()
			if err != nil {
				return p.n43, errors.New("malformed document")
			}
			peekLineType, err := getLineType(peekLine[:2])
			if err != nil {
				return p.n43, err
			}

			if peekLineType == MOVEMENT_EXTRA_INFO_LINE {
				_, err = p.next()
				if err != nil {
					return p.n43, err
				}
				goto extraInfoLine
			}

		}

		lineType, err = p.next()
		if err != nil {
			return p.n43, err
		}
	}

	if lineType == FOOTER_LINE {
		// Process footer
		f, err := p.parseFooter()
		if err != nil {
			return p.n43, err
		}
		p.n43.Accounts[len(p.n43.Accounts)-1].Footer = f
	}

	if lineType == HEADER_LINE {
		goto header
	}

	if lineType == END_OF_FILE_LINE {
		// Process EOF
		line := p.getLine()
		reportedEntities, err := strconv.Atoi(line[20:])
		if err != nil {
			return p.n43, err
		}
		p.n43.ReportedEntries = reportedEntities
	}

	return p.n43, nil
}

func (p *Parser) parseHeader() (*Header, error) {
	h := new(Header)
	var err error

	line := p.getLine()

	h.BankCode = line[2:6]
	h.BranchCode = line[6:10]
	h.AccountNumber = line[10:20]
	h.StartDate, err = extract_date(line[20:26], ENGLISH_DATE)
	if err != nil {
		return h, err
	}
	h.EndDate, err = extract_date(line[26:32], ENGLISH_DATE)
	if err != nil {
		return h, err
	}

	initialBalanceSign := float64(1)
	if line[32:33] != "2" {
		initialBalanceSign = -1
	}
	balance, err := strconv.ParseFloat(line[33:47], 64)
	if err != nil {
		return h, err
	}
	h.InitialBalance = initialBalanceSign * balance / 100
	h.Currency = line[47:50]
	h.InformationModeCode = line[50:51]
	h.AccountName = line[51:]

	return h, nil
}

func (p *Parser) parseMovementLine() (*Movement, error) {
	m := new(Movement)
	var err error

	line := p.getLine()

	m.BranchCode = line[6:10]
	m.TransactionDate, err = extract_date(line[10:16], ENGLISH_DATE)
	if err != nil {
		return m, err
	}
	m.ValueDate, err = extract_date(line[16:22], ENGLISH_DATE)
	if err != nil {
		return m, err
	}
	amountSign := float64(1)
	if line[27:28] != "2" {
		amountSign = -1
	}
	amount, err := strconv.ParseFloat(line[28:42], 64)
	if err != nil {
		return m, err
	}
	m.Amount = amountSign * amount / 100
	m.Description = line[52:]

	if len(p.n43.Accounts[len(p.n43.Accounts)-1].Movements) == 0 {
		if p.n43.Accounts[len(p.n43.Accounts)-1].Header.InitialBalance == 0 {
			m.Balance = float64(0)
		} else {
			m.Balance = p.n43.Accounts[len(p.n43.Accounts)-1].Header.InitialBalance
		}
	} else {
		m.Balance = p.n43.Accounts[len(p.n43.Accounts)-1].Movements[len(p.n43.Accounts[len(p.n43.Accounts)-1].Movements)-1].Balance
	}
	m.Balance = m.Balance + m.Amount

	return m, nil
}

func (p *Parser) parseFooter() (*Footer, error) {
	f := new(Footer)

	line := p.getLine()

	f.BankCode = line[2:6]
	f.BranchCode = line[6:10]
	f.AccountNumber = line[10:20]
	debitEntries, err := strconv.Atoi(line[20:25])
	if err != nil {
		return f, err
	}
	f.DebitEntries = int(debitEntries)
	debitAmount, err := strconv.ParseFloat(line[25:39], 64)
	if err != nil {
		return f, err
	}
	f.DebitAmount = debitAmount / 100
	creditEntries, err := strconv.Atoi(line[39:44])
	if err != nil {
		return f, err
	}
	f.CreditEntries = creditEntries
	creditAmount, err := strconv.ParseFloat(line[44:58], 64)
	if err != nil {
		return f, err
	}
	f.CreditAmount = creditAmount
	finalBalanceSign := float64(1)
	if line[58:59] != "2" {
		finalBalanceSign = -1
	}
	finalBalance, err := strconv.ParseFloat(line[59:73], 64)
	if err != nil {
		return f, err
	}
	f.FinalBalance = finalBalanceSign * finalBalance / 100
	f.Currency = line[73:76]

	return f, nil
}

func (p *Parser) parseMovementLineExtraInfo() {
	line := p.getLine()

	lastAccount := len(p.n43.Accounts) - 1
	lastMovement := len(p.n43.Accounts[lastAccount].Movements) - 1

	p.n43.Accounts[lastAccount].Movements[lastMovement].ExtraInformation = append(p.n43.Accounts[lastAccount].Movements[lastMovement].ExtraInformation, line[4:])
}

func (p *Parser) purgeLastMovemnt() {
	lastAccount := len(p.n43.Accounts) - 1
	lastMovement := len(p.n43.Accounts[lastAccount].Movements) - 1

	p.n43.Accounts[lastAccount].Movements = remove(p.n43.Accounts[lastAccount].Movements, lastMovement)
}

func remove[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
