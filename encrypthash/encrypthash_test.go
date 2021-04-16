package encrypthash

import (
	"fmt"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_Box(t *testing.T) {
	is := testutil.New(t)
	box, err := New([]byte("abcdefg"), nil)
	is.NoErr(err)

	t.Run("encryption", func(t *testing.T) {
		is := testutil.New(t)
		plaintext := []byte("lorem ipsum dolor sit amet")
		ciphertext, err := box.Base64Encrypt(plaintext)
		is.NoErr(err)
		fmt.Println(ciphertext)
		got, err := box.Base64Decrypt(ciphertext)
		is.NoErr(err)
		is.Equal(plaintext, got)
		_, err = box.Base64Decrypt(ciphertext + "tampered")
		is.True(err != nil)
	})

	t.Run("hash", func(t *testing.T) {
		is := testutil.New(t)
		msg := []byte("lorem ipsum dolor sit amet")
		hashedmsg, err := box.Base64Hash(msg)
		is.NoErr(err)
		fmt.Println(hashedmsg)
		got, err := box.Base64VerifyHash(hashedmsg)
		is.NoErr(err)
		is.Equal(msg, got)
		_, err = box.Base64VerifyHash(hashedmsg + "tampered")
		is.True(err != nil)
	})
}
