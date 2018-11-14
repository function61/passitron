package mac

import (
	"github.com/function61/gokit/assert"
	"testing"
)

const (
	key1 = "fooAuthenticationKey"
	key2 = "barAuthenticationKey"
)

func TestAuthenticate(t *testing.T) {
	msg := "hello world"

	signature := New(key1, msg).Sign()

	assert.EqualString(t, signature, "f2a78249534b6b01")

	assert.True(t, New(key1, msg).Authenticate(signature) == nil)
	assert.True(t, New(key1, msg).Authenticate("wrong signature") == ErrMacValidationFailed)
}

func TestDifferentMessagesProduceDifferentSignatures(t *testing.T) {
	assert.EqualString(t, New(key1, "msg A").Sign(), "93945caa95ab362f")
	assert.EqualString(t, New(key1, "msg B").Sign(), "00070486a22d08b4")
}

func TestDifferentKeysProduceDifferentSignatures(t *testing.T) {
	msg := "message to authenticate"

	assert.EqualString(t, New(key1, msg).Sign(), "765b53a2f01b28ea")
	assert.EqualString(t, New(key2, msg).Sign(), "19207bd1bc992f97")
}
