package privacy_v2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy/key"

	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumbernoprivacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/utils"
	"github.com/incognitochain/incognito-chain/privacy/proof/agg_interface"
)

// For conversion proof, its version will be counted down from 255 -> 0
// It should contain inputCoins of v1 and outputCoins of v2 because it convert v1 to v2
type ConversionProofVer1ToVer2 struct {
	Version                    uint8
	inputCoins                 []*coin.PlainCoinV1
	outputCoins                []*coin.CoinV2
	serialNumberNoPrivacyProof []*serialnumbernoprivacy.SNNoPrivacyProof
}

func (proof ConversionProofVer1ToVer2) MarshalJSON() ([]byte, error) {
	data := proof.Bytes()
	//temp := base58.Base58Check{}.Encode(data, common.ZeroByte)
	temp := base64.StdEncoding.EncodeToString(data)
	return json.Marshal(temp)
}

func (proof *ConversionProofVer1ToVer2) UnmarshalJSON(data []byte) error {
	dataStr := common.EmptyString
	errJson := json.Unmarshal(data, &dataStr)
	if errJson != nil {
		return errJson
	}
	temp, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return err
	}
	errSetBytes := proof.SetBytes(temp)
	if errSetBytes != nil {
		return errSetBytes
	}
	return nil
}

func (proof ConversionProofVer1ToVer2) Init() {
	proof.Version = ConversionProofVersion
	proof.inputCoins = []*coin.PlainCoinV1{}
	proof.outputCoins = []*coin.CoinV2{}
	proof.serialNumberNoPrivacyProof = []*serialnumbernoprivacy.SNNoPrivacyProof{}
}

func (proof ConversionProofVer1ToVer2) GetVersion() uint8 { return ConversionProofVersion }
func (proof *ConversionProofVer1ToVer2) SetVersion(uint8) { proof.Version = ConversionProofVersion }

func (proof ConversionProofVer1ToVer2) GetInputCoins() []coin.PlainCoin {
	res := make([]coin.PlainCoin, len(proof.inputCoins))
	for i := 0; i < len(proof.inputCoins); i += 1 {
		res[i] = proof.inputCoins[i]
	}
	return res
}
func (proof ConversionProofVer1ToVer2) GetOutputCoins() []coin.Coin {
	res := make([]coin.Coin, len(proof.outputCoins))
	for i := 0; i < len(proof.outputCoins); i += 1 {
		res[i] = proof.outputCoins[i]
	}
	return res
}

// InputCoins should be all ver1, else it would crash
func (proof *ConversionProofVer1ToVer2) SetInputCoins(v []coin.PlainCoin) error {
	proof.inputCoins = make([]*coin.PlainCoinV1, len(v))
	for i := 0; i < len(v); i += 1 {
		coin, ok := v[i].(*coin.PlainCoinV1)
		if !ok {
			return errors.New("Input coins should all be PlainCoinV1")
		}
		proof.inputCoins[i] = coin
	}
	return nil
}

// v should be all coinv2 or else it would crash
func (proof *ConversionProofVer1ToVer2) SetOutputCoins(v []coin.Coin) error {
	proof.outputCoins = make([]*coin.CoinV2, len(v))
	for i := 0; i < len(v); i += 1 {
		coin, ok := v[i].(*coin.CoinV2)
		if !ok {
			return errors.New("Output coins should all be CoinV2")
		}
		proof.outputCoins[i] = coin
	}
	return nil
}

// Conversion does not have range proof, everything is nonPrivacy
func (proof ConversionProofVer1ToVer2) GetAggregatedRangeProof() agg_interface.AggregatedRangeProof {
	return nil
}

func (proof ConversionProofVer1ToVer2) Bytes() []byte {
	proofBytes := []byte{ConversionProofVersion}

	// InputCoins
	proofBytes = append(proofBytes, byte(len(proof.inputCoins)))
	for i := 0; i < len(proof.inputCoins); i++ {
		inputCoins := proof.inputCoins[i].Bytes()
		proofBytes = append(proofBytes, byte(len(inputCoins)))
		proofBytes = append(proofBytes, inputCoins...)
	}

	// OutputCoins
	proofBytes = append(proofBytes, byte(len(proof.outputCoins)))
	for i := 0; i < len(proof.outputCoins); i++ {
		outputCoins := proof.outputCoins[i].Bytes()
		proofBytes = append(proofBytes, byte(len(outputCoins)))
		proofBytes = append(proofBytes, outputCoins...)
	}

	// SNNoPrivacyProofSize
	proofBytes = append(proofBytes, byte(len(proof.serialNumberNoPrivacyProof)))
	for i := 0; i < len(proof.serialNumberNoPrivacyProof); i++ {
		snNoPrivacyProof := proof.serialNumberNoPrivacyProof[i].Bytes()
		proofBytes = append(proofBytes, byte(utils.SnNoPrivacyProofSize))
		proofBytes = append(proofBytes, snNoPrivacyProof...)
	}

	return proofBytes
}

func (proof *ConversionProofVer1ToVer2) SetBytes(proofBytes []byte) *errhandler.PrivacyError {
	if len(proofBytes) == 0 {
		return errhandler.NewPrivacyErr(errhandler.InvalidInputToSetBytesErr, errors.New("Proof bytes = 0"))
	}
	if proofBytes[0] != proof.GetVersion() {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Proof bytes version is not correct"))
	}
	if proof == nil {
		proof = new(ConversionProofVer1ToVer2)
	}
	proof.SetVersion(ConversionProofVersion)
	offset := 1

	//InputCoins  []*coin.PlainCoinV1
	if offset >= len(proofBytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
	}
	lenInputCoinsArray := int(proofBytes[offset])
	offset += 1
	proof.inputCoins = make([]*coin.PlainCoinV1, lenInputCoinsArray)
	for i := 0; i < lenInputCoinsArray; i++ {
		if offset >= len(proofBytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		lenInputCoin := int(proofBytes[offset])
		offset += 1

		if offset+lenInputCoin > len(proofBytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range input coins"))
		}
		coinBytes := proofBytes[offset : offset+lenInputCoin]
		if pc, err := coin.NewPlainCoinFromByte(coinBytes); err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		} else {
			var ok bool
			if proof.inputCoins[i], ok = pc.(*coin.PlainCoinV1); !ok {
				err := errors.New("Cannot assert type of PlainCoin to PlainCoinV1")
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
			}
		}
		offset += lenInputCoin
	}

	//OutputCoins  []*coin.CoinV2
	if offset >= len(proofBytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
	}
	lenOutputCoinsArray := int(proofBytes[offset])
	offset += 1
	proof.outputCoins = make([]*coin.CoinV2, lenOutputCoinsArray)
	for i := 0; i < lenOutputCoinsArray; i++ {
		if offset >= len(proofBytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		lenOutputCoin := int(proofBytes[offset])
		offset += 1

		if offset+lenOutputCoin > len(proofBytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range output coins"))
		}
		coinBytes := proofBytes[offset : offset+lenOutputCoin]
		if pc, err := coin.NewPlainCoinFromByte(coinBytes); err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		} else {
			var ok bool
			if proof.outputCoins[i], ok = pc.(*coin.CoinV2); !ok {
				err := errors.New("Cannot assert type of Coin to CoinV2")
				return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
			}
		}
		offset += lenOutputCoin
	}

	// SNNoPrivacyProof
	// Set SNNoPrivacyProofSize
	if offset >= len(proofBytes) {
		return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number no privacy proof"))
	}
	lenSNNoPrivacyProofArray := int(proofBytes[offset])
	offset += 1
	proof.serialNumberNoPrivacyProof = make([]*serialnumbernoprivacy.SNNoPrivacyProof, lenSNNoPrivacyProofArray)
	for i := 0; i < lenSNNoPrivacyProofArray; i++ {
		if offset >= len(proofBytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number no privacy proof"))
		}
		lenSNNoPrivacyProof := int(proofBytes[offset])
		offset += 1

		proof.serialNumberNoPrivacyProof[i] = new(serialnumbernoprivacy.SNNoPrivacyProof).Init()
		if offset+lenSNNoPrivacyProof > len(proofBytes) {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, errors.New("Out of range serial number no privacy proof"))
		}
		err := proof.serialNumberNoPrivacyProof[i].SetBytes(proofBytes[offset : offset+lenSNNoPrivacyProof])
		if err != nil {
			return errhandler.NewPrivacyErr(errhandler.SetBytesProofErr, err)
		}
		offset += lenSNNoPrivacyProof
	}
	return nil
}

func (proof *ConversionProofVer1ToVer2) IsPrivacy() bool {
	return false
}

func ProveConversion(inputCoins []coin.PlainCoin, outputCoins []*coin.CoinV2, serialnumberWitness []*serialnumbernoprivacy.SNNoPrivacyWitness) (*ConversionProofVer1ToVer2, error) {
	var err error
	proof := new(ConversionProofVer1ToVer2)
	proof.SetVersion(ConversionProofVersion)
	if err = proof.SetInputCoins(inputCoins); err != nil {
		Logger.Log.Errorf("Cannot set input coins in payment_v2 proof: err %v", err)
		return nil, err
	}
	outputCoinsV2 := make([]coin.Coin, len(outputCoins))
	for i := 0; i < len(outputCoins); i += 1 {
		outputCoinsV2[i] = outputCoins[i]
	}
	if err = proof.SetOutputCoins(outputCoinsV2); err != nil {
		Logger.Log.Errorf("Cannot set output coins in payment_v2 proof: err %v", err)
		return nil, err
	}

	// Proving that serial number is derived from the committed derivator
	for i := 0; i < len(inputCoins); i++ {
		snNoPrivacyProof, err := serialnumberWitness[i].Prove(nil)
		if err != nil {
			return nil, errhandler.NewPrivacyErr(errhandler.ProveSerialNumberNoPrivacyErr, err)
		}
		proof.serialNumberNoPrivacyProof = append(proof.serialNumberNoPrivacyProof, snNoPrivacyProof)
	}
	// Hide the keyimage :D
	for i := 0; i < len(proof.outputCoins); i++ {
		proof.outputCoins[i].SetKeyImage(nil)
	}
	return proof, nil
}

func (proof *ConversionProofVer1ToVer2) ValidateSanity() (bool, error) {
	if proof.Version != ConversionProofVersion {
		return false, errors.New("Proof version of ConversionProof is not right")
	}
	if len(proof.outputCoins) != 1 {
		return false, errors.New("ConversionProof len(outputCoins) should be 1, found len(outputCoins) != 1")
	}
	for i := 0; i < len(proof.serialNumberNoPrivacyProof); i++ {
		if !proof.serialNumberNoPrivacyProof[i].ValidateSanity() {
			return false, errors.New("validate sanity Serial number no privacy proof failed")
		}
	}
	if len(proof.inputCoins) != len(proof.serialNumberNoPrivacyProof) {
		return false, errors.New("validate sanity Serial number length is not the same with input coins length")
	}
	// check input coins without privacy
	for i := 0; i < len(proof.inputCoins); i++ {
		if !proof.inputCoins[i].GetCommitment().PointValid() {
			return false, errors.New("validate sanity CoinCommitment of input coin failed")
		}
		if !proof.inputCoins[i].GetPublicKey().PointValid() {
			return false, errors.New("validate sanity PublicKey of input coin failed")
		}
		if !proof.inputCoins[i].GetKeyImage().PointValid() {
			return false, errors.New("validate sanity Serial number of input coin failed")
		}
		if !proof.inputCoins[i].GetRandomness().ScalarValid() {
			return false, errors.New("validate sanity Randomness of input coin failed")
		}
		if !proof.inputCoins[i].GetSNDerivator().ScalarValid() {
			return false, errors.New("validate sanity SNDerivator of input coin failed")
		}
		if proof.inputCoins[i].IsEncrypted() {
			return false, errors.New("validate sanity input coin isEncrypted failed")
		}
	}


	// check output coins without privacy
	for i := 0; i < len(proof.outputCoins); i++ {
		if !proof.outputCoins[i].GetCommitment().PointValid() {
			return false, errors.New("validate sanity CoinCommitment of output coin failed")
		}
		if !proof.outputCoins[i].GetPublicKey().PointValid() {
			return false, errors.New("validate sanity PublicKey of output coin failed")
		}
		if !proof.outputCoins[i].GetRandomness().ScalarValid() {
			return false, errors.New("validate sanity Randomness of output coin failed")
		}
		if proof.outputCoins[i].IsEncrypted() {
			return false, errors.New("validate sanity input coin isEncrypted failed")
		}
	}
	return true, nil
}

func (proof ConversionProofVer1ToVer2) Verify(hasPrivacy bool, pubKey key.PublicKey, fee uint64, shardID byte, tokenID *common.Hash, isBatch bool, additionalData interface{}) (bool, error) {
	//Step to verify ConversionProofVer1ToVer2
	//	- verify sumInput = sumOutput + fee
	//	- verify if serial number of each input coin has been derived correctly
	//	- verify input coins' randomness
	//	- verify if output coins' commitment has been calculated correctly
	if hasPrivacy {
		return false, errors.New("ConversionProof does not have privacy, something is wrong")
	}
	sumInput, sumOutput := uint64(0), uint64(0)
	for i := 0; i < len(proof.inputCoins); i += 1 {
		if proof.inputCoins[i].IsEncrypted() {
			return false, errors.New("ConversionProof input should not be encrypted")
		}
		if !bytes.Equal(pubKey, proof.inputCoins[i].GetPublicKey().ToBytesS()) {
			return false, errors.New("ConversionProof inputCoins.publicKey should equal to pubkey")
		}
		//Check the consistency of input coins' commitment
		commitment := proof.inputCoins[i].GetCommitment()
		proof.inputCoins[i].CommitAll()
		if !bytes.Equal(commitment.ToBytesS(), proof.inputCoins[i].GetCommitment().ToBytesS()){
			return false, errors.New("ConversionProof inputCoins.commitment is not correct")
		}
		sumInput += proof.inputCoins[i].GetValue()
	}
	for i := 0; i < len(proof.outputCoins); i += 1 {
		if proof.outputCoins[i].IsEncrypted() {
			return false, errors.New("ConversionProof output should not be encrypted")
		}
		sumOutput += proof.outputCoins[i].GetValue()
	}
	if sumInput != sumOutput+fee {
		return false, errors.New("ConversionProof input should be equal to fee + sum output")
	}
	if len(proof.inputCoins) != len(proof.serialNumberNoPrivacyProof) {
		return false, errors.New("The number of input coins should be equal to the number of proofs")
	}

	for i := 0; i < len(proof.inputCoins); i += 1 {
		valid, err := proof.serialNumberNoPrivacyProof[i].Verify(nil)
		if !valid {
			Logger.Log.Errorf("Verify serial number no privacy proof failed")
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, err)
		}
	}
	return true, nil
}
