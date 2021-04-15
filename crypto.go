package pagemanager

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/sq"
	"github.com/bokwoon95/pagemanager/tables"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/ssh/terminal"
)

type keyDerivation struct {
	// params used to derive the key
	argon2Version int
	memory        uint32
	time          uint32
	threads       uint8
	keyLen        uint32
	salt          []byte
	// the derived key
	key []byte
}

func (kd keyDerivation) Marshal() string {
	s := "$argon2id$v=" + strconv.Itoa(kd.argon2Version) +
		"$m=" + strconv.FormatUint(uint64(kd.memory), 10) +
		",t=" + strconv.FormatUint(uint64(kd.time), 10) +
		",p=" + strconv.FormatUint(uint64(kd.threads), 10) +
		",l=" + strconv.FormatUint(uint64(kd.keyLen), 10) +
		"$" + base64.RawURLEncoding.EncodeToString(kd.salt) + "$"
	if len(kd.key) > 0 {
		s += base64.RawURLEncoding.EncodeToString(kd.key)
	}
	return s
}

func (kd keyDerivation) MarshalParams() string {
	return "$argon2id$v=" + strconv.Itoa(kd.argon2Version) +
		"$m=" + strconv.FormatUint(uint64(kd.memory), 10) +
		",t=" + strconv.FormatUint(uint64(kd.time), 10) +
		",p=" + strconv.FormatUint(uint64(kd.threads), 10) +
		",l=" + strconv.FormatUint(uint64(kd.keyLen), 10) +
		"$" + base64.RawURLEncoding.EncodeToString(kd.salt) + "$"
}

func (kd *keyDerivation) Unmarshal(s string) error {
	// parts[0] = empty string
	// parts[1] = argon2id
	// parts[2] = v=%d
	// parts[3] = m=%d,t=%d,p=%d,l=%d
	// parts[4] = base64 URL encoded salt
	// parts[5] = base64 URL encoded key (can be empty, which indicates that key should be re-derived from the above params)
	var err error
	parts := strings.Split(s, "$")
	kd.argon2Version, err = strconv.Atoi(parts[2])
	if err != nil {
		return erro.Wrap(err)
	}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d,l=%d", &kd.memory, &kd.time, &kd.threads, &kd.keyLen)
	if err != nil {
		return erro.Wrap(err)
	}
	kd.salt, err = base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil {
		return erro.Wrap(err)
	}
	if len(parts[5]) > 0 {
		kd.key, err = base64.RawURLEncoding.DecodeString(parts[5])
		if err != nil {
			return erro.Wrap(err)
		}
	}
	return nil
}

func deriveKeyFromPassword(password string) (keyDerivation, error) {
	kd := keyDerivation{
		argon2Version: argon2.Version,
		memory:        63 * 1024,
		time:          1,
		threads:       4,
		keyLen:        32,
		salt:          make([]byte, 16),
	}
	_, err := rand.Read(kd.salt)
	if err != nil {
		return kd, erro.Wrap(err)
	}
	kd.key = argon2.IDKey([]byte(password), kd.salt, kd.time, kd.memory, kd.threads, kd.keyLen)
	return kd, nil
}

func verifyHashAndPassword(passwordHash string, password string) error {
	var kd keyDerivation
	err := kd.Unmarshal(passwordHash)
	if err != nil {
		return erro.Wrap(err)
	}
	derivedKey := argon2.IDKey([]byte(password), kd.salt, kd.time, kd.memory, kd.threads, kd.keyLen)
	if subtle.ConstantTimeCompare(kd.key, derivedKey) != 1 {
		return fmt.Errorf("password is invalid")
	}
	return nil
}

func readPassword(prompt string) (pw []byte, err error) {
	fd := int(os.Stdin.Fd())
	if terminal.IsTerminal(fd) {
		fmt.Fprint(os.Stderr, prompt)
		pw, err = terminal.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		return
	}
	var b [1]byte
	for {
		n, err := os.Stdin.Read(b[:])
		// terminal.ReadPassword discards any '\r', so we do the same
		if n > 0 && b[0] != '\r' {
			if b[0] == '\n' {
				return pw, nil
			}
			pw = append(pw, b[0])
			// limit size, so that a wrong input won't fill up the memory
			if len(pw) > 1024 {
				err = errors.New("password too long")
			}
		}
		if err != nil {
			// terminal.ReadPassword accepts EOF-terminated passwords
			// if non-empty, so we do the same
			if err == io.EOF && len(pw) > 0 {
				err = nil
			}
			return pw, err
		}
	}
}

func encrypt(key []byte, plaintext string) (ciphertext string, err error) {
	hashedKey := blake2b.Sum256(key)
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", erro.Wrap(err)
	}
	ciphertextBytes := secretbox.Seal(nonce[:], []byte(plaintext), &nonce, &hashedKey)
	ciphertext = base64.RawURLEncoding.EncodeToString(ciphertextBytes)
	return ciphertext, nil
}

func decrypt(key []byte, ciphertext string) (plaintext string, ok bool, err error) {
	hashedKey := blake2b.Sum256(key)
	ciphertextBytes, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", false, erro.Wrap(err)
	}
	var nonce [24]byte
	copy(nonce[:], ciphertextBytes[:24])
	plaintextBytes, ok := secretbox.Open(nil, ciphertextBytes[24:], &nonce, &hashedKey)
	if !ok {
		return "", false, erro.Wrap(fmt.Errorf("decryption error"))
	}
	return string(plaintextBytes), ok, nil
}

func (pm *PageManager) Encrypt(plaintext string) (ciphertext string, err error) {
	ctx := context.Background()
	var encryptionKeyCiphertext sql.NullString
	ENCRYPTION_KEYS := tables.NEW_ENCRYPTION_KEYS(ctx, "")
	_, err = sq.Fetch(pm.superadminDB, sq.SQLite.
		From(ENCRYPTION_KEYS).
		Where(ENCRYPTION_KEYS.ID.EqInt(1)),
		func(row *sq.Row) error {
			encryptionKeyCiphertext = row.NullString(ENCRYPTION_KEYS.KEY_CIPHERTEXT)
			return nil
		},
	)
	if err != nil {
		return "", erro.Wrap(err)
	}
	if !encryptionKeyCiphertext.Valid {
		return "", erro.Wrap(fmt.Errorf("no encryption keys found"))
	}
	encryptionKey, ok, err := decrypt(pm.innerEncryptionKey, encryptionKeyCiphertext.String)
	if err != nil {
		return "", erro.Wrap(err)
	}
	if !ok {
		return "", erro.Wrap(fmt.Errorf("decryption error"))
	}
	ciphertext, err = encrypt([]byte(encryptionKey), plaintext)
	if err != nil {
		return "", erro.Wrap(err)
	}
	return ciphertext, nil
}

func (pm *PageManager) Decrypt(ciphertext string) (plaintext string, err error) {
	ctx := context.Background()
	var encryptionKeyCiphertexts []string
	ENCRYPTION_KEYS := tables.NEW_ENCRYPTION_KEYS(ctx, "")
	_, err = sq.Fetch(pm.superadminDB, sq.SQLite.
		From(ENCRYPTION_KEYS).
		OrderBy(ENCRYPTION_KEYS.ID),
		func(row *sq.Row) error {
			encryptionKeyCiphertext := row.String(ENCRYPTION_KEYS.KEY_CIPHERTEXT)
			return row.Accumulate(func() error {
				encryptionKeyCiphertexts = append(encryptionKeyCiphertexts, encryptionKeyCiphertext)
				return nil
			})
		},
	)
	if err != nil {
		return "", erro.Wrap(err)
	}
	for _, encryptionKeyCiphertext := range encryptionKeyCiphertexts {
		encryptionKey, ok, err := decrypt(pm.innerEncryptionKey, encryptionKeyCiphertext)
		if err != nil {
			return "", erro.Wrap(err)
		}
		if !ok {
			return "", erro.Wrap(fmt.Errorf("decryption error"))
		}
		plaintext, ok, err = decrypt([]byte(encryptionKey), ciphertext)
		if err != nil {
			return "", erro.Wrap(err)
		}
		if !ok {
			continue
		}
		return plaintext, nil
	}
	return "", erro.Wrap(fmt.Errorf("decryption error"))
}

func makeMAC(key []byte, data string) (mac string) {
	b := blake2b.Sum512([]byte(key))
	h, _ := blake2b.New512(b[:])
	h.Reset()
	h.Write([]byte(data))
	sum := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum)
}

func verifyMAC(key []byte, data string, mac string) bool {
	b := blake2b.Sum512([]byte(key))
	h, _ := blake2b.New512(b[:])
	h.Reset()
	h.Write([]byte(data))
	computedSum := h.Sum(nil)
	providedSum, err := base64.RawURLEncoding.DecodeString(mac)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(computedSum, providedSum) == 1
}
