package idx

import (
	"github.com/Fantom-foundation/go-lachesis/src/common/bigendian"
)

type (
	// SuperFrame numeration.
	SuperFrame uint32

	// Event numeration.
	Event uint32

	// Txn numeration.
	Txn uint32

	// Block numeration.
	Block uint64

	// Lamport numeration.
	Lamport uint32

	// Frame numeration.
	Frame uint32
)

// Bytes gets the byte representation of the index.
func (sf SuperFrame) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(sf))
}

// Bytes gets the byte representation of the index.
func (e Event) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(e))
}

// Bytes gets the byte representation of the index.
func (t Txn) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(t))
}

// Bytes gets the byte representation of the index.
func (b Block) Bytes() []byte {
	return bigendian.Int64ToBytes(uint64(b))
}

// Bytes gets the byte representation of the index.
func (l Lamport) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(l))
}

// Bytes gets the byte representation of the index.
func (f Frame) Bytes() []byte {
	return bigendian.Int32ToBytes(uint32(f))
}

// BytesToEvent converts bytes to event index.
func BytesToEvent(b []byte) Event {
	return Event(bigendian.BytesToInt32(b))
}

// BytesToTxn converts bytes to transaction index.
func BytesToTxn(b []byte) Txn {
	return Txn(bigendian.BytesToInt32(b))
}

// BytesToBlock converts bytes to block index.
func BytesToBlock(b []byte) Block {
	return Block(bigendian.BytesToInt64(b))
}

// BytesToLamport converts bytes to block index.
func BytesToLamport(b []byte) Lamport {
	return Lamport(bigendian.BytesToInt32(b))
}

// BytesToFrame converts bytes to block index.
func BytesToFrame(b []byte) Frame {
	return Frame(bigendian.BytesToInt32(b))
}

// MaxLamport return max value
func MaxLamport(x, y Lamport) Lamport {
	if x > y {
		return x
	}
	return y
}