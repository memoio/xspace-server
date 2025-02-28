package config

import "github.com/ethereum/go-ethereum/common"

const DefaultSpace int = 10

var (
	DevChain     = "https://devchain.metamemo.one:8501"
	TestChain    = "https://testchain.metamemo.one:24180"
	ProductChain = "https://chain.metamemo.one:8501"

	// dev
	DevNFTAddr = common.HexToAddress("0xa75150D716423c069529A3B2908Eb454e0a00Dfc")
	// test
	TestnetNFTAddr = common.HexToAddress("0x4044c388E5d8BC5b7E53e329f72e6f6633904cca")
	// product
	ProductNFTAddr = common.HexToAddress("0x00db967F78E46Db1082418194EE6B6A64fc8Fc88")
)

func GetContractInfoByChain(chain string) (string, common.Address) {
	switch chain {
	case "dev":
		return DevChain, DevNFTAddr
	case "test":
		return TestChain, TestnetNFTAddr
	case "product":
		return ProductChain, ProductNFTAddr
	}
	return DevChain, DevNFTAddr
}
