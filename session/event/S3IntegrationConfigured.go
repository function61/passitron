package event

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type S3IntegrationConfigured struct {
	eventbase.Event
	Bucket string
	ApiKey string
	Secret string
}

func (e S3IntegrationConfigured) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "S3IntegrationConfigured " + string(asJson)
}

func S3IntegrationConfiguredFromSerialized(payload []byte) *S3IntegrationConfigured {
	var e S3IntegrationConfigured
	if err := json.Unmarshal(payload, &e); err != nil {
		panic(err)
	}
	return &e
}

func (e S3IntegrationConfigured) Apply() {
	state.Inst.S3ExportBucket = e.Bucket
	state.Inst.S3ExportApiKey = e.ApiKey
	state.Inst.S3ExportSecret = e.Secret
}
