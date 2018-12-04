package privacy

const (
	CompressedPointSize      = 33
	PointCompressed     byte = 0x2

	SK            = byte(0x00)
	VALUE         = byte(0x01)
	SND           = byte(0x02)
	SHARDID				= byte(0x03)
	RAND          = byte(0x04)
	FULL          = byte(0x05)

	CMRingSize      = 8 // 2^3
	CMRingSizeExp   = 3

	ComInputOpeningsProofSize  = CompressedPointSize*2+BigIntSize*5
	//ComOutputOpeningsProofSize =
	// OneOfManyProofSize              = 0
	EqualityOfCommittedValProofSize = 30
	//ComMultiRangeProofSize          = 0
	ComZeroProofSize                = 99
	ComZeroOneProofSize             = 0
	// EllipticPointCompressSize       = 33
	CommitmentSize = 0
	BigIntSize     = 32
)
