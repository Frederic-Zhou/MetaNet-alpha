package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"hash/crc32"

	"github.com/google/uuid"
)

// 生成md5
func MD5bytes(data []byte) []byte {
	c := md5.New()
	c.Write(data)
	return c.Sum(nil)
	// hex.EncodeToString(c.Sum(nil))
}

//生成sha1
func SHA1(data []byte) []byte {
	c := sha1.New()
	c.Write(data)
	return c.Sum(nil)
	// hex.EncodeToString(c.Sum(nil))
}

func CRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

func UUID() string {
	return uuid.New().String()
}
