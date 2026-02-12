package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"

	"github.com/top-system/light-admin/pkg/str"
)

// MD5 哈希值
func MD5(s string) string {
	sum := md5.Sum(str.S(s).Bytes())
	return hex.EncodeToString(sum[:])
}

// SHA1 哈希值
func SHA1(s string) string {
	sum := sha1.Sum(str.S(s).Bytes())
	return hex.EncodeToString(sum[:])
}

// SHA256 哈希值
func SHA256(s string) string {
	sum := sha256.Sum256(str.S(s).Bytes())
	return hex.EncodeToString(sum[:])
}

// BcryptHash 使用 bcrypt 对密码进行哈希
func BcryptHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// BcryptCheck 验证密码是否与 bcrypt 哈希匹配
func BcryptCheck(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// IsBcryptHash 判断密码是否为 bcrypt 格式
func IsBcryptHash(s string) bool {
	return len(s) == 60 && s[0] == '$'
}
