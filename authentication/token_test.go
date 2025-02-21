package auth

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/xerrors"
)

func TestToken(t *testing.T) {
	claims := &Claims{
		Type:         1,
		IsRegistered: true,
		ChainID:      985,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Audience:  Domain,
			Issuer:    Domain,
			Subject:   "0x2da24D5cc8F180727A588065D1cF39B2417B74c5",
		},
	}

	key := []byte("memo.io")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(accessToken)

	t.Log(time.Now().Unix())
}

var baseUrl = "https://test-xs-api.memolabs.net/v1"
var globalPrivateKey = "593b0434faac6e71a8d55545a56653d3f0cbe309b174735ec09d7a4ac05ff75f"

func TestLogin(t *testing.T) {
	privateKey, err := crypto.HexToECDSA(globalPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	text, err := challenge(address)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(text)

	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(text), text)))
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	sig := hexutil.Encode(signature)

	err = login(text, sig)
	if err != nil {
		t.Fatal(err)
	}
}

func challenge(address string) (string, error) {
	client := &http.Client{Timeout: time.Minute}
	url := baseUrl + "/challenge"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	params := req.URL.Query()
	params.Add("address", address)
	req.URL.RawQuery = params.Encode()
	req.Header.Set("Origin", "https://memo.io")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", xerrors.Errorf("Respond code[%d]: %s", res.StatusCode, string(body))
	}

	return string(body), nil
}

func login(message, signature string) error {
	client := &http.Client{Timeout: time.Minute}
	url := baseUrl + "/login"

	var payload = make(map[string]string)
	payload["message"] = message
	payload["signature"] = signature

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return xerrors.Errorf("Respond code[%d]: %s", res.StatusCode, string(body))
	}

	fmt.Println(string(body))

	return nil
}
