package task

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/ngmmartins/asyncq/internal/validator"
)

type PayloadDecoder func(data json.RawMessage, v *validator.Validator) (any, error)

var decoderRegistry = map[Task]PayloadDecoder{}

func RegisterDecoder(task Task, fn PayloadDecoder) {
	decoderRegistry[task] = fn
}

func DecodePayload(task Task, data json.RawMessage) (any, error) {
	decoder, ok := decoderRegistry[task]
	if !ok {
		return nil, errors.New("no decoder registered for task: " + string(task))
	}
	return decoder(data, nil)
}

func DecodeAndValidatePayload(task Task, data json.RawMessage, v *validator.Validator) (any, error) {
	decoder, ok := decoderRegistry[task]
	if !ok {
		return nil, errors.New("no decoder registered for task: " + string(task))
	}
	return decoder(data, v)
}

func init() {
	RegisterDecoder(WebhookTask, func(data json.RawMessage, v *validator.Validator) (any, error) {
		var p WebhookPayload
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()

		if err := dec.Decode(&p); err != nil {
			return nil, err
		}

		if v != nil {
			ValidateWebhookPayload(v, &p)
		}
		return p, nil
	})

	RegisterDecoder(SendEmailTask, func(data json.RawMessage, v *validator.Validator) (any, error) {
		var p SendEmailPayload
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()

		if err := dec.Decode(&p); err != nil {
			return nil, err
		}
		if v != nil {
			ValidateSendEmailPayload(v, &p)
		}
		return p, nil
	})
}
