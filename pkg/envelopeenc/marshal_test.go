package envelopeenc

import (
	"github.com/function61/gokit/assert"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	out, err := (&Envelope{
		KeySlots: []envelopeKeySlot{
			{
				KekId:        "foo",
				DekEncrypted: []byte{0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x00},
			},
		},
		EncryptedContent: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
	}).Marshal()
	assert.Ok(t, err)

	env, err := Unmarshal(out)
	assert.Ok(t, err)

	assert.EqualJson(t, env, `{
  "key_slots": [
    {
      "kek_id": "foo",
      "dek_encrypted": "BwYFBAMCAQA="
    }
  ],
  "content": "AAECAwQFBgc="
}`)
}
