package instruction

import (
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

//UnstakeInstruction : Hold and verify data for unstake action
type UnstakeInstruction struct {
	PublicKeys []string
}

//NewUnstakeInstructionWithValue : Constructor with value
func NewUnstakeInstructionWithValue(publicKeys []string) *UnstakeInstruction {
	return &UnstakeInstruction{PublicKeys: publicKeys}
}

//NewUnstakeInstruction : Default constructor
func NewUnstakeInstruction() *UnstakeInstruction {
	return &UnstakeInstruction{}
}

//GetType : Get type of unstake instruction
func (unstakeIns *UnstakeInstruction) GetType() string {
	return UNSTAKE_ACTION
}

//ToString : Convert class to string
func (unstakeIns *UnstakeInstruction) ToString() []string {
	unstakeInstructionStr := []string{UNSTAKE_ACTION}
	unstakeInstructionStr = append(unstakeInstructionStr, strings.Join(unstakeIns.PublicKeys, SPLITTER))
	return unstakeInstructionStr
}

//ValidateAndImportUnstakeInstructionFromString : Validate and import unstake instruction from string
func ValidateAndImportUnstakeInstructionFromString(instruction []string) (*UnstakeInstruction, error) {
	if err := ValidateUnstakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportUnstakeInstructionFromString(instruction), nil
}

//ImportUnstakeInstructionFromString : Import unstake instruction from string
func ImportUnstakeInstructionFromString(instruction []string) *UnstakeInstruction {
	unstakeInstruction := NewUnstakeInstruction()
	if len(instruction[1]) > 0 {
		publicKeys := strings.Split(instruction[1], SPLITTER)
		unstakeInstruction.PublicKeys = publicKeys
	}
	return unstakeInstruction
}

//ValidateUnstakeInstructionSanity : Validate unstake instruction data type
func ValidateUnstakeInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != UNSTAKE_ACTION {
		return fmt.Errorf("invalid stop auto stake action, %+v", instruction)
	}
	return nil
}

//InsertIntoStateDB : insert unstake instruction to statedb
func (unstakeIns *UnstakeInstruction) InsertIntoStateDB(sDB *statedb.StateDB) error {
	return nil
}