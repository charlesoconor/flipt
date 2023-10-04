package audit

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRetrier_Failure(t *testing.T) {
	retrier := NewRetrier(zap.NewNop(), 5*time.Second)

	gock.New("https://respond.io").
		MatchHeader("Content-Type", "application/json").
		Post("/webhook").
		Reply(500)
	defer gock.Off()

	rc := func(ctx context.Context, body []byte) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodPost, "https://respond.io/webhook", bytes.NewBuffer(body))
	}

	err := retrier.RequestRetry(context.TODO(), []byte(`{"hello": "world"}`), rc)

	assert.EqualError(t, err, "failed to send event to webhook")
}

func TestRetrier_Success(t *testing.T) {
	for _, status := range []int{
		200,
		202,
	} {
		t.Run(strcov.Itoa(status), func(t *testing.T) {
			retrier := NewRetrier(zap.NewNop(), 5*time.Second)

			gock.New("https://respond.io").
				MatchHeader("Content-Type", "application/json").
				Post("/webhook").
				Reply(status)
			defer gock.Off()

			rc := func(ctx context.Context, body []byte) (*http.Request, error) {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://respond.io/webhook", bytes.NewBuffer(body))
				if err != nil {
					return nil, err
				}

				req.Header.Add("Content-Type", "application/json")

				return req, nil
			}

			err := retrier.RequestRetry(context.TODO(), []byte(`{"hello": "world"}`), rc)

			assert.Nil(t, err)
		})
	}
}
