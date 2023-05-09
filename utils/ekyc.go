package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	base64 "encoding/base64"
	"encoding/pem"
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

	privateKey, errPkey := BytesToPrivateKey(byteRsaPrivateValue)
	if errPkey != nil {
		return errPkey, signature
	}

	msgHash := sha256.New()
	_, errmhash := msgHash.Write([]byte(dataStr))
	if errmhash != nil {
		return errmhash, signature
	}
	msgHashSum := msgHash.Sum(nil)

	signatureByte, errSig := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, msgHashSum, nil)
	if errSig != nil {
		return errSig, signature
	}

	//base64 encode
	signature = base64.StdEncoding.EncodeToString(signatureByte)

	return nil, signature
}

/*
func VerifyVNPayRsa(phone, bankCode, bankName, token string) error {
	msgStr := phone + bankCode + bankName
	log.Println("VerifyVNPayRsa msgStr", msgStr)
	hashed := sha256.Sum256([]byte(msgStr))
	// msg := []byte(msgStr)

	var err error

	// Base 64 token
	byteTokenDec, errTokenDc := base64.StdEncoding.DecodeString(token)
	if errTokenDc != nil {
		log.Println("verifyVNPayRsa errTokenDc ", errTokenDc.Error())
		return errTokenDc
	}
	signature := byteTokenDec // == token

	// Get public key from local
	f, errF := os.Open("vnp_public.rsa")
	if errF != nil {
		log.Println("verifyVNPayRsa errF", errF.Error())
		return errF
	}
	defer f.Close()

	byteRsaPublicValue, errRA := ioutil.ReadAll(f)

	if errRA != nil {
		log.Println("verifyVNPayRsa errRA", errRA.Error())
		return errRA
	}

	rsaPublicValueStr := string(byteRsaPublicValue)

	byteRsaPublicValueDec, errDc := base64.StdEncoding.DecodeString(rsaPublicValueStr)
	if errDc != nil {
		log.Println("verifyVNPayRsa errDc ", errDc.Error())
		return errDc
	}

	// public key
	publicKey, errPK := BytesToPublicKey(byteRsaPublicValueDec)
	if errPK != nil {
		log.Println("verifyVNPayRsa errPK", errPK.Error())
		return errPK
	}

	// Verify
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		fmt.Println("verifyVNPayRsa could not verify signature: ", err.Error())
		return err
	}

	// signature is valid
	fmt.Println("verifyVNPayRsa signature verified")
	return nil
}
*/
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
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil, err
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil, err
	}
	return key, nil
}
