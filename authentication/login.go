package auth

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spruceid/siwe-go"
	"golang.org/x/xerrors"
)

var purposeStatement = "The message is only used for login"

type EIP4361Request struct {
	EIP191Message string `json:"message,omitempty"`
	Signature     string `json:"signature,omitempty"`

	// // used for registe
	// Recommender string `json:"recommender,omitempty"`
	// Source      string `json:"source,omitempty"`
}

var (
	ErrNullToken      = xerrors.New("Token is Null, not found in `Authorization: Bearer ` header")
	ErrValidToken     = xerrors.New("Invalid token")
	ErrValidTokenType = xerrors.New("InValid token type")

	// ChainID = 985
	Version = 1

	JWTKey []byte

	DidToken     = 0
	AccessToken  = 1
	RefreshToken = 2

	LensMod = 0x10
	EthMod  = 0x11
)

func InitAuthConfig(jwtKey string, domain string, url string) {
	var err error
	JWTKey, err = hex.DecodeString(jwtKey)
	if err != nil {
		JWTKey = []byte("memo.io")
	}
}

type AuthController struct {
	*NonceManager
	jwtKey []byte
}

func NewAuthController(jwtKey string) (*AuthController, error) {
	JWTKey, err := hex.DecodeString(jwtKey)
	if err != nil {
		JWTKey = []byte("memo.io")
	}
	return &AuthController{
		NonceManager: NewNonceManager(30*int64(time.Second.Seconds()), 1*int64(time.Minute.Seconds())),
		jwtKey:       JWTKey,
	}, nil
}

func (c *AuthController) Challenge(domain, address, uri string, chainID int) (string, error) {
	var opt = map[string]interface{}{
		"chainId":   chainID,
		"statement": purposeStatement,
	}
	msg, err := siwe.InitMessage(domain, address, uri, c.GetNonce(), opt)
	if err != nil {
		return "", err
	}
	fmt.Println(opt["statement"])
	return msg.String(), nil
}

func (c *AuthController) Login(request interface{}) (string, string, error) {
	req, ok := request.(EIP4361Request)
	if !ok {
		return "", "", fmt.Errorf("")
	}
	return c.loginWithEth(req)
}

func (c *AuthController) loginWithEth(request EIP4361Request) (string, string, error) {
	message, err := parseLensMessage(request.EIP191Message)
	if err != nil {
		return "", "", err
	}

	// if message.GetChainID() != ChainID {
	// 	return "", "", "", logs.AuthenticationFailed{Message: "Got wrong chain id"}
	// }

	if !c.VerifyNonce(message.GetNonce()) {
		return "", "", xerrors.New("Got wrong nonce")
	}

	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(request.EIP191Message), request.EIP191Message)))
	sig, err := hexutil.Decode(request.Signature)
	if err != nil {
		return "", "", err
	}

	sig[len(sig)-1] %= 27
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return "", "", err
	}

	if message.GetAddress().Hex() != crypto.PubkeyToAddress(*pubKey).Hex() {
		return "", "", xerrors.New("Got wrong address/signature")
	}

	accessToken, err := genAccessToken(message.GetAddress().Hex(), message.GetChainID())
	if err != nil {
		return "", "", err
	}

	refreshToken, err := genRefreshToken(message.GetAddress().Hex(), message.GetChainID())

	return accessToken, refreshToken, err
}

func parseLensMessage(message string) (*siwe.Message, error) {
	message = strings.TrimPrefix(message, "\n")
	message = strings.TrimPrefix(message, "https://")
	message = strings.TrimPrefix(message, "http://")
	message = strings.TrimSuffix(message, "\n ")

	return siwe.ParseMessage(message)
}
