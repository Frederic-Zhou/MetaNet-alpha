package network

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestDataFormat(t *testing.T) {

	var group, index, check []byte

	data := []byte("1")

	groupNum := CRC32(data)
	binary.LittleEndian.PutUint32(group, groupNum)

	indexNum := uint32(123)
	binary.LittleEndian.PutUint32(index, indexNum)

	check = MD5bytes(data)

	data = append(group, data...)
	data = append(index, data...)
	data = append(check, data...)

	fmt.Println(groupNum, data, len(group), len(index), len(check), len(data))

}
