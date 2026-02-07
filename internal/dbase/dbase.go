package dbase

import (
	"encoding/binary"
	"io"
	"strings"
)

type Dbase struct {
	Header Header
	Fields []FieldDescriptor
}

func Parse(r io.Reader) (*Dbase, error) {
	db := &Dbase{}
	if err := binary.Read(r, binary.LittleEndian, &db.Header); err != nil {
		return nil, err
	}

	count := int(db.Header.HeaderLength-32-1) / 32
	db.Fields = make([]FieldDescriptor, 0, count)

	for range count {
		var fd FieldDescriptor
		if err := binary.Read(r, binary.LittleEndian, &fd); err != nil {
			return nil, err
		}
		db.Fields = append(db.Fields, fd)
	}

	var terminator [1]byte
	if _, err := r.Read(terminator[:]); err != nil {
		// Missing terminator byte, soft failure.
	}

	return db, nil
}

type Header struct {
	Version               uint8
	YY                    uint8
	MM                    uint8
	DD                    uint8
	RecordCount           uint32
	HeaderLength          uint16
	RecordLength          uint16
	Reserved1             [2]byte
	IncompleteTransaction uint8
	EncryptionFlag        uint8
	FreeRecordThread      [4]byte
	Reserved2             [8]byte
	MDXFlag               uint8
	LanguageDriver        uint8
	Reserved3             [2]byte
}

type FieldDescriptor struct {
	Name           [11]byte
	Type           uint8
	Address        [4]byte
	Length         uint8
	DecimalCount   uint8
	Reserved1      [2]byte
	WorkAreaId     uint8
	Reserved2      [2]byte
	SetFieldsFlag  uint8
	Reserved3      [7]byte
	IndexFieldFlag uint8
}

func (fd *FieldDescriptor) GetName() string {
	s := strings.TrimRight(string(fd.Name[:]), "\x00")
	return strings.TrimSpace(s)
}

func (fd *FieldDescriptor) Read(r io.Reader) (string, error) {
	data := make([]byte, fd.Length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
