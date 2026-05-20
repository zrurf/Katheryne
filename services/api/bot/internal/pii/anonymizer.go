package pii

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
)

type Anonymizer struct {
	salt string
}

func NewAnonymizer(salt string) *Anonymizer {
	return &Anonymizer{salt: salt}
}

func (a *Anonymizer) AnonymizeUID(uid int64) string {
	mac := hmac.New(sha256.New, []byte(a.salt))
	mac.Write([]byte(strconv.FormatInt(uid, 10)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (a *Anonymizer) AnonymizeConvID(convID int64) string {
	mac := hmac.New(sha256.New, []byte(a.salt+"_conv"))
	mac.Write([]byte(strconv.FormatInt(convID, 10)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (a *Anonymizer) AnonymizeGroupID(groupID int64) string {
	mac := hmac.New(sha256.New, []byte(a.salt+"_group"))
	mac.Write([]byte(strconv.FormatInt(groupID, 10)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}