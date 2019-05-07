package btcec

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/daglabs/btcd/dagconfig/daghash"
)

var testVectors = []struct {
	dataElementHex string
	point          [2]string
	ecmhHash       string
	cumulativeHash string
}{
	{
		"982051fd1e4ba744bbbe680e1fee14677ba1a3c3540bf7b1cdb606e857233e0e00000000010000000100f2052a0100000043410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac",
		[2]string{"4f9a5dce69067bf28603e73a7af4c3650b16539b95bad05eee95dfc94d1efe2c", "346d5b777881f2729e7f89b2de4e8e79c7f2f42d1a0b25a8f10becb66e2d0f98"},
		"f883195933a687170c34fa1adec66fe2861889279fb12c03a3fb0ca68ad87893",
		"",
	},
	{
		"d5fdcc541e25de1c7a5addedf24858b8bb665c9f36ef744ee42c316022c90f9b00000000020000000100f2052a010000004341047211a824f55b505228e4c3d5194c1fcfaa15a456abdf37f9b9d97a4040afc073dee6c89064984f03385237d92167c13e236446b417ab79a0fcae412ae3316b77ac",
		[2]string{"68cf91eb2388a0287c13d46011c73fb8efb6be89c0867a47feccb2d11c390d2d", "f42ba72b1079d3d941881836f88b5dcd7c207a6a4839f129272c77ebb7194d42"},
		"ef85d123a15da95d8aff92623ad1e1c9fcda3baa801bd40bc567a83a6fdcf3e2",
		"fabafd38d07370982a34547daf5b57b8a4398696d6fd2294788abda07b1faaaf",
	},
	{
		"44f672226090d85db9a9f2fbfe5f0f9609b387af7be5b7fbb7a1767c831c9e9900000000030000000100f2052a0100000043410494b9d3e76c5b1629ecf97fff95d7a4bbdac87cc26099ada28066c6ff1eb9191223cd897194a08d0c2726c5747f1db49e8cf90e75dc3e3550ae9b30086f3cd5aaac",
		[2]string{"359c6f59859d1d5af8e7081905cb6bb734c010be8680c14b5a89ee315694fc2b", "fb6ba531d4bd83b14c970ad1bec332a8ae9a05706cd5df7fd91a2f2cc32482fe"},
		"cfadf40fc017faff5e04ccc0a2fae0fd616e4226dd7c03b1334a7a610468edff",
		"1cbccda23d7ce8c5a8b008008e1738e6bf9cffb1d5b86a92a4e62b5394a636e2",
	},
}

func TestHashToPoint(t *testing.T) {
	for _, test := range testVectors {
		data, err := hex.DecodeString(test.dataElementHex)
		if err != nil {
			t.Fatal(err)
		}
		x, y := hashToPoint(S256(), data)
		if hex.EncodeToString(x.Bytes()) != test.point[0] || hex.EncodeToString(y.Bytes()) != test.point[1] {
			t.Fatal("hashToPoint return incorrect point")
		}
	}
}

func TestMultiset_Hash(t *testing.T) {
	for _, test := range testVectors {
		data, err := hex.DecodeString(test.dataElementHex)
		if err != nil {
			t.Fatal(err)
		}
		x, y := hashToPoint(S256(), data)
		m := NewMultiset(S256())
		m.x, m.y = x, y
		if m.Hash().String() != test.ecmhHash {
			t.Fatal("Multiset-Hash returned incorrect hash serialization")
		}
	}
	m := NewMultiset(S256())
	emptySet := m.Hash()
	zeroHash := daghash.Hash{}
	if !bytes.Equal(emptySet[:], zeroHash[:]) {
		t.Fatal("Empty set did not return zero hash")
	}
}

func TestMultiset_AddRemove(t *testing.T) {
	m := NewMultiset(S256())
	for _, test := range testVectors {
		data, err := hex.DecodeString(test.dataElementHex)
		if err != nil {
			t.Fatal(err)
		}
		m.Add(data)
		if test.cumulativeHash != "" && m.Hash().String() != test.cumulativeHash {
			t.Fatal("Multiset-Add returned incorrect hash")
		}
	}

	for i := len(testVectors) - 1; i > 0; i-- {
		data, err := hex.DecodeString(testVectors[i].dataElementHex)
		if err != nil {
			t.Fatal(err)
		}
		m.Remove(data)
		if testVectors[i-1].cumulativeHash != "" && m.Hash().String() != testVectors[i-1].cumulativeHash {
			t.Fatal("Multiset-Remove returned incorrect hash")
		}
	}
}

func TestMultiset_Commutativity(t *testing.T) {
	m := NewMultiset(S256())
	zeroHash := m.Hash().String()

	for _, test := range testVectors {
		data, err := hex.DecodeString(test.dataElementHex)
		if err != nil {
			t.Fatal(err)
		}
		m.Remove(data)
	}

	for _, test := range testVectors {
		data, err := hex.DecodeString(test.dataElementHex)
		if err != nil {
			t.Fatal(err)
		}
		m.Add(data)
	}
	if m.Hash().String() != zeroHash {
		t.Fatalf("multiset was expected to return to zero hash, but was %s instead", m.Hash())
	}

	removeIndex := 0
	removeData, err := hex.DecodeString(testVectors[removeIndex].dataElementHex)
	if err != nil {
		t.Fatal(err)
	}

	m1 := NewMultiset(S256())
	m1.Remove(removeData)

	for i, test := range testVectors {
		if i != removeIndex {
			data, err := hex.DecodeString(test.dataElementHex)
			if err != nil {
				t.Fatal(err)
			}
			m1.Add(data)
		}
	}

	m2 := NewMultiset(S256())
	for i, test := range testVectors {
		if i != removeIndex {
			data, err := hex.DecodeString(test.dataElementHex)
			if err != nil {
				t.Fatal(err)
			}
			m2.Add(data)
		}
	}
	m2.Remove(removeData)

	if m1.Hash().String() != m2.Hash().String() {
		t.Fatalf("m1 and m2 was exepcted, but got instead m1 %s and m2 %s", m1.Hash(), m2.Hash())
	}
}
