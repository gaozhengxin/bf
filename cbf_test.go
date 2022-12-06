package cbf

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(111111)
	//.rand.Seed(1111)
}

var letterBytes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func TestHash(t *testing.T) {
	input := []byte("1111111122222222111111112222222211111111222222221111111122222222")
	res1 := Hash(input, 1)
	res2 := Hash(input, 2)
	res3 := Hash(input, 3)
	res4 := Hash(input, 4)
	t.Logf("%v\n", res1)
	t.Logf("%v\n", res2)
	t.Logf("%v\n", res3)
	t.Logf("%v\n", res4)
}

func TestCBF(t *testing.T) {
	cbf := NewCBF(uint(128), uint(4))

	input := RandStringBytes(64)

	assert.Equal(t, cbf.count, uint(0), "CBF count should be 0.")

	test0 := cbf.Test(input)
	assert.Equal(t, test0, false, "Test0 should be false.")

	cbf = cbf.Add(input).(*CBF)
	assert.Equal(t, cbf.count, uint(1), "CBF count should be 1.")

	test1 := cbf.Test(input)
	assert.Equal(t, test1, true, "Test1 should be true.")

	cbf = NewCBF(uint(1024), uint(4)) // 32 slot * 20000 gas / slot = 640000 gas

	inputs := make([][]byte, 0)
	for i := 0; i < 100; i++ {
		inputs = append(inputs, RandStringBytes(64))
	}

	collision := 0
	for _, input := range inputs {
		test := cbf.TestAndAdd(input)
		if test {
			collision++
		}
	}
	t.Logf("Collision %v\n", collision)

	assert.Equal(t, cbf.count, uint(100), "CBF count should be 0.")

	for _, input := range inputs {
		test := cbf.Test(input)
		assert.Equal(t, test, true, "Test should be true.")
	}

	var input2 []byte
	for {
		input2 = RandStringBytes(64)
		test := cbf.Test(input2)
		if !test {
			cbf.Add(input2)
			break
		}
	}
	test2 := cbf.Test(input2)
	assert.Equal(t, test2, true, "Test2 should be true.")

	test3 := cbf.TestAndRemove(input2)
	assert.Equal(t, test3, true, "Test3 should be true.")

	test4 := cbf.Test(input2)
	assert.Equal(t, test4, false, "Test4 should be false.")

	serialized := cbf.Marshal()
	unserialized := new(CBF).Unmarshal(serialized)
	assert.Equal(t, Equals(cbf, unserialized), true, "Unserialized error.")

	assert.Equal(t, cbf.data, cbf.Data(), "Get cbf data error")
}
