package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/mattetti/filebuffer"
	"net/http"
	"time"
)

type WriteKeepassRequest struct {
}

func handleWriteKeepassRequestInternal(req *WriteKeepassRequest) error {
	if state.Inst.S3ExportBucket == "" || state.Inst.S3ExportApiKey == "" || state.Inst.S3ExportSecret == "" {
		return errors.New("S3ExportBucket, S3ExportApiKey or S3ExportSecret undefined")
	}

	var keepassOutFile bytes.Buffer

	if err := keepassExport(state.Inst.GetMasterPassword(), &keepassOutFile); err != nil {
		return err
	}

	manualCredential := credentials.NewStaticCredentials(
		state.Inst.S3ExportApiKey,
		state.Inst.S3ExportSecret,
		"")

	awsSession, errSession := session.NewSession()
	if errSession != nil {
		return errSession
	}

	s3Client := s3.New(awsSession, aws.NewConfig().WithCredentials(manualCredential).WithRegion(endpoints.UsEast1RegionID))

	// why filebuffer?
	// https://stackoverflow.com/questions/20602131/io-writeseeker-and-io-readseeker-from-byte-or-file#comment60685488_20602219

	_, errS3Put := s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(state.Inst.S3ExportBucket),
		Key:    aws.String("/databases/" + time.Now().UTC().Format(time.RFC3339) + ".kdbx"),
		Body:   filebuffer.New(keepassOutFile.Bytes()),
	})
	if errS3Put != nil {
		return errS3Put
	}

	return nil
}

func HandleWriteKeepassRequest(w http.ResponseWriter, r *http.Request) {
	var req WriteKeepassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := handleWriteKeepassRequestInternal(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.CommandGenericSuccess(w, r)
}
