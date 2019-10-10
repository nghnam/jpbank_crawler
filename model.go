package main

import (
	"database/sql"
	"strings"
)

// JPBank ...
type JPBank struct {
	ID        int            `db:"id"`
	KanjiName string         `db:"kanji_name"`
	KataName  sql.NullString `db:"kata_name"`
	BankCode  string         `db:"bank_code"`
}

// JPBankBranch ...
type JPBankBranch struct {
	ID          int            `db:"id"`
	KanjiName   string         `db:"kanji_name"`
	KataName    string         `db:"kata_name"`
	BankCode    string         `db:"bank_code"`
	BranchCode  string         `db:"branch_code"`
	Address     sql.NullString `db:"address"`
	Phone       sql.NullString `db:"phone"`
	setKataName bool
}

func setBranchAttribute(field string, value string, branch *JPBankBranch) {
	switch field {
	case "支店名":
		branch.KanjiName = value
	case "フリガナ":
		if !branch.setKataName {
			branch.setKataName = true
		} else {
			branch.KataName = value
		}
	case "金融機関コード":
		branch.BankCode = value
	case "支店コード":
		branch.BranchCode = value
	case "住所":
		address := strings.Split(value, " ")[0]
		branch.Address = sql.NullString{
			String: address,
			Valid:  true,
		}
	case "電話番号":
		branch.Phone = sql.NullString{
			String: value,
			Valid:  true,
		}
	}
}
