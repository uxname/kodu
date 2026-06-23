// Package tokenizer считает токены и оценивает стоимость контекста.
//
// Паритет с src/shared/tokenizer/tokenizer.service.ts: кодировка o200k_base
// (модель gpt-4o), цена DEFAULT_PRICE_PER_MILLION. Словарь встроен через
// offline-loader — без обращений в сеть на старте.
package tokenizer

import (
	"sync"

	"github.com/pkoukk/tiktoken-go"
	tiktokenloader "github.com/pkoukk/tiktoken-go-loader"
)

// DefaultPricePerMillion — цена за 1M токенов в USD (constants.ts:2).
const DefaultPricePerMillion = 5.0

// Estimate — результат подсчёта.
type Estimate struct {
	Tokens      int
	USDEstimate float64
}

// Tokenizer лениво инициализирует кодировщик и переиспользует его.
type Tokenizer struct {
	once    sync.Once
	encoder *tiktoken.Tiktoken
	initErr error
}

// New создаёт токенизатор (кодировщик инициализируется при первом Count).
func New() *Tokenizer { return &Tokenizer{} }

func (t *Tokenizer) get() (*tiktoken.Tiktoken, error) {
	t.once.Do(func() {
		// Встроенный словарь вместо сетевой загрузки.
		tiktoken.SetBpeLoader(tiktokenloader.NewOfflineLoader())
		enc, err := tiktoken.EncodingForModel("gpt-4o")
		if err != nil {
			enc, err = tiktoken.GetEncoding("o200k_base")
		}
		t.encoder, t.initErr = enc, err
	})
	return t.encoder, t.initErr
}

// Count возвращает число токенов и оценку стоимости для текста.
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
