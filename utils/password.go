package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"io"
	mt "math/rand"
	"strconv"
	"strings"
)

const (
	saltSize           = 16
	delimiter          = "|"
	stretchingPassword = 500
	saltLocalSecret    = "wWnN&^bnmIIIEbW**WL"
)

func PasswordHash(pass string) (string, error) {
	saltSecret, err := saltSecret()
	if err != nil {
		return "", err
	}

	salt, err := salt(saltLocalSecret + saltSecret)
	if err != nil {
		return "", err
	}

	interation := randInt(1, 20)
	hash, err := hash(pass, saltSecret, salt, int64(interation))
	if err != nil {
		return "", err
	}
	interationStr := strconv.Itoa(interation)
	password := saltSecret + delimiter + interationStr + delimiter + hash + delimiter + salt

	return password, nil

}

// 验证密码，hashing是数据库中的密码，pass是请求中的未加密的密码
func PasswordVerify(hashing string, pass string) (bool, error) {
	// 解析哈希字符串，获取迭代次数、盐值等数据
	data := trimSaltHash(hashing)

	// 将数据库中密码存储的迭代次数从字符串类型转换成十进制int64类型
	interation, _ := strconv.ParseInt(data["interation_string"], 10, 64)

	// 将pass根据指定的盐值进行过加密，得到has
	has, err := hash(pass, data["salt_secret"], data["salt"], int64(interation))
	if err != nil {
		return false, err
	}

	// 最后将各相关信息拼接在一起，与数据库中的加密密码进行比较
	if (data["salt_secret"] + delimiter + data["interation_string"] + delimiter + has + delimiter + data["salt"]) == hashing {
		// 如果相同，则证明密码验证成功
		return true, nil
	}
	return false, nil
}

// 密码哈希函数的实现， pass是请求携带的未加密的密码
// saltSecret和salt是数据库中密码加密的信息，iteration是迭代次数
func hash(pass string, saltSecret string, salt string, interation int64) (string, error) {
	// 构建包含密码和盐值的字符串
	var passSalt = saltSecret + pass + salt + saltSecret + pass + salt + pass + pass + salt
	var i int

	// 初始化哈希函数
	hashPass := saltLocalSecret
	hashStart := sha512.New()
	hashCenter := sha256.New()
	hashOutput := sha256.New224()

	// 进行密码哈希的三个阶段循环
	i = 0
	// 1. 将密码字符串拉伸
	for i <= stretchingPassword {
		i = i + 1
		hashStart.Write([]byte(passSalt + hashPass))
		// 使用 hashStart.Sum(nil) 获取哈希结果的摘要值
		// 而后将其转换为16进制字符串的哈希值
		hashPass = hex.EncodeToString(hashStart.Sum(nil))
	}

	i = 0
	// 2. 将哈希值复制到到iteration倍
	for int64(i) <= interation {
		i = i + 1
		hashPass = hashPass + hashPass
	}

	i = 0
	// 3. 再进行一次字符串拉伸
	for i <= stretchingPassword {
		i = i + 1
		hashCenter.Write([]byte(hashPass + saltSecret))
		hashPass = hex.EncodeToString(hashCenter.Sum(nil))
	}
	// 最后将哈希字符串加上盐值
	hashOutput.Write([]byte(hashPass + saltLocalSecret))
	hashPass = hex.EncodeToString(hashOutput.Sum(nil))

	return hashPass, nil
}

func trimSaltHash(hash string) map[string]string {
	// 使用分隔符拆分哈希字符串
	str := strings.Split(hash, delimiter)

	// 将拆分后的部分存储到映射中
	return map[string]string{
		"salt_secret":       str[0],
		"interation_string": str[1],
		"hash":              str[2],
		"salt":              str[3],
	}
}
func salt(secret string) (string, error) {
	buf := make([]byte, saltSize, saltSize+md5.Size)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}

	hash := md5.New()
	hash.Write(buf)
	hash.Write([]byte(secret))
	return hex.EncodeToString(hash.Sum(buf)), nil
}

func saltSecret() (string, error) {
	rb := make([]byte, randInt(10, 100))
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(rb), nil
}

func randInt(min int, max int) int {
	return min + mt.Intn(max-min)
}
