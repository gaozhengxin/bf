package cbf

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/gaozhengxin/go-iden3-crypto/constants"
	"github.com/gaozhengxin/go-iden3-crypto/mimc7"
)

type CBF struct {
	m     uint
	k     uint
	data  []uint8
	count uint
}

func NewCBF(m uint, k uint) *CBF {
	return &CBF{
		m:     m,
		k:     k,
		data:  make([]uint8, m),
		count: 0,
	}
}

func (cbf *CBF) M() uint {
	return cbf.m
}

func (cbf *CBF) K() uint {
	return cbf.k
}

func (cbf *CBF) Data() []uint8 {
	var data []uint8
	copy(data, cbf.data)
	return data
}

func (cbf *CBF) Count() uint {
	return cbf.count
}

func (cbf *CBF) String() string {
	return string(cbf.Marshal())
}

func Equals(cbf1, cbf2 *CBF) bool {
	equal := cbf1.m == cbf2.m && cbf1.k == cbf2.k && cbf1.count == cbf2.count && len(cbf1.data) == len(cbf2.data)
	if equal {
		for i := 0; i < len(cbf1.data); i++ {
			if cbf1.data[i] != cbf2.data[i] {
				return false
			}
		}
		return true
	}
	return false
}

func (cbf *CBF) Marshal() []byte {
	cbfSerialized := struct {
		M     uint   `json:"M"`
		K     uint   `json:"K"`
		Data  string `json:"Data"`
		Count uint   `json:"Count"`
	}{
		M:     cbf.m,
		K:     cbf.k,
		Data:  hex.EncodeToString([]byte(cbf.data)),
		Count: cbf.count,
	}
	bjson, err := json.Marshal(cbfSerialized)
	if err != nil {
		panic(err)
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(bjson)))
	base64.StdEncoding.Encode(dst, bjson)
	return dst
}

func (cbf *CBF) Unmarshal(src []byte) *CBF {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, _ := base64.StdEncoding.Decode(dst, src)
	cbfSerialized := struct {
		M     uint   `json:"M"`
		K     uint   `json:"K"`
		Data  string `json:"Data"`
		Count uint   `json:"Count"`
	}{}
	dst = dst[:n]

	err := json.Unmarshal(dst, &cbfSerialized)
	if err != nil {
		panic(err)
	}
	cbf.m = cbfSerialized.M
	cbf.k = cbfSerialized.K
	cbf.count = cbfSerialized.Count
	bz, _ := hex.DecodeString(cbfSerialized.Data)
	cbf.data = []uint8(bz)
	return cbf
}

type Filter interface {
	Test([]byte) bool

	Add([]byte) Filter

	TestAndAdd([]byte) bool

	TestAndRemove([]byte) bool
}

func EncodeData(data []byte) *big.Int {
	bigInt := new(big.Int).SetBytes(data)
	bigInt = new(big.Int).Mod(bigInt, constants.Q)
	return bigInt
}

func Hash(data []byte, k uint) *big.Int {
	arr := make([]*big.Int, 0)
	arr = append(arr, EncodeData(data))
	arr = append(arr, big.NewInt(int64(k)))
	res, err := mimc7.Hash(arr, nil, 5)
	if err != nil {
		panic(err)
	}
	return res
}

func (cbf *CBF) Test(data []byte) bool {
	for i := uint(0); i < cbf.k; i++ {
		h := Hash(data, i)
		h = new(big.Int).Mod(h, big.NewInt(int64(cbf.m)))
		if cbf.data[h.Int64()] == 0 {
			return false
		}
	}
	return true
}

func (cbf *CBF) Add(data []byte) Filter {
	for i := uint(0); i < cbf.k; i++ {
		h := Hash(data, i)
		h = new(big.Int).Mod(h, big.NewInt(int64(cbf.m)))
		cbf.data[h.Int64()]++
	}
	cbf.count++
	return cbf
}

func (cbf *CBF) TestAndAdd(data []byte) bool {
	member := true
	for i := uint(0); i < cbf.k; i++ {
		h := Hash(data, i)
		h = new(big.Int).Mod(h, big.NewInt(int64(cbf.m)))
		if cbf.data[h.Int64()] == 0 {
			member = false
		}
		cbf.data[h.Int64()]++
	}
	cbf.count++
	return member
}

func (cbf *CBF) TestAndRemove(data []byte) bool {
	member := true
	var indexBuffer []int64
	for i := uint(0); i < cbf.k; i++ {
		h := Hash(data, i)
		h = new(big.Int).Mod(h, big.NewInt(int64(cbf.m)))
		indexBuffer = append(indexBuffer, h.Int64())
		if cbf.data[h.Int64()] == 0 {
			member = false
		}
	}
	if member {
		for _, idx := range indexBuffer {
			cbf.data[idx]--
		}
		cbf.count++
	}
	return member
}
