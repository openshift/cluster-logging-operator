package misc

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
)

// corruptDiskBufferPayload corrupts protobuf payloads in a Vector disk buffer
// data file while keeping CRC32 checksums valid, causing ReaderError::Decode
// on startup seek. The last record is left intact (writer validates it).
// Returns the number of corrupted records.
func corruptDiskBufferPayload(datFilePath string) (int, error) {
	data, err := os.ReadFile(datFilePath)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	type record struct{ start, length int }
	var records []record
	offset := 0
	for offset+8 < len(data) {
		recLen := int(binary.BigEndian.Uint64(data[offset : offset+8]))
		recStart := offset + 8
		if recLen < 32 || recStart+recLen > len(data) {
			break
		}
		records = append(records, record{recStart, recLen})
		offset = recStart + recLen
	}

	if len(records) < 2 {
		return 0, fmt.Errorf("need at least 2 records to corrupt (found %d); last record must stay intact for writer validation", len(records))
	}

	corruptUpTo := len(records) - 1
	corrupted := 0
	for i := 0; i < corruptUpTo; i++ {
		r := records[i]
		if err := corruptRecordPayload(data[r.start : r.start+r.length]); err != nil {
			return 0, fmt.Errorf("corrupt record %d: %w", i, err)
		}
		corrupted++
	}

	if err := os.WriteFile(datFilePath, data, 0o640); err != nil {
		return 0, err
	}
	return corrupted, nil
}

// corruptRecordPayload XOR-flips the payload bytes of a single rkyv-serialized
// record and recalculates the CRC32 checksum so the record passes checksum
// validation but fails protobuf decode.
func corruptRecordPayload(recordData []byte) error {
	if len(recordData) < 32 {
		return fmt.Errorf("record too small: %d bytes", len(recordData))
	}

	// ArchivedRecord layout (repr(C), little-endian, 32 bytes at end of record):
	//   +0: checksum u32, +4: padding, +8: id u64,
	//   +16: metadata u32, +20: rel_ptr i32, +24: payload_len u32, +28: padding
	structOff := len(recordData) - 32

	id := binary.LittleEndian.Uint64(recordData[structOff+8 : structOff+16])
	metadata := binary.LittleEndian.Uint32(recordData[structOff+16 : structOff+20])
	relPtr := int32(binary.LittleEndian.Uint32(recordData[structOff+20 : structOff+24]))
	payloadLen := binary.LittleEndian.Uint32(recordData[structOff+24 : structOff+28])

	payloadStart := int(structOff+20) + int(relPtr)
	if payloadStart < 0 || payloadStart+int(payloadLen) > len(recordData) {
		return fmt.Errorf("payload out of bounds: start=%d len=%d record=%d", payloadStart, payloadLen, len(recordData))
	}

	payload := recordData[payloadStart : payloadStart+int(payloadLen)]
	for i := range payload {
		payload[i] ^= 0xFF
	}

	// Recalculate CRC32-IEEE: BE(id) + BE(metadata) + payload
	h := crc32.NewIEEE()
	idBE := make([]byte, 8)
	binary.BigEndian.PutUint64(idBE, id)
	_, _ = h.Write(idBE)
	metaBE := make([]byte, 4)
	binary.BigEndian.PutUint32(metaBE, metadata)
	_, _ = h.Write(metaBE)
	_, _ = h.Write(payload)
	binary.LittleEndian.PutUint32(recordData[structOff:structOff+4], h.Sum32())

	return nil
}
