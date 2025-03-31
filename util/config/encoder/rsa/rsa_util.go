package rsa

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"strings"
)

// 生成RSA密钥对
func generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Fatal(err)
	}
	return privateKey, &privateKey.PublicKey
}

// GetPrivateKeyFromPEM 从PEM格式的私钥字符串获取*rsa.PrivateKey
func GetPrivateKeyFromPEM(privateKey string) (*rsa.PrivateKey, error) {
	// 私钥通常是PEM格式，以"-----BEGIN PRIVATE KEY-----"开头
	privateKeyBytes := []byte(privateKey)
	// 去掉PEM格式中的换行符和头尾信息
	privateKeyBytes = []byte(strings.TrimSpace(string(privateKeyBytes)))
	privateKeyBytes = []byte(strings.Join(strings.Split(string(privateKeyBytes), "\n"), ""))

	// 使用base64解码
	decodedKey, err := base64.StdEncoding.DecodeString(string(privateKeyBytes))
	if err != nil {
		return nil, err
	}

	// 使用x509解析私钥
	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(decodedKey)
	if err != nil {
		return nil, err
	}

	privateKeyRSA, ok := privateKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("failed to parse private key as RSA private key")
	}

	return privateKeyRSA, nil
}

// GetPublicKeyFromBase64 将原始公钥字符串转换为 PEM 格式
func GetPublicKeyFromBase64(base64PublicKey string) (*rsa.PublicKey, error) {

	publicKeyBytes := []byte(base64PublicKey)
	// 去掉PEM格式中的换行符和头尾信息
	publicKeyBytes = []byte(strings.TrimSpace(string(publicKeyBytes)))
	publicKeyBytes = []byte(strings.Join(strings.Split(string(publicKeyBytes), "\n"), ""))

	// 使用base64解码
	block, err := base64.StdEncoding.DecodeString(string(publicKeyBytes))
	if err != nil {
		return nil, err
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block)
	if err != nil {
		return nil, err
	}

	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return publicKey, nil
}

// Encrypt 使用公钥加密数据
func Encrypt(data string, publicKeyStr string) (string, error) {
	publicKey, err := GetPublicKeyFromBase64(publicKeyStr)
	if err != nil {
		return "", err
	}
	// 使用 PKCS1v15 加密
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(data))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// Decrypt RSA解密
func Decrypt(encryptedData string, privateKeyStr string) (string, error) {
	privateKey, err := GetPrivateKeyFromPEM(privateKeyStr)
	if err != nil {
		return "", err
	}
	decodedData, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}
	//decryptedData, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privateKey, decodedData, nil)
	//if err != nil {
	//	return "", err
	//}
	decryptedData, err := rsa.DecryptPKCS1v15(nil, privateKey, decodedData)
	if err != nil {
		return "", err
	}
	return string(decryptedData), nil
}

// RSA签名
func sign(data string, privateKey *rsa.PrivateKey) (string, error) {
	h := sha1.New()
	h.Write([]byte(data))
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, hashed)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// RSA验签
func verify(data string, signature string, publicKey *rsa.PublicKey) bool {
	h := sha1.New()
	h.Write([]byte(data))
	hashed := h.Sum(nil)

	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, hashed, signatureBytes)
	if err != nil {
		return false
	}
	return true
}

// CalculateMD5 计算字符串的MD5哈希值
func CalculateMD5(input string) string {
	hash := md5.Sum([]byte(input))
	// 将[]byte转换为十六进制字符串
	return hex.EncodeToString(hash[:])
}

func main() {
	//calculateMD51 := CalculateMD5(strconv.FormatInt(1045832034, 10) + strconv.FormatInt(16808, 10))
	//calculateMD52 := CalculateMD5(strconv.FormatInt(712459609, 10) + strconv.FormatInt(1045832034, 10))
	//log.Println(calculateMD51)
	//log.Println(calculateMD52)
	// 示例私钥字符串，通常来源于文件或数据库
	publicKey := "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCcbsc7X1y3xn7BvBL/bDCOqfngytBvn8mpvgZkOtEMcCLPmZu145BYn01OuZ7HQdb6tK7n7d5/y57avzZyJiAsVGR346FaU2AmvoNieoJ96K6GlnKHo8CgAyCwF3dVxp6TfIUHwGs4Z65m73XyXvrbKWW+BInKK3XoG/qbdxdbpQIDAQAB"
	privateKey := "MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAJxuxztfXLfGfsG8Ev9sMI6p+eDK0G+fyam+BmQ60QxwIs+Zm7XjkFifTU65nsdB1vq0ruft3n/Lntq/NnImICxUZHfjoVpTYCa+g2J6gn3oroaWcoejwKADILAXd1XGnpN8hQfAazhnrmbvdfJe+tspZb4Eicordegb+pt3F1ulAgMBAAECgYAg7r1oxXG6isJCvPpu5XLvhd9CMNBiv4vv/T5ROYSrDqx1cgwy5Z6M2bSnvzIrFrRQgVtVHmG6G77spFas/1PES+evxGOV5AlXbyck2EwsRIKkIVOkUTAZwUDobF1z9eawDy54W1ko7uRIIDZIMJldSETSWfaKjBs5fwp5jxqb3QJBAOzGq3iVwYEiukyj50NcmKg63M2OEcO21urPTRrePd4zxJG4TrBapB3UT7Px9/InKkPtpdchiEvucdQfuGft3DMCQQCpIjFayOftXNi9YU8aQghYPZ6wiMT6LJOmlWCWjJTZW3bXFbBTqzDaQnYAQzuz9KC98g/Zq++D33TBF6SE2hDHAkEAwF7RZdFWPBL5BdeMx1/t75CTYLZynG5qwq/WV2QFJAkvRa1W0VVzTYD3mJ2Y8zb60eG9AcKOuBJsjQmQi2/nnQJALnycbiR8QqxbUioV0NTHcGF3ZXQiF9T6vDWgd6CqJNfT4Sgv779EzSipQEc6eKrLJ4oJuz1btrZLY+s4p9877wJBAMRM/E56TUPMedcOo7krWi/Rc4jfNWb0FFErNXJO6EEX+LmneUXF+zYqvGWjnC1SxqkYw7rCo+QwHu4lL5CEjMM="

	encrypt, err := Encrypt("sdfasdfSADFASADSFASDFSSD的方式12", publicKey)
	if err != nil {
		log.Printf("err:%v", err)
	}
	log.Printf("encrypt:%v", encrypt)

	decrypt, err := Decrypt(encrypt, privateKey)

	if err != nil {
		log.Printf("decrypt err:%v", err)
	}
	log.Printf("decrypt:%v", decrypt)
}
