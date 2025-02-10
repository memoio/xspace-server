package auth

import (
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/xerrors"
)

var Domain = "xspace.com"

type Claims struct {
	Type         int  `json:"type,omitempty"`
	IsRegistered bool `json:"isRegistered,omitempty"`
	ChainID      int  `josn:"chainid,omitempty"`
	// Nonce string `json:"nonce,omitempty"`
	jwt.StandardClaims
}

func VerifyAccessToken(tokenString string) (string, int, error) {
	claims, err := verifyJsonWebToken(tokenString, AccessToken)
	if err != nil {
		return "", 0, err
	}

	return claims.Subject, claims.ChainID, nil
}

func VerifyRefreshToken(tokenString string) (string, error) {
	claims, err := verifyJsonWebToken(tokenString, RefreshToken)
	if err != nil {
		return "", err
	}

	return genAccessTokenWithFlag(claims.Subject, claims.ChainID, claims.IsRegistered)
}

func genAccessToken(subject string, chainID int) (string, error) {
	return genJsonWebTokenWithFlag(subject, chainID, AccessToken, false)
}

func genAccessTokenWithFlag(subject string, chainID int, isRegistered bool) (string, error) {
	return genJsonWebTokenWithFlag(subject, chainID, AccessToken, isRegistered)
}

func genRefreshToken(subject string, chainID int) (string, error) {
	return genJsonWebTokenWithFlag(subject, chainID, RefreshToken, false)
}

func genRefreshTokenWithFlag(subject string, chainID int, isRegistered bool) (string, error) {
	return genJsonWebTokenWithFlag(subject, chainID, RefreshToken, isRegistered)
}

func verifyJsonWebToken(tokenString string, jwtType int) (*Claims, error) {
	parts := strings.SplitN(tokenString, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return nil, ErrNullToken
	}

	claims := &Claims{}
	_, _, err := new(jwt.Parser).ParseUnverified(parts[1], claims)
	if err != nil {
		return nil, ErrValidToken
	}

	// check Audience
	if claims.Audience != Domain || claims.Issuer != Domain {
		return nil, ErrValidToken
	}

	// check token type
	if claims.Type != jwtType {
		return nil, ErrValidTokenType
	}

	// check signature, Expires time and Issued time
	token, err := parseToken(parts[1])
	if err != nil || !token.Valid {
		return nil, ErrValidToken
	}

	return claims, nil
}

func genJsonWebTokenWithFlag(subject string, chainID, jwtType int, isRegistered bool) (string, error) {
	var expireTime int64
	if jwtType == AccessToken {
		expireTime = time.Now().Add(2 * time.Hour).Unix()
	} else if jwtType == RefreshToken {
		expireTime = time.Now().Add(7 * 24 * time.Hour).Unix()
	} else {
		return "", xerrors.Errorf("unsupported json web token type")
	}

	claims := &Claims{
		Type:         jwtType,
		IsRegistered: isRegistered,
		ChainID:      chainID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime,
			IssuedAt:  time.Now().Unix(),
			Audience:  Domain,
			Issuer:    Domain,
			Subject:   subject,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTKey)
}

// func ParseDidToken(tokenString string, did string) (*jwt.Token, error) {
//     return jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
//     	parts := strings.Split(did, ":")
//     	if len(parts) != 3 || parts[0] != "did" || parts[1] != "eth" {
//     		return nil, ErrValidToken
//     	}

//     	pubKeyBytes, err := hex.DecodeString(parts[2])
//     	if err != nil {
//     		return nil, err
//     	}

//         return crypto.UnmarshalPubkey(pubKeyBytes)
//     })
// }

func parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		return JWTKey, nil
	})
}
