package cbf

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/gaozhengxin/go-iden3-crypto/constants"
	"github.com/gaozhengxin/go-iden3-crypto/mimc7"
)

type Filter interface {
	Test([]byte) bool

	Add([]byte) Filter

	TestAndAdd([]byte) bool

	TestAndRemove([]byte) bool

	Equals(Filter) bool
}

func Equals(f1, f2 Filter) bool {
	return f1.Equals(f2)
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
	data := make([]uint8, len(cbf.data))
	copy(data, cbf.data)
	return data
}

func (cbf *CBF) Count() uint {
	return cbf.count
}

func (cbf *CBF) String() string {
	return string(cbf.Marshal())
}

func (cbf1 *CBF) Equals(f2 Filter) bool {
	cbf2, ok := f2.(*CBF)
	if !ok {
		return false
	}
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

type BF CBF

func (cbf *CBF) ToBF() *BF {
	bf := new(BF)
	bf.m = cbf.m
	bf.k = cbf.k
	bf.count = cbf.count
	bf.data = make([]uint8, bf.m)
	for i, b := range cbf.data {
		bf.data[i] = b / b
	}
	return bf
}

func NewBF(m uint, k uint) *BF {
	return &BF{
		m:     m,
		k:     k,
		data:  make([]uint8, m),
		count: 0,
	}
}

func (bf *BF) M() uint {
	return bf.m
}

func (bf *BF) K() uint {
	return bf.k
}

func (bf *BF) Data() []uint8 {
	data := make([]uint8, len(bf.data))
	copy(data, bf.data)
	return data
}

func (bf *BF) Count() uint {
	return bf.count
}

func (bf *BF) String() string {
	return string(bf.Marshal())
}

func (bf1 *BF) Equals(f2 Filter) bool {
	bf2, ok := f2.(*BF)
	if !ok {
		return false
	}
	equal := bf1.m == bf2.m && bf1.k == bf2.k && bf1.count == bf2.count && len(bf1.data) == len(bf2.data)
	if equal {
		for i := 0; i < len(bf1.data); i++ {
			if bf1.data[i] != bf2.data[i] {
				return false
			}
		}
		return true
	}
	return false
}

func (bf *BF) Marshal() []byte {
	cbfSerialized := struct {
		M     uint   `json:"M"`
		K     uint   `json:"K"`
		Data  string `json:"Data"`
		Count uint   `json:"Count"`
	}{
		M:     bf.m,
		K:     bf.k,
		Data:  hex.EncodeToString([]byte(bf.data)),
		Count: bf.count,
	}
	bjson, err := json.Marshal(cbfSerialized)
	if err != nil {
		panic(err)
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(bjson)))
	base64.StdEncoding.Encode(dst, bjson)
	return dst
}

func (bf *BF) Unmarshal(src []byte) *BF {
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
	bf.m = cbfSerialized.M
	bf.k = cbfSerialized.K
	bf.count = cbfSerialized.Count
	bz, _ := hex.DecodeString(cbfSerialized.Data)
	bf.data = []uint8(bz)
	return bf
}

func (bf *BF) Test(data []byte) bool {
	for i := uint(0); i < bf.k; i++ {
		h := Hash(data, i)
		h = new(big.Int).Mod(h, big.NewInt(int64(bf.m)))
		if bf.data[h.Int64()] == 0 {
			return false
		}
	}
	return true
}

func (bf *BF) Add(data []byte) Filter {
	for i := uint(0); i < bf.k; i++ {
		h := Hash(data, i)
		h = new(big.Int).Mod(h, big.NewInt(int64(bf.m)))
		bf.data[h.Int64()] = 1
	}
	bf.count++
	return bf
}

func (bf *BF) TestAndAdd(data []byte) bool {
	member := true
	for i := uint(0); i < bf.k; i++ {
		h := Hash(data, i)
		h = new(big.Int).Mod(h, big.NewInt(int64(bf.m)))
		if bf.data[h.Int64()] == 0 {
			member = false
			bf.data[h.Int64()] = 1
		}
	}
	bf.count++
	return member
}

func (bf *BF) TestAndRemove(data []byte) bool {
	// Cannot remove elements from from a BF
	return false
}
