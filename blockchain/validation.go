package blockchain

/*
Use these function to validate common data in blockchain
*/

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
	"golang.org/x/crypto/sha3"
)

/*
IsSalaryTx determines whether or not a transaction is a salary.
*/
func (self *BlockChain) IsSalaryTx(tx transaction.Transaction) bool {
	// Check normal tx(not an action tx)
	if tx.GetType() != common.TxSalaryType {
		normalTx, ok := tx.(*transaction.Tx)
		if !ok {
			return false
		}
		// Check nullifiers in every Descs
		descs := normalTx.Descs
		if len(descs) != 1 {
			return false
		} else {
			if descs[0].Reward > 0 {
				return true
			}
		}
		return false
	}
	return false
}

// ValidateDoubleSpend - check double spend for any transaction type
func (self *BlockChain) ValidateDoubleSpend(tx transaction.Transaction, chainID byte) error {
	txHash := tx.Hash()
	txViewPoint, err := self.FetchTxViewPoint(chainID)
	if err != nil {
		str := fmt.Sprintf("Can not check double spend for tx")
		err := NewBlockChainError(CanNotCheckDoubleSpendError, errors.New(str))
		return err
	}
	nullifierDb := txViewPoint.ListNullifiers()
	var descs []*transaction.JoinSplitDesc
	if tx.GetType() == common.TxNormalType {
		descs = tx.(*transaction.Tx).Descs
	} else if tx.GetType() == common.TxRegisterCandidateType {
		descs = tx.(*transaction.TxRegisterCandidate).Descs
	}
	for _, desc := range descs {
		for _, nullifer := range desc.Nullifiers {
			existed, err := common.SliceBytesExists(nullifierDb, nullifer)
			if err != nil {
				str := fmt.Sprintf("Can not check double spend for tx")
				err := NewBlockChainError(CanNotCheckDoubleSpendError, errors.New(str))
				return err
			}
			if existed {
				str := fmt.Sprintf("Nullifiers of transaction %+v already existed", txHash.String())
				err := NewBlockChainError(CanNotCheckDoubleSpendError, errors.New(str))
				return err
			}
		}
	}
	return nil
}

func (self *BlockChain) ValidateTxLoanRequest(tx transaction.Transaction, chainID byte) error {
	txLoan, ok := tx.(*transaction.TxLoanRequest)
	if !ok {
		return fmt.Errorf("Fail parsing LoanRequest transaction")
	}

	// Check if loan's params are correct
	currentParams := self.BestState[chainID].BestBlock.Header.LoanParams
	if txLoan.Params != currentParams {
		return fmt.Errorf("LoanRequest transaction has incorrect params")
	}

	// Check if loan id is unique across all chains
	// TODO(@0xbunyip): should we check in db/chain or only in best state?
	for chainID, bestState := range self.BestState {
		for _, id := range bestState.LoanIDs {
			if bytes.Equal(txLoan.LoanID, id) {
				return fmt.Errorf("LoanID already existed on chain %d", chainID)
			}
		}
	}
	return nil
}

func (self *BlockChain) ValidateTxLoanResponse(tx transaction.Transaction, chainID byte) error {
	txResponse, ok := tx.(*transaction.TxLoanResponse)
	if !ok {
		return fmt.Errorf("Fail parsing LoanResponse transaction")
	}

	// Check if a loan request with the same id exists on any chain
	txHashes, err := self.config.DataBase.GetLoanTxs(txResponse.LoanID)
	if err != nil {
		return err
	}
	found := false
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := self.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetType() {
		case common.TxLoanResponse:
			{
				return fmt.Errorf("Loan already had response")
			}
		case common.TxLoanRequest:
			{
				_, ok := txOld.(*transaction.TxLoanRequest)
				if !ok {
					return fmt.Errorf("Error parsing loan request tx")
				}
				found = true
			}
		}
	}

	if found == false {
		return fmt.Errorf("Corresponding loan request not found")
	}
	return nil
}

func (self *BlockChain) ValidateTxLoanPayment(tx transaction.Transaction, chainID byte) error {
	txPayment, ok := tx.(*transaction.TxLoanPayment)
	if !ok {
		return fmt.Errorf("Fail parsing LoanPayment transaction")
	}

	// Check if a loan request with the same id exists on any chain
	txHashes, err := self.config.DataBase.GetLoanTxs(txPayment.LoanID)
	if err != nil {
		return err
	}
	found := false
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := self.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetType() {
		case common.TxLoanResponse:
			{
				found = true
			}
		}
	}

	if found == false {
		return fmt.Errorf("Corresponding loan response not found")
	}
	return nil
}

func (self *BlockChain) ValidateTxLoanWithdraw(tx transaction.Transaction, chainID byte) error {
	txWithdraw, ok := tx.(*transaction.TxLoanWithdraw)
	if !ok {
		return fmt.Errorf("Fail parsing LoanResponse transaction")
	}

	// Check if a loan response with the same id exists on any chain
	txHashes, err := self.config.DataBase.GetLoanTxs(txWithdraw.LoanID)
	if err != nil {
		return err
	}
	foundResponse := false
	keyCorrect := false
	for _, txHash := range txHashes {
		hash := &common.Hash{}
		copy(hash[:], txHash)
		_, _, _, txOld, err := self.GetTransactionByHash(hash)
		if txOld == nil || err != nil {
			return fmt.Errorf("Error finding corresponding loan request")
		}
		switch txOld.GetType() {
		case common.TxLoanRequest:
			{
				// Check if key is correct
				txRequest, ok := tx.(*transaction.TxLoanRequest)
				if !ok {
					return fmt.Errorf("Error parsing corresponding loan request")
				}
				h := make([]byte, 32)
				sha3.ShakeSum256(h, txWithdraw.Key)
				if bytes.Equal(h, txRequest.KeyDigest) {
					keyCorrect = true
				}
			}
		case common.TxLoanResponse:
			{
				// Check if loan is accepted
				txResponse, ok := tx.(*transaction.TxLoanResponse)
				if !ok {
					return fmt.Errorf("Error parsing corresponding loan response")
				}
				if txResponse.Response != transaction.Accept {
					foundResponse = true
				}
			}

		}
	}

	if !foundResponse {
		return fmt.Errorf("Corresponding loan response not found")
	} else if !keyCorrect {
		return fmt.Errorf("Provided key is incorrect")
	}
	return nil
}

func (self *BlockChain) GetAmountPerAccount(proposal *transaction.PayoutProposal) (uint64, [][]byte, []uint64, error) {
	// TODO(@0xbunyip): cache list so that list of receivers is fixed across blocks
	tokenHolders, err := self.GetListTokenHolders(proposal.TokenID)
	if err != nil {
		return 0, nil, nil, err
	}

	// Get total token supply
	totalTokenSupply := uint64(0)
	for _, holder := range tokenHolders {
		utxos := self.GetAccountUTXO(holder)
		for i := 0; i < len(utxos); i += 1 {
			// TODO(@0xbunyip): get amount from utxo hash
			value := uint64(0)
			totalTokenSupply += value
		}
	}

	// Get amount per account
	rewardHolders := [][]byte{}
	amounts := []uint64{}
	for _, holder := range tokenHolders {
		utxos := self.GetAccountUTXO(holder) // Cached data
		amount := uint64(0)
		for i := 0; i < len(utxos); i += 1 {
			reward, err := self.GetUTXOReward(utxos[i]) // Data from latest block
			if err != nil {
				return 0, nil, nil, err
			}
			if reward < proposal.PayoutID {
				// TODO(@0xbunyip): get amount from utxo hash
				value := uint64(0)
				amount += value
			}
		}

		if amount > 0 {
			rewardHolders = append(rewardHolders, holder)
			amounts = append(amounts, amount)
		}
	}
	return totalTokenSupply, rewardHolders, amounts, nil
}

func (self *BlockChain) ValidateTxDividendPayout(tx transaction.Transaction, chainID byte) error {
	txPayout, ok := tx.(*transaction.TxDividendPayout)
	if !ok {
		return fmt.Errorf("Fail parsing DividendPayout transaction")
	}

	// Check if there's a proposal to pay dividend
	// TODO(@0xbunyip): get current proposal and check if it is dividend payout
	proposal := &transaction.PayoutProposal{}
	_, tokenHolders, amounts, err := self.GetAmountPerAccount(proposal)
	if err != nil {
		return err
	}

	// Check if user is not rewarded and amount is correct
	for _, desc := range txPayout.Descs {
		for _, note := range desc.Note {
			// Check if user is not rewarded
			utxos := self.GetAccountUTXO(note.Apk[:])
			for _, utxo := range utxos {
				reward, err := self.GetUTXOReward(utxo)
				if err != nil {
					return err
				}
				if reward >= proposal.PayoutID {
					return fmt.Errorf("UTXO %s has already received dividend payment", string(utxo))
				}
			}

			// Check amount
			found := 0
			for i, holder := range tokenHolders {
				if bytes.Equal(holder, note.Apk[:]) {
					found += 1
					if amounts[i] != note.Value {
						return fmt.Errorf("Payment amount for user %s incorrect, found %d instead of %d", holder, note.Value, amounts[i])
					}
				}
			}

			if found == 0 {
				return fmt.Errorf("User %s isn't eligible for receiving dividend", note.Apk[:])
			} else if found > 1 {
				return fmt.Errorf("Multiple dividend payments found for user %s", note.Apk[:])
			}
		}
	}

	return nil
}
