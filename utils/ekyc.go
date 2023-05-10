package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	base64 "encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"os"
)

/*
public String genSignature(String data) {
	String signature = null;
	try {
		byte[] arrPk = Base64.decode(new String(FileUtil.readBytesFromFile("prv.rsa")));
		PrivateKey prvKey = KeyFactory.getInstance("RSA").generatePrivate(new PKCS8EncodedKeySpec(arrPk));
		signature = RSA.genSignature(prvKey, data);
		System.out.println("chữ ký: " + signature);
	} catch (Exception e) {
		e.printStackTrace();
	}
	return signature;
 }
*/

func EkycGenSignature(dataStr string) (error, string) {
	signature := ""

	// Get private key from local
	// test local
	//f, errPrv := os.Open("../prv.rsa")
	//staging , prod
	f, errPrv := os.Open("prv.rsa")
	if errPrv != nil {
		log.Println("EkycGenSignature errPrv", errPrv.Error())
		return errPrv, signature
	}
	defer f.Close()

	byteRsaPrivateValue, errRA := ioutil.ReadAll(f)

	if errRA != nil {
		log.Println("EkycGenSignature errRA", errRA.Error())
		return errRA, signature
	}

	base64EncodeRsaPrivate, errDecodeByteRsa := base64.StdEncoding.DecodeString(string(byteRsaPrivateValue))
	if errDecodeByteRsa != nil {
		log.Println("EkycGenSignature errDecodeByteRsa", errDecodeByteRsa.Error())
		return errDecodeByteRsa, signature
	}

	privateKey, errPkey := BytesToPrivateKey(base64EncodeRsaPrivate)
	if errPkey != nil {
		return errPkey, signature
	}

	// log.Println("EkycGenSignature privateKey ok")

	msgHash := sha256.New()
	_, errmhash := msgHash.Write([]byte(dataStr))
	if errmhash != nil {
		return errmhash, signature
	}
	msgHashSum := msgHash.Sum(nil)

	// signatureByte, errSig := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, rand.Reader, privateKey, nil)
	// if errSig != nil {
	// 	return errSig, signature
	// }

	signatureByte, errSig := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, msgHashSum)
	if errSig != nil {
		return errSig, signature
	}

	//base64 encode
	signature = base64.StdEncoding.EncodeToString(signatureByte)

	return nil, signature
}

// BytesToPublicKey bytes to public key
func BytesToPublicKey(pub []byte) (*rsa.PublicKey, error) {
	// block, _ := pem.Decode(pub)
	// if block == nil {
	// 	return nil, errors.New("block is nill")
	// }
	// enc := x509.IsEncryptedPEMBlock(block)
	// b := block.Bytes
	// var err error
	// if enc {
	// 	log.Println("is encrypted pem block")
	// 	b, err = x509.DecryptPEMBlock(block, nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	ifc, err := x509.ParsePKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		// log.Error("not ok")
		return nil, errors.New("not ok")
	}
	return key, nil
}

// BytesToPrivateKey bytes to private key
func BytesToPrivateKey(priv []byte) (*rsa.PrivateKey, error) {
	// block, _ := pem.Decode(priv)
	// enc := x509.IsEncryptedPEMBlock(block)
	// b := block.Bytes
	// var err error
	// if enc {
	// 	log.Println("is encrypted pem block")
	// 	b, err = x509.DecryptPEMBlock(block, nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	ifc, err1 := x509.ParsePKCS8PrivateKey(priv)
	if err1 != nil {
		log.Println("BytesToPrivateKey err1", err1.Error())
	}

	key, ok := ifc.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("BytesToPrivateKey not ok")
	}

	// key, err := x509.ParsePKCS1PrivateKey(priv)
	// if err != nil {
	// 	log.Println("BytesToPrivateKey", err.Error())
	// 	return nil, err
	// }
	return key, nil
}
