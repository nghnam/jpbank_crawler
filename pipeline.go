package main

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

// JPBankPipeline ...
type JPBankPipeline struct {
	db     *sqlx.DB
	itemCh <-chan interface{}
	doneCh chan struct{}
}

// NewJPBankPipeline ...
func NewJPBankPipeline(db *sqlx.DB, itemCh chan interface{}, doneCh chan struct{}) *JPBankPipeline {
	return &JPBankPipeline{db: db, itemCh: itemCh, doneCh: doneCh}
}

func (p *JPBankPipeline) Run() {
	for {
		item, more := <-p.itemCh
		if more {
			bank, isBank := item.(JPBank)
			if isBank {
				existed, err := p.IsBankExisted(bank.BankCode)
				if err != nil {
					log.Println(err)
				}
				if !existed {
					if err := p.CreateBank(bank.KanjiName, bank.BankCode); err != nil {
						log.Println(err)
					}
				}
			}

			branch, isBranch := item.(JPBankBranch)
			if isBranch {
				existed, err := p.IsBranchExisted(branch.BankCode, branch.BranchCode)
				if err != nil {
					log.Println(err)
				}
				if !existed {
					if err := p.CreateBranch(&branch); err != nil {
						log.Println(err)
					}
				}
			}
		} else {
			p.doneCh <- struct{}{}
		}
	}
}

func (p *JPBankPipeline) CreateBank(kanjiName, bankCode string) error {
	tx := p.db.MustBegin()
	tx.MustExec("INSERT INTO jp_bank(kanji_name, bank_code) VALUES ($1, $2)", kanjiName, bankCode)
	return tx.Commit()
}

func (p *JPBankPipeline) IsBankExisted(bankCode string) (bool, error) {
	bank := JPBank{}
	err := p.db.Get(&bank, "SELECT bank_code FROM jp_bank WHERE bank_code=$1", bankCode)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if bank.BankCode == bankCode {
		return true, nil
	}

	return false, nil
}

func (p *JPBankPipeline) CreateBranch(branch *JPBankBranch) error {
	tx := p.db.MustBegin()
	tx.NamedExec("INSERT INTO jp_bank_branch(kanji_name, kata_name, branch_code, bank_code, address, phone) VALUES (:kanji_name, :kata_name, :branch_code, :bank_code, :address, :phone)",
		branch)
	return tx.Commit()
}

func (p *JPBankPipeline) IsBranchExisted(bankCode, branchCode string) (bool, error) {
	branch := JPBankBranch{}
	err := p.db.Get(&branch, "SELECT bank_code, branch_code FROM jp_bank_branch WHERE bank_code=$1 and branch_code=$2", bankCode, branchCode)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if branch.BankCode == bankCode && branch.BranchCode == branchCode {
		return true, nil
	}

	return false, nil
}
