package magicalauth

import (
	"context"
	"strings"
	"sync"
)

const MaxPrepareMultiple = 10

type PrepareMultipleRequest struct {
	PhoneNumbers []string
	PLMN         *PLMN
	UseCase      UseCase
	ConsentData  *ConsentData
	ClientInfo   ClientInfo
	Options      *PrepareOptions

	AuthenticationStrategy string
	SessionTimeout         *int
	PublicKey              string
}

type PrepareMultipleResult struct {
	PhoneNumber string
	Result      *PrepareResult
	Err         error
}

func (r PrepareMultipleResult) OK() bool { return r.Err == nil }

// forNumber builds a fresh single-number request. Prepare writes Nonce and FeHash
// back into the request it is given, so a shared pointer would race and reuse one nonce.
func (r *PrepareMultipleRequest) forNumber(phoneNumber string) *PrepareRequest {
	return &PrepareRequest{
		PhoneNumber:            phoneNumber,
		PLMN:                   r.PLMN,
		UseCase:                r.UseCase,
		ConsentData:            r.ConsentData,
		ClientInfo:             r.ClientInfo,
		Options:                r.Options,
		AuthenticationStrategy: r.AuthenticationStrategy,
		SessionTimeout:         r.SessionTimeout,
		PublicKey:              r.PublicKey,
	}
}

func validatePrepareMultipleRequest(req *PrepareMultipleRequest) error {
	if req == nil {
		return ErrNilRequest
	}
	if len(req.PhoneNumbers) == 0 {
		return ErrNoPhoneNumbers
	}
	if len(req.PhoneNumbers) > MaxPrepareMultiple {
		return ErrTooManyPhoneNumbers
	}
	for _, n := range req.PhoneNumbers {
		if strings.TrimSpace(n) == "" {
			return ErrBlankPhoneNumber
		}
	}
	return nil
}

func (c *Client) PrepareMultiple(
	ctx context.Context,
	req *PrepareMultipleRequest,
	opts ...*RequestOptions,
) ([]PrepareMultipleResult, error) {
	if err := validatePrepareMultipleRequest(req); err != nil {
		return nil, err
	}

	derived := make([]*PrepareRequest, len(req.PhoneNumbers))
	for i, phoneNumber := range req.PhoneNumbers {
		derived[i] = req.forNumber(phoneNumber)
		if err := c.validatePrepareRequest(derived[i]); err != nil {
			return nil, err
		}
	}

	c.logger.Info("preparing auth for multiple numbers", "count", len(req.PhoneNumbers))

	results := make([]PrepareMultipleResult, len(req.PhoneNumbers))
	var wg sync.WaitGroup
	for i, phoneNumber := range req.PhoneNumbers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := c.Prepare(ctx, derived[i], opts...)
			results[i] = PrepareMultipleResult{PhoneNumber: phoneNumber, Result: res, Err: err}
		}()
	}
	wg.Wait()

	return results, nil
}
