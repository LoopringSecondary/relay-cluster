package accessor

import "math/big"

type BlockIterator struct {
	startNumber   *big.Int
	endNumber     *big.Int
	currentNumber *big.Int
	ethClient     *ethNodeAccessor
	withTxData    bool
	confirms      uint64
}
