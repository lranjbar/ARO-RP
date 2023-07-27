package jsonutils

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestUpdateJsonString(t *testing.T) {
	type test struct {
		name             string
		jsonString       string
		jsonKeyValuePair map[string]string
		expectedString   string
		expectedChanged  bool
		expectedLog      *logrus.Entry
	}
	jsonStringByte, _ := json.Marshal(map[string]string{"test": "test", "test2": "test2"})

	for _, tt := range []*test{
		{
			name:             "no update to json string",
			jsonString:       string(jsonStringByte),
			jsonKeyValuePair: map[string]string{"test": "test"},
			expectedString:   "{\"test\":\"test\",\"test2\":\"test2\"}",
			expectedChanged:  false,
		},
		{
			name:             "update one key in json string",
			jsonString:       string(jsonStringByte),
			jsonKeyValuePair: map[string]string{"test": "updated"},
			expectedString:   "{\"test\":\"updated\",\"test2\":\"test2\"}",
			expectedChanged:  true,
		},
		{
			name:             "error while updating empty json string",
			jsonString:       "{}",
			jsonKeyValuePair: map[string]string{"test": "updated"},
			expectedString:   "{}",
			expectedChanged:  false,
			expectedLog:      &logrus.Entry{Level: logrus.ErrorLevel, Message: "key test does not exist in json string"},
		},
		{
			name:             "error while updating nonexistant key in json string",
			jsonString:       "{}",
			jsonKeyValuePair: map[string]string{"nottest": "updated"},
			expectedString:   "{}",
			expectedChanged:  false,
			expectedLog:      &logrus.Entry{Level: logrus.ErrorLevel, Message: "key nottest does not exist in json string"},
		},
	} {
		t.Run(tt.name, func(*testing.T) {
			logger := &logrus.Logger{
				Out:       io.Discard,
				Formatter: new(logrus.TextFormatter),
				Hooks:     make(logrus.LevelHooks),
				Level:     logrus.TraceLevel,
			}
			var hook = logtest.NewLocal(logger)

			actualString, actualChanged, err := UpdateJsonString(tt.jsonString, tt.jsonKeyValuePair)
			if err != nil {
				logger.Log(logrus.ErrorLevel, err)
			}

			actualLog := hook.LastEntry()
			if actualLog == nil {
				assert.Equal(t, tt.expectedLog, actualLog)
			} else {
				assert.Equal(t, tt.expectedLog.Level, actualLog.Level)
				assert.Equal(t, tt.expectedLog.Message, actualLog.Message)
			}

			assert.Equal(t, tt.expectedString, actualString)
			assert.Equal(t, tt.expectedChanged, actualChanged)
		})
	}
}
