package envelopeenc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const marshalVersion = 1

/*	Format:

	uvarint  version (always 1)
	uvarint  length of EncryptedContent
	[]byte   EncryptedContent
	uvarint  amount of key slots

	for each key slot

		uvarint  length of KekId
		string   KekId
		uvarint  length of DekEncrypted
		[]byte   DekEncrypted

	NOTE: I would've gladly used Protobuf, but looks like you need a metric shit-ton of
	      imported (even runtime) code to use it.
*/
func (e *Envelope) Marshal() ([]byte, error) {
	out := bytes.Buffer{}

	var err error

	writeBytes := func(data []byte) {
		if err != nil {
			return
		}

		_, err = out.Write(data)
	}

	writeUvarint := func(num uint64) {
		buf := make([]byte, binary.MaxVarintLen64)
		writeBytes(buf[0:binary.PutUvarint(buf, num)])
	}

	writeUvarint(uint64(marshalVersion))

	writeUvarint(uint64(len(e.EncryptedContent)))

	writeBytes(e.EncryptedContent)

	writeUvarint(uint64(len(e.KeySlots)))

	for _, keySlot := range e.KeySlots {
		writeUvarint(uint64(len(keySlot.KekId)))

		writeBytes([]byte(keySlot.KekId))

		writeUvarint(uint64(len(keySlot.DekEncrypted)))

		writeBytes(keySlot.DekEncrypted)
	}

	return out.Bytes(), err
}

func Unmarshal(buf []byte) (*Envelope, error) {
	bufReader := bytes.NewBuffer(buf)

	version, err := binary.ReadUvarint(bufReader)
	if err != nil {
		return nil, err
	}

	if version != marshalVersion {
		return nil, fmt.Errorf(
			"unexpected version: %d, expected %d",
			version,
			marshalVersion)
	}

	readByteSlice := func() ([]byte, error) {
		byteSliceLen, err := binary.ReadUvarint(bufReader)
		if err != nil {
			return nil, err
		}

		byteSlice := make([]byte, byteSliceLen)

		if _, err := io.ReadFull(bufReader, byteSlice); err != nil {
			return nil, err
		}

		return byteSlice, nil
	}

	encryptedContent, err := readByteSlice()
	if err != nil {
		return nil, err
	}

	keySlots := []envelopeKeySlot{}

	lenKeySlots, err := binary.ReadUvarint(bufReader)
	if err != nil {
		return nil, err
	}

	for i := uint64(0); i < lenKeySlots; i++ {
		kekId, err := readByteSlice()
		if err != nil {
			return nil, err
		}

		dekEncrypted, err := readByteSlice()
		if err != nil {
			return nil, err
		}

		keySlots = append(keySlots, envelopeKeySlot{
			KekId:        string(kekId),
			DekEncrypted: dekEncrypted,
		})
	}

	return &Envelope{
		EncryptedContent: encryptedContent,
		KeySlots:         keySlots,
	}, nil
}
