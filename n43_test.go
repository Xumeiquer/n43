package n43

import (
	"strings"
	"testing"
	"time"
)

func Test_n43(t *testing.T) {
	data := `111111222233334444122002032002102000000002463439783ACCOUNT NAME ************
22    22222002032002041240810000000000239900000000000000000000001234567890123456
2301COMPRA TARG 1234XXXXXXXX3456 SHOP TO BUY SEVERAL THINGS IN THERE.
22    2222200203200203032041000000000070290000000000AHSOWMSOWI8765SJWISU76WU
2301INSURANCE COMPANY ABC DEF GHI JKL MNO PQRS TUVWXYZ
22    2222200203200203032041000000000070290000000000A224E4ERF000000000123FFF
2301INSURANCE COMPANY ABC DEF GHI JKL MNO PQRS TUVWXYZ
22    22222002032002061240810000000001385700000000000000000000001234567890123456
2301CREDIT CARD 1234567890123456 1234 .SUPERMARKET WHATEVER NAME INC.
22    22222002032002031240810000000000010000000000000000000000001234567890123456
2301CREDIT CARD 1234567890123456 1234 .CAR GARAGE REPAIR.
3311112222333344441200015000000000661840000100000000050000200000000230159978
88999999999999999999000034`

	ops := new(ParserOptions)
	ops.Trim = true
	ops.TimeFormat = ENGLISH_DATE

	parser := NewParser(strings.Split(data, "\n"), ops)
	out, err := parser.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Accounts) != 1 {
		t.Errorf("Expected 1 account, but %d found", len(out.Accounts))
	}

	if len(out.Accounts[0].Movements) != 5 {
		t.Errorf("Expected 5 line movements, but %d found", len(out.Accounts[0].Movements))
	}

	if out.Accounts[0].Header.BankCode != "1111" {
		t.Errorf("Expected BankCode to be 1111, but %s found", out.Accounts[0].Header.BankCode)
	}

	if out.Accounts[0].Header.BranchCode != "2222" {
		t.Errorf("Expected BranchCode to be 2222, but %s found", out.Accounts[0].Header.BranchCode)
	}

	if out.Accounts[0].Header.AccountNumber != "3333444412" {
		t.Errorf("Expected AccountNumber to be 3333444412, but %s found", out.Accounts[0].Header.AccountNumber)
	}

	if out.Accounts[0].Header.InitialBalance != 2463.43 {
		t.Errorf("Expected InitialBalance to be 2463.43, but %f found", out.Accounts[0].Header.InitialBalance)
	}

	if out.Accounts[0].Header.Currency != "978" {
		t.Errorf("Expected Currency to be 978, but %s found", out.Accounts[0].Header.Currency)
	}

	if out.Accounts[0].Header.InformationModeCode != "3" {
		t.Errorf("Expected InformationModeCode to be 3, but %s found", out.Accounts[0].Header.InformationModeCode)
	}

	if out.Accounts[0].Header.AccountName != "ACCOUNT NAME ************" {
		t.Errorf("Expected AccountName to be ACCOUNT NAME ************, but %s found", out.Accounts[0].Header.BankCode)
	}

	ti, _ := time.Parse("2006-01-02 03:04:05 MST", "2020-02-03 00:00:00 UTC")
	if out.Accounts[0].Header.StartDate != ti {
		t.Errorf("Expected StartDate to be 2020-02-03 00:00:00 UTC, but %s found", out.Accounts[0].Header.StartDate)
	}

	ti, _ = time.Parse("2006-01-02 03:04:05 MST", "2020-02-10 00:00:00 UTC")
	if out.Accounts[0].Header.EndDate != ti {
		t.Errorf("Expected EndDate to be 2020-02-10 00:00:00 UTC, but %s found", out.Accounts[0].Header.EndDate)
	}

	if out.Accounts[0].Movements[0].BranchCode != "2222" {
		t.Errorf("Expected BranchCode in first movement to be 2222, but %s found", out.Accounts[0].Movements[0].BranchCode)
	}

	if out.Accounts[0].Movements[0].Amount != -23.99 {
		t.Errorf("Expected Amount in first movement to be -23.99, but %f found", out.Accounts[0].Movements[0].Amount)
	}

	if out.Accounts[0].Movements[0].Balance != 2439.44 {
		t.Errorf("Expected Balance in first movement to be 2439.44, but %f found", out.Accounts[0].Movements[0].Balance)
	}

	if out.Accounts[0].Movements[0].Description != "0000000000001234567890123456" {
		t.Errorf("Expected Description in first movement to be '0000000000001234567890123456', but %s found", out.Accounts[0].Movements[0].Description)
	}

	if out.Accounts[0].Movements[0].ExtraInformation[0] != "COMPRA TARG 1234XXXXXXXX3456 SHOP TO BUY SEVERAL THINGS IN THERE." {
		t.Errorf("Expected Description in first movement to be 'COMPRA TARG 1234XXXXXXXX3456 SHOP TO BUY SEVERAL THINGS IN THERE.', but %s found", out.Accounts[0].Movements[0].ExtraInformation[0])
	}

	ti, _ = time.Parse("2006-01-02 03:04:05 MST", "2020-02-03 00:00:00 UTC")
	if out.Accounts[0].Movements[0].TransactionDate != ti {
		t.Errorf("Expected EndDate to be 2020-02-03 00:00:00 UTC, but %s found", out.Accounts[0].Header.EndDate)
	}

	ti, _ = time.Parse("2006-01-02 03:04:05 MST", "2020-02-04 00:00:00 UTC")
	if out.Accounts[0].Movements[0].ValueDate != ti {
		t.Errorf("Expected EndDate to be 2020-02-04 00:00:00 UTC, but %s found", out.Accounts[0].Header.EndDate)
	}

	if out.Accounts[0].Footer.BankCode != "1111" {
		t.Errorf("Expected BankCode to be 1111, but %s found", out.Accounts[0].Footer.BankCode)
	}

	if out.Accounts[0].Footer.BranchCode != "2222" {
		t.Errorf("Expected BranchCode to be 2222, but %s found", out.Accounts[0].Footer.BranchCode)
	}

	if out.Accounts[0].Footer.AccountNumber != "3333444412" {
		t.Errorf("Expected AccountNumber to be 3333444412, but %s found", out.Accounts[0].Footer.AccountNumber)
	}

	if out.Accounts[0].Footer.DebitEntries != 15 {
		t.Errorf("Expected DebitEntries to be 15, but %d found", out.Accounts[0].Footer.DebitEntries)
	}

	if out.Accounts[0].Footer.DebitAmount != 661.84 {
		t.Errorf("Expected DebitAmount to be 661.84, but %f found", out.Accounts[0].Footer.DebitAmount)
	}

	if out.Accounts[0].Footer.CreditEntries != 1 {
		t.Errorf("Expected CreditEntries to be 1, but %d found", out.Accounts[0].Footer.CreditEntries)
	}

	if out.Accounts[0].Footer.CreditAmount != 50000.000000 {
		t.Errorf("Expected CreditAmount to be 50000.000000, but %f found", out.Accounts[0].Footer.CreditAmount)
	}

	if out.Accounts[0].Footer.FinalBalance != 2301.59 {
		t.Errorf("Expected FinalBalance to be 2301.59, but %f found", out.Accounts[0].Footer.FinalBalance)
	}

	if out.Accounts[0].Footer.Currency != "978" {
		t.Errorf("Expected Currency to be 978, but %s found", out.Accounts[0].Footer.Currency)
	}
}
