package crypto

import (
	"github.com/stretchr/testify/assert"
	"encoding/base64"
	"golang.org/x/crypto/openpgp/armor"
	"io/ioutil"
	"strings"
	"testing"
)

var decodedSymmetricKey, _ = base64.StdEncoding.DecodeString("ExXmnSiQ2QCey20YLH6qlLhkY3xnIBC1AwlIXwK/HvY=");

var testSymmetricKey = &SymmetricKey{
	Key:  decodedSymmetricKey,
	Algo: "aes256",
}

// Corresponding key in testdata/keyring_privateKey
const testMailboxPassword = "apple"

// Corresponding key in testdata/keyring_privateKeyLegacy
const testMailboxPasswordLegacy = "123"

const testToken = "d79ca194a22810a5363eeddfdef7dfbc327c6229"

var (
	testPrivateKeyRing *KeyRing
	testPublicKeyRing  *KeyRing
)

var testIdentity = &Identity{
	Name:  "UserID",
	Email: "",
}

func readTestFile(name string) string {
	data, err := ioutil.ReadFile("testdata/" + name)
	if err != nil {
	  panic(err)
	}
	return string(data)
}

func init() {
	var err error
	if testPrivateKeyRing, err = ReadArmoredKeyRing(strings.NewReader(readTestFile("keyring_privateKey"))); err != nil {
		panic(err)
	}

	if testPublicKeyRing, err = ReadArmoredKeyRing(strings.NewReader(readTestFile("keyring_publicKey"))); err != nil {
		panic(err)
	}

	if err := testPrivateKeyRing.Unlock([]byte(testMailboxPassword)); err != nil {
		panic(err)
	}
}

func TestKeyRing_Decrypt(t *testing.T) {
	ss, err := testPrivateKeyRing.DecryptString(readTestFile("keyring_token"))
	if err != nil {
		t.Fatal("Cannot decrypt token:", err)
	}

	assert.Exactly(t, testToken, ss.String)
}

func TestKeyRing_Encrypt(t *testing.T) {
	encrypted, err := testPublicKeyRing.EncryptString(testToken, nil)
	if err != nil {
		t.Fatal("Cannot encrypt token:", err)
	}

	// We can't just check if encrypted == testEncryptedToken
	// Decrypt instead
	ss, err := testPrivateKeyRing.DecryptString(encrypted)
	if err != nil {
		t.Fatal("Cannot decrypt token:", err)
	}

	assert.Exactly(t, testToken, ss.String)
}

func TestKeyRing_ArmoredPublicKeyString(t *testing.T) {
	s, err := testPrivateKeyRing.ArmoredPublicKeyString()
	if err != nil {
		t.Fatal("Expected no error while getting armored public key, got:", err)
	}

	// Decode armored keys
	block, err := armor.Decode(strings.NewReader(s))
	if err != nil {
		t.Fatal("Expected no error while decoding armored public key, got:", err)
	}

	expected, err := armor.Decode(strings.NewReader(readTestFile("keyring_publicKey")))
	if err != nil {
		t.Fatal("Expected no error while decoding expected armored public key, got:", err)
	}

	assert.Exactly(t, expected.Type, block.Type)

	b, err := ioutil.ReadAll(block.Body)
	if err != nil {
		t.Fatal("Expected no error while reading armored public key body, got:", err)
	}

	eb, err := ioutil.ReadAll(expected.Body)
	if err != nil {
		t.Fatal("Expected no error while reading expected armored public key body, got:", err)
	}

	assert.Exactly(t, eb, b)
}
