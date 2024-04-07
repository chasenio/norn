package common

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"strings"
)

type WebhookPayload struct {
	Action string `json:"action"`
}

func WebhookToFile(payload interface{}, action string) {
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		logrus.Errorf("marshal payload error: %v", err)
		return
	}
	// hash request body
	h := md5.New()
	h.Write(content)
	hashVal := h.Sum(nil)

	plTypeName := reflect.TypeOf(payload).String()
	eventType := plTypeName[strings.LastIndex(plTypeName, ".")+1:]

	logrus.Infof("request body hash: %x, action: %s", hashVal, action)

	// generate file name
	fileName := fmt.Sprintf("%s.%s.%x.json", eventType, action, hashVal[:4])
	f, err := os.Create(fmt.Sprintf("example/%s", fileName))
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	_, _ = f.Write(content)
}
