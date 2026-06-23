// Package tokenizer counts tokens and estimates the cost of a context.
//
// Parity with src/shared/tokenizer/tokenizer.service.ts: the o200k_base encoding
// (the gpt-4o model), priced at DEFAULT_PRICE_PER_MILLION. The vocabulary is embedded
// via an offline loader — no network access at startup.
package tokenizer

import (
	"sync"

	"github.com/pkoukk/tiktoken-go"
	tiktokenloader "github.com/pkoukk/tiktoken-go-loader"
)

// DefaultPricePerMillion — the price per 1M tokens in USD (constants.ts:2).
const DefaultPricePerMillion = 5.0

// Estimate — the result of a count.
type Estimate struct {
	Tokens      int
	USDEstimate float64
}

// Tokenizer lazily initializes the encoder and reuses it.
type Tokenizer struct {
	once    sync.Once
	encoder *tiktoken.Tiktoken
	initErr error
}

// New creates a tokenizer (the encoder is initialized on the first Count).
func New() *Tokenizer { return &Tokenizer{} }

func (t *Tokenizer) get() (*tiktoken.Tiktoken, error) {
	t.once.Do(func() {
		// Embedded vocabulary instead of a network download.
		tiktoken.SetBpeLoader(tiktokenloader.NewOfflineLoader())
		enc, err := tiktoken.EncodingForModel("gpt-4o")
		if err != nil {
			enc, err = tiktoken.GetEncoding("o200k_base")
		}
		t.encoder, t.initErr = enc, err
	})
	return t.encoder, t.initErr
}

// Count returns the token count and a cost estimate for the text.
func (t *Tokenizer) Count(text string) (Estimate, error) {
	enc, err := t.get()
	if err != nil {
		return Estimate{}, err
	}
	tokens := len(enc.Encode(text, nil, nil))
	return Estimate{
		Tokens:      tokens,
		USDEstimate: (float64(tokens) / 1_000_000) * DefaultPricePerMillion,
	}, nil
}
