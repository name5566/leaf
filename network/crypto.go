package network

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

/**
***************************************** 加密模块 *****************************************
基本要求：
保证传输数据的安全性
保证数据的完整性
能够验证客户端的身份

基本流程：
1. 服务器端(server)和客户端(client)分别生成自己的密钥对
2. server和client分别交换自己的公钥(暂未实现)
3. client生成AES密钥(aesKey)
4. client使用自己的RSA私钥(privateKey)对请求明文数据(params)进行数字签名
5. 将签名加入到请求参数中(附加数据)得到orgData
6. client使用aesKey对orgData进行加密得到密文(data)
7. client使用sever的RSA公钥对aesKey进行加密(encryptKey)
8. 分别将data和encryptKey作为参数传输给服务器端

服务器端进行请求响应时将上面流程反过来即可。
*/

/*
流程补充。
加密目的：
1.验证通信双方的身份
2.保证数据不能被（第三方）查看
3.保证数据没有被（第三方）篡改

加密流程：
首先公钥和密钥在加密的流程中其实牵涉到多个密钥。
其中包括：a.CA机构的公钥和私钥。b.CA认证的Server私钥和证书中的公钥。c.CA认证的Client的私钥和证书中的公钥。d.对称加密的密钥。
有7个之多，但不管有多少密钥，只要是成对出现的，都是一个公钥，一个私钥。这种成对出现的密钥用于非对称加密。如RSA加密。公钥用来加密，私钥用来签名，用法是固定的。
加密和签名的区别在于：
加密的数据针对的是接收者，而不是发送者。即只希望被接收者一个人看到加密之后的明文数据。
签名针对的则是发送者，不是被接收者。即解密之后，任何接收者都可以查看这个被加密的明文，因为这个明文只是为了说明发送者的身份，并没有其他不想让别人知道的数据。
正常情况下需要CA颁发的证书，这个的作用就相当于现实生活中的身份证。
一般的认证流程为：
Server        			         																	Client
																--下发CA的证书-->

																									验证证书的有效性。
																									[在OS或者浏览器中自带的证书中，找到该CA对应的证书（注意是证书链，不是单个证书）
																									比如 用chrome打开www.wosign.com 然后点更多工具->开发者工具->Security我们就能看到证书了。
																									打开证书，在证书的路径这一选项卡中，详细显示了证书链：
																									DigiCert（A）
																									:
																									 WoTrus EV SSL Pro CA（B）
																									  :
																									   www.wosign.com（C）
																									可以看到最下层是CA颁发给这个网站的证书。往上则是颁发证书的中间证书颁发机构。最上面则是受信任的根证书颁发机构。
																									且者三个证书都是可以查看的，主要包括颁发者，使用者，有效期，公钥，指纹。
																									这里先验证证书的有效性。
																									要验证C的有效性，先要用它的上一级CA命名为B_CA的公钥来解密C上的CA（即B_CA）签名，解签后如果明文和上一级的B_CA信息匹配，说明证书C确实是B_CA
																									颁发的。但是还要证明B_CA的有效性，所以要验证证书B的可靠性，同样要用到证书A上的公钥要解密B上的CA（即A_CA）签名，解签后，如果明文和上一级
																									的A_CA信息匹配，则说明B证书确实是A_CA颁发的。这样递归检测父证书，直到出现信任的根证书（证书列表内置于操作系统）。由此可见，除了最底层的证书
																									的公钥没有用到之外，上层的每一层证书的公钥都用来解密下一层证书的证书指纹签名。所以证书是否可靠呢？一句话，根证书可靠，整个证书链就可靠。而根
																									证书要看是否在操作系统或浏览器内置的可信证书内。在的话就可靠。
																									此时，Client就可以确定Server的真实身份了。
																									]


															<--将Client的证书发给Server--

Server验证Client证书的有效性。


使用单向加密（MD5，SHA）生成数据D的特征码X。 得到DX=D+X。
保证数据不被篡改。

用Server私钥将X加密生成数据签名S。			得到DS=D+S。


对DS使用Server的AES密钥AK进行AES对称加密。	得到DS=>New_DS。
对AK使用Client公钥加密生成EAK，并附加在DS上。得到EDS=New_DS+EAK;

															--发送EDS到客户端-->

																									先用Client私钥解密EDS中的EAD。				得到AK。此时数据有New_DS+AK.
																									再用AK解密EDS中的New_DDS.。				得到DS(D+S)。此时数据有.D+S.
																									使用Server公钥解密数据签名S。				得到特征码X。此时数据有D+X.
																									对数据D进行单向加密。						得到Y特征码。
																									和进行对比。如果X==Y,说明数据D没有被篡改。

==============================================================================================
/**但在本Server中没有使用CA颁发的证书，而是预先使用ssl命令获得RSA密钥和公钥.客户端服务器都省略了各自验证身份
的这一步骤。只是保证了数据不被第三方篡改和查看内容，并未实现各自身份的校验。所以服务器流程先各自预备好了，自己的
私钥和公钥，
**/

// AES加密密钥
// 可以做成随机字符串，这一步暂时没完成。
const (
	selfAESKey = "YfWVs4vtcNf6FPFR"
)

const (
	// RSA密钥
	selfRSAPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCzGLI+4WEN+rREDrfbyeZAPfc3PpM9TxVU8PImKk24+Q5WcGYG
gZlFUNTgUcC6n5XJ6QslSOh+BCmdKz5gvSy6AxRS0b+USOvJubHNz46kj3l1MSKS
4qVKwoo0sIhW14bJFCQHoLc9zlAzRDsTqFKb4OnpivgzmJCAz4tbTUp1EwIDAQAB
AoGBAITBIqcPo0SceHEWQ90MnLsz84MkxDmm3FYZQDVgGDqriqAyMr5R5I4H67PX
hbgQQRToxNU/ZO68ISiafGNy9qouT82jfdnA7eB51VlpMcVtHHyp6S32WkbedjjL
JyIKrcUeq4d2LcZwxqHOdxmdDFETgR9MPqSqrwNfVZ/E5FbhAkEA+Js12pfejQLn
Gh6kmJUrlDTIHdPSCiRRhquYIAWN+NyQzBipAJgydh1jnqUnO6f3zmsOhoJfJQr0
CUEO8IlGowJBALhsRAX+APClbhrkwbwX1sdZ2ec6E03hnrNdx740lebOxaKk40W2
PxiTgh4UUC4tlAFzuhGj6MECkHTQe1hHrtECQEC66QLJmEDPCK1cXS79aCNmutRJ
Wt8ZJcES3ME5sQWjKHB720U0W681Z8Le7aAy0+sDJP0Q5QUYHQJr1h/7HlECQFAs
qQnd2fTERnCkoGCwEGxL8IIoajoCaubZTzuuSrizjZHekvs8doOtpPSEqjLZF63l
7K88jbRS9BAEjorbZvECQQCGDdB+6W2Heb7jh2kkk6eFxYqsZt3Af+pTyM3g1Gn2
C9aSRipADTjzsURjI/ltkltoQX4J3GzYgg49Ht9/kHJa
-----END RSA PRIVATE KEY-----`

	// RSA公钥
	selfRSAPublicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCzGLI+4WEN+rREDrfbyeZAPfc3
PpM9TxVU8PImKk24+Q5WcGYGgZlFUNTgUcC6n5XJ6QslSOh+BCmdKz5gvSy6AxRS
0b+USOvJubHNz46kj3l1MSKS4qVKwoo0sIhW14bJFCQHoLc9zlAzRDsTqFKb4Onp
ivgzmJCAz4tbTUp1EwIDAQAB
-----END PUBLIC KEY-----`
)


func init() {

}

func Q1PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText) % blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func Q1PKCS7UnPadding(plantText []byte, blockSize int) []byte {
	length := len(plantText)
	unPadding := int(plantText[length - 1])
	return plantText[:(length - unPadding)]
}

// AES加密
func Q1AESEncrypt(orgData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	orgData = Q1PKCS7Padding(orgData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(key))
	encrypt := make([]byte, len(orgData))
	blockMode.CryptBlocks(encrypt, orgData)

	return encrypt, nil
}

// AES解密
func Q1AESDecrypt(encrypt, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockModel := cipher.NewCBCDecrypter(block, []byte(key))
	decrypt := make([]byte, len(encrypt))
	blockModel.CryptBlocks(decrypt, []byte(encrypt))
	decrypt = Q1PKCS7UnPadding(decrypt, block.BlockSize())

	return decrypt, nil
}

// AES加密（base64编码方式）
func Q1AESEncryptWithBase64(orgData string) (string, error) {
	encrypt, err := Q1AESEncrypt([]byte(orgData), []byte(selfAESKey))
	if err != nil {
		return "", err
	}

	// base64进行编码
	base64Str := base64.StdEncoding.EncodeToString(encrypt)
	return base64Str, nil
}

// AES解密（base64编码方式）
func Q1AESDecryptWithBase64(base64Str string, key []byte) ([]byte, error) {
	encrypt, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}

	decrypt, err := Q1AESDecrypt(encrypt, key)
	if err != nil {
		return nil, err
	}

	return decrypt, nil
}

// 用RSA私钥加密数据的散列码，生成RSA签名
func Q1RSASignDataHash(src []byte, hash crypto.Hash) ([]byte, error) {
	// 解析私钥结构，获得私钥数据。注：pem是采用Base64编码的x.509格式。
	block, _ := pem.Decode([]byte(selfRSAPrivateKey))
	if block == nil {
		return nil, errors.New("private key error!")
	}

	// 从509格式转为PKCS#1格式的私钥结构体
	privateKeyObj, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 生成源数据的散列码
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)

	// 使用PKCS#1格式的私钥生成源数据散列码的签名，并返回这个签名
	return rsa.SignPKCS1v15(rand.Reader, privateKeyObj, hash, hashed)
}

// RSA验签
func Q1RSAVerify(src []byte, sign []byte, hash crypto.Hash, publicKey []byte) error {
	block, _ := pem.Decode(publicKey)	// 将密钥解析成公钥实例
	if block == nil {
		return errors.New("public key error")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)	// 解析pem.Decode（）返回的Block指针实例
	if err != nil {
		return err
	}

	publicKeyObj := pubInterface.(*rsa.PublicKey)

	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.VerifyPKCS1v15(publicKeyObj, hash, hashed, sign)
}

// RSA加密
func Q1RSAEncrypt(orgData []byte, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)	// 将密钥解析成公钥实例
	if block == nil {
		return nil, errors.New("public key error")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)	// 解析pem.Decode（）返回的Block指针实例
	if err != nil {
		return nil, err
	}

	publicKeyObj := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, publicKeyObj, orgData)
}

func Q1RSAEncryptSelfAESKey(publicKey []byte) ([]byte, error) {
	return Q1RSAEncrypt([]byte(selfAESKey), publicKey)
}

// RSA解密
func Q1RSADecrypt(cipherText []byte) ([]byte, error) {
	return Q1RSADecryptWithKey(cipherText, []byte(selfRSAPrivateKey))
}

func Q1RSADecryptWithKey(cipherText []byte, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)	// 将密钥解析成私钥实例
	if block == nil {
		return nil, errors.New("private key error!")
	}

	privateKeyObj, err := x509.ParsePKCS1PrivateKey(block.Bytes)	// 解析pem.Decode（）返回的Block指针实例
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privateKeyObj, cipherText)
}