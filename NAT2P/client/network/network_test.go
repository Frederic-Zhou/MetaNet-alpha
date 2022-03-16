package network

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/client/utils"
)

func TestDataFormat(t *testing.T) {

	var id = make([]byte, 4)
	var seq = make([]byte, 4)
	var check = make([]byte, 4)

	data := []byte("1")

	idNum := uint32(123)
	binary.LittleEndian.PutUint32(id, idNum)

	indexNum := uint32(123)
	binary.LittleEndian.PutUint32(seq, indexNum)

	checkNum := utils.CRC32(data)
	binary.LittleEndian.PutUint32(check, checkNum)

	data = append(id, data...)
	data = append(seq, data...)
	data = append(check, data...)

	fmt.Println(data, len(id), len(seq), len(check), len(data))

}
