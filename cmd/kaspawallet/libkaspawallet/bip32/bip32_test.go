package bip32

import (
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func TestBIP32SpecVectors(t *testing.T) {
	type testPath struct {
		path   string
		extPub string
		extPrv string
	}

	type testVector struct {
		seed   string
		pathes []testPath
	}

	// test vectors are copied from https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki#Test_Vectors
	testVectors := []testVector{
		{
			seed: "000102030405060708090a0b0c0d0e0f",
			pathes: []testPath{
				{
					path:   "m",
					extPub: "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
					extPrv: "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
				},
				{
					path:   "m/0'",
					extPub: "xpub68Gmy5EdvgibQVfPdqkBBCHxA5htiqg55crXYuXoQRKfDBFA1WEjWgP6LHhwBZeNK1VTsfTFUHCdrfp1bgwQ9xv5ski8PX9rL2dZXvgGDnw",
					extPrv: "xprv9uHRZZhk6KAJC1avXpDAp4MDc3sQKNxDiPvvkX8Br5ngLNv1TxvUxt4cV1rGL5hj6KCesnDYUhd7oWgT11eZG7XnxHrnYeSvkzY7d2bhkJ7",
				},
				{
					path:   "m/0'/1",
					extPub: "xpub6ASuArnXKPbfEwhqN6e3mwBcDTgzisQN1wXN9BJcM47sSikHjJf3UFHKkNAWbWMiGj7Wf5uMash7SyYq527Hqck2AxYysAA7xmALppuCkwQ",
					extPrv: "xprv9wTYmMFdV23N2TdNG573QoEsfRrWKQgWeibmLntzniatZvR9BmLnvSxqu53Kw1UmYPxLgboyZQaXwTCg8MSY3H2EU4pWcQDnRnrVA1xe8fs",
				},
				{
					path:   "m/0'/1/2'",
					extPub: "xpub6D4BDPcP2GT577Vvch3R8wDkScZWzQzMMUm3PWbmWvVJrZwQY4VUNgqFJPMM3No2dFDFGTsxxpG5uJh7n7epu4trkrX7x7DogT5Uv6fcLW5",
					extPrv: "xprv9z4pot5VBttmtdRTWfWQmoH1taj2axGVzFqSb8C9xaxKymcFzXBDptWmT7FwuEzG3ryjH4ktypQSAewRiNMjANTtpgP4mLTj34bhnZX7UiM",
				},
				{
					path:   "m/0'/1/2'/2",
					extPub: "xpub6FHa3pjLCk84BayeJxFW2SP4XRrFd1JYnxeLeU8EqN3vDfZmbqBqaGJAyiLjTAwm6ZLRQUMv1ZACTj37sR62cfN7fe5JnJ7dh8zL4fiyLHV",
					extPrv: "xprvA2JDeKCSNNZky6uBCviVfJSKyQ1mDYahRjijr5idH2WwLsEd4Hsb2Tyh8RfQMuPh7f7RtyzTtdrbdqqsunu5Mm3wDvUAKRHSC34sJ7in334",
				},
				{
					path:   "m/0'/1/2'/2/1000000000",
					extPub: "xpub6H1LXWLaKsWFhvm6RVpEL9P4KfRZSW7abD2ttkWP3SSQvnyA8FSVqNTEcYFgJS2UaFcxupHiYkro49S8yGasTvXEYBVPamhGW6cFJodrTHy",
					extPrv: "xprvA41z7zogVVwxVSgdKUHDy1SKmdb533PjDz7J6N6mV6uS3ze1ai8FHa8kmHScGpWmj4WggLyQjgPie1rFSruoUihUZREPSL39UNdE3BBDu76",
				},
			},
		},
		{
			seed: "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
			pathes: []testPath{
				{
					path:   "m",
					extPub: "xpub661MyMwAqRbcFW31YEwpkMuc5THy2PSt5bDMsktWQcFF8syAmRUapSCGu8ED9W6oDMSgv6Zz8idoc4a6mr8BDzTJY47LJhkJ8UB7WEGuduB",
					extPrv: "xprv9s21ZrQH143K31xYSDQpPDxsXRTUcvj2iNHm5NUtrGiGG5e2DtALGdso3pGz6ssrdK4PFmM8NSpSBHNqPqm55Qn3LqFtT2emdEXVYsCzC2U",
				},
				{
					path:   "m/0",
					extPub: "xpub69H7F5d8KSRgmmdJg2KhpAK8SR3DjMwAdkxj3ZuxV27CprR9LgpeyGmXUbC6wb7ERfvrnKZjXoUmmDznezpbZb7ap6r1D3tgFxHmwMkQTPH",
					extPrv: "xprv9vHkqa6EV4sPZHYqZznhT2NPtPCjKuDKGY38FBWLvgaDx45zo9WQRUT3dKYnjwih2yJD9mkrocEZXo1ex8G81dwSM1fwqWpWkeS3v86pgKt",
				},
				{
					path:   "m/0/2147483647'",
					extPub: "xpub6ASAVgeehLbnwdqV6UKMHVzgqAG8Gr6riv3Fxxpj8ksbH9ebxaEyBLZ85ySDhKiLDBrQSARLq1uNRts8RuJiHjaDMBU4Zn9h8LZNnBC5y4a",
					extPrv: "xprv9wSp6B7kry3Vj9m1zSnLvN3xH8RdsPP1Mh7fAaR7aRLcQMKTR2vidYEeEg2mUCTAwCd6vnxVrcjfy2kRgVsFawNzmjuHc2YmYRmagcEPdU9",
				},
				{
					path:   "m/0/2147483647'/1",
					extPub: "xpub6DF8uhdarytz3FWdA8TvFSvvAh8dP3283MY7p2V4SeE2wyWmG5mg5EwVvmdMVCQcoNJxGoWaU9DCWh89LojfZ537wTfunKau47EL2dhHKon",
					extPrv: "xprv9zFnWC6h2cLgpmSA46vutJzBcfJ8yaJGg8cX1e5StJh45BBciYTRXSd25UEPVuesF9yog62tGAQtHjXajPPdbRCHuWS6T8XA2ECKADdw4Ef",
				},
				{
					path:   "m/0/2147483647'/1/2147483646'",
					extPub: "xpub6ERApfZwUNrhLCkDtcHTcxd75RbzS1ed54G1LkBUHQVHQKqhMkhgbmJbZRkrgZw4koxb5JaHWkY4ALHY2grBGRjaDMzQLcgJvLJuZZvRcEL",
					extPrv: "xprvA1RpRA33e1JQ7ifknakTFpgNXPmW2YvmhqLQYMmrj4xJXXWYpDPS3xz7iAxn8L39njGVyuoseXzU6rcxFLJ8HFsTjSyQbLYnMpCqE2VbFWc",
				},
				{
					path:   "m/0/2147483647'/1/2147483646'/2",
					extPub: "xpub6FnCn6nSzZAw5Tw7cgR9bi15UV96gLZhjDstkXXxvCLsUXBGXPdSnLFbdpq8p9HmGsApME5hQTZ3emM2rnY5agb9rXpVGyy3bdW6EEgAtqt",
					extPrv: "xprvA2nrNbFZABcdryreWet9Ea4LvTJcGsqrMzxHx98MMrotbir7yrKCEXw7nadnHM8Dq38EGfSh6dqA9QWTyefMLEcBYJUuekgW4BYPJcr9E7j",
				},
			},
		},
		{
			seed: "4b381541583be4423346c643850da4b320e46a87ae3d2a4e6da11eba819cd4acba45d239319ac14f863b8d5ab5a0d0c64d2e8a1e7d1457df2e5a3c51c73235be",
			pathes: []testPath{
				{
					path:   "m",
					extPub: "xpub661MyMwAqRbcEZVB4dScxMAdx6d4nFc9nvyvH3v4gJL378CSRZiYmhRoP7mBy6gSPSCYk6SzXPTf3ND1cZAceL7SfJ1Z3GC8vBgp2epUt13",
					extPrv: "xprv9s21ZrQH143K25QhxbucbDDuQ4naNntJRi4KUfWT7xo4EKsHt2QJDu7KXp1A3u7Bi1j8ph3EGsZ9Xvz9dGuVrtHHs7pXeTzjuxBrCmmhgC6",
				},
				{
					path:   "m/0'",
					extPub: "xpub68NZiKmJWnxxS6aaHmn81bvJeTESw724CRDs6HbuccFQN9Ku14VQrADWgqbhhTHBaohPX4CjNLf9fq9MYo6oDaPPLPxSb7gwQN3ih19Zm4Y",
					extPrv: "xprv9uPDJpEQgRQfDcW7BkF7eTya6RPxXeJCqCJGHuCJ4GiRVLzkTXBAJMu2qaMWPrS7AANYqdq6vcBcBUdJCVVFceUvJFjaPdGZ2y9WACViL4L",
				},
			},
		},
	}

	for i, vector := range testVectors {
		seed, err := hex.DecodeString(vector.seed)
		if err != nil {
			t.Fatalf("DecodeString: %+v", err)
		}

		masterKey, err := NewMaster(seed, BitcoinMainnetPrivate)
		if err != nil {
			t.Fatalf("NewMaster: %+v", err)
		}

		for j, path := range vector.pathes {
			extPrv, err := masterKey.Path(path.path)
			if err != nil {
				t.Fatalf("Path: %+v", err)
			}

			if extPrv.String() != path.extPrv {
				t.Fatalf("Test (%d, %d): expected extPrv %s but got %s", i, j, path.extPrv, extPrv.String())
			}

			decodedExtPrv, err := DeserializeExtendedPrivateKey(extPrv.String())
			if err != nil {
				t.Fatalf("DeserializeExtendedPrivateKey: %+v", err)
			}

			if extPrv.String() != decodedExtPrv.String() {
				t.Fatalf("Test (%d, %d): deserializing and serializing the ext prv didn't preserve the data", i, j)
			}

			extPub, err := extPrv.Public()
			if err != nil {
				t.Fatalf("Public: %+v", err)
			}

			if extPub.String() != path.extPub {
				t.Fatalf("Test (%d, %d): expected extPub %s but got %s", i, j, path.extPub, extPub.String())
			}

			decodedExtPub, err := DeserializeExtendedPublicKey(extPub.String())
			if err != nil {
				t.Fatalf("DeserializeExtendedPublicKey: %+v", err)
			}

			if extPub.String() != decodedExtPub.String() {
				t.Fatalf("Test (%d, %d): deserializing and serializing the ext pub didn't preserve the data", i, j)
			}
		}
	}
}

func TestExtendedPublicKey_Path(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	seed, err := GenerateSeed()
	if err != nil {
		t.Fatalf("GenerateSeed: %+v", err)
	}

	master, err := NewMaster(seed, KaspaMainnetPrivate)
	if err != nil {
		t.Fatalf("GenerateSeed: %+v", err)
	}

	masterPublic, err := master.Public()
	if err != nil {
		t.Fatalf("Public: %+v", err)
	}

	for i := 0; i < 10; i++ {
		numIndexes := r.Intn(100)
		indexes := make([]string, numIndexes)
		for i := 0; i < numIndexes; i++ {
			index := r.Intn(hardenedIndexStart)
			indexes[i] = strconv.Itoa(int(index))
		}

		indexesStr := strings.Join(indexes, "/")
		pathPrivate := "m/" + indexesStr
		pathPublic := "M/" + indexesStr

		extPrv, err := master.Path(pathPrivate)
		if err != nil {
			t.Fatalf("Path: %+v", err)
		}

		extPubFromPrv, err := extPrv.Public()
		if err != nil {
			t.Fatalf("Public: %+v", err)
		}

		extPub, err := masterPublic.Path(pathPublic)
		if err != nil {
			t.Fatalf("Path: %+v", err)
		}

		if extPubFromPrv.String() != extPub.String() {
			t.Fatalf("Path gives different result from private and public master keys")
		}
	}
}
