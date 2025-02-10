package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

type NonceManager struct {
	handledNonce  *sync.Map
	handlingNonce *sync.Map
	modifyMutex   sync.Mutex

	ExpireEpoch    int64
	ModifyEpoch    int64
	LastModifyTime int64
}

func NewNonceManager(expireEpoch int64, modifyEpoch int64) *NonceManager {
	return &NonceManager{
		handlingNonce:  new(sync.Map),
		handledNonce:   new(sync.Map),
		ExpireEpoch:    expireEpoch,
		ModifyEpoch:    modifyEpoch,
		LastModifyTime: time.Now().Unix(),
	}
}

func (non *NonceManager) GetNonce() string {
	now := time.Now().Unix()
	if now-non.LastModifyTime >= non.ModifyEpoch {
		non.clearExpiredNonce()
	}

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	nonce := hex.EncodeToString(crypto.Keccak256(b, []byte(time.Now().String())))
	non.handlingNonce.Store(nonce, time.Now().Unix()+non.ExpireEpoch)

	return nonce
}

func (non *NonceManager) VerifyNonce(nonce string) bool {
	if nonce == "" {
		return false
	}

	now := time.Now().Unix()
	if now-non.LastModifyTime >= non.ModifyEpoch {
		non.clearExpiredNonce()
	}

	expireTime, ok := non.handlingNonce.Load(nonce)
	if ok {
		non.handlingNonce.Delete(nonce)
		expireTimeInt, _ := expireTime.(int64)
		if now < expireTimeInt {
			return true
		} else {
			return false
		}
	}

	if time.Now().Unix()-non.LastModifyTime < non.ExpireEpoch {
		expireTime, ok = non.handledNonce.Load(nonce)
		if ok {
			non.handledNonce.Delete(nonce)
			expireTimeInt, _ := expireTime.(int64)
			if now < expireTimeInt {
				return true
			}
		}
	}

	return false
}

func (non *NonceManager) clearExpiredNonce() {
	now := time.Now().Unix()
	non.modifyMutex.Lock()
	defer non.modifyMutex.Unlock()
	if now-non.LastModifyTime >= non.ModifyEpoch {
		non.handledNonce = non.handlingNonce
		non.handlingNonce = new(sync.Map)
		non.LastModifyTime = now
	}
}
