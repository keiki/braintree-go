package braintree

import (
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strings"
)

type SignatureError struct {
	message string
}

func (i SignatureError) Error() string {
	if i.message == "" {
		return "Invalid Signature"
	}
	return i.message
}

func newHmacer(bt *Braintree) hmacer {
	return hmacer{bt}
}

type hmacer struct {
	*Braintree
}

func (h hmacer) verifySignature(signature, payload string) (bool, error) {
	signature, err := h.parseSignature(signature)
	if err != nil {
		return false, err
	}
	expectedSignature, err := h.hmac(payload)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(expectedSignature), []byte(signature)), nil
}

func (h hmacer) parseSignature(signatureKeyPair string) (string, error) {
	pairs := strings.Split(signatureKeyPair, "&")

	pubKeyToSignature := map[string]string{}
	for _, pair := range pairs {
		keyAndSig := strings.Split(pair, "|")
		if len(keyAndSig) != 2 {
			return "", SignatureError{"Signature-key pair is in incorrect format"}
		}
		pubKeyToSignature[keyAndSig[0]] = keyAndSig[1]
	}

	sig, ok := pubKeyToSignature[h.Braintree.PublicKey]
	if !ok {
		return "", SignatureError{"Signature-key pair contains the wrong public key!"}
	}

	return sig, nil
}

func (h hmacer) hmac(payload string) (string, error) {
	s := sha1.New()
	_, err := io.WriteString(s, h.PrivateKey)
	if err != nil {
		return "", errors.New("Could not write private key to SHA1")
	}
	mac := hmac.New(sha1.New, s.Sum(nil))
	_, err = mac.Write([]byte(payload))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", mac.Sum(nil)), nil
}
