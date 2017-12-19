package main

import (
	"database/sql"
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
	ID         int            `db:"id"`
	KanjiName  string         `db:"kanji_name"`
	KataName   string         `db:"kata_name"`
	BankCode   string         `db:"bank_code"`
	BranchCode string         `db:"branch_code"`
	Address    sql.NullString `db:"address"`
	Phone      sql.NullString `db:"phone"`
}
