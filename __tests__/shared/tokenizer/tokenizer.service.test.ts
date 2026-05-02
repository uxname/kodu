import { beforeEach, describe, expect, it } from 'vitest';

describe('TokenizerService', () => {
  let tokenizer: {
    count: (text: string) => { tokens: number; usdEstimate: number };
  };

  beforeEach(async () => {
    const { TokenizerService } = await import(
      '../../../src/shared/tokenizer/tokenizer.service'
    );
    tokenizer = new TokenizerService();
  });

  it('should count tokens for empty string', () => {
    const result = tokenizer.count('');

    expect(result.tokens).toBe(0);
    expect(result.usdEstimate).toBe(0);
  });

  it('should count tokens for simple text', () => {
    const result = tokenizer.count('hello world');

    expect(result.tokens).toBeGreaterThan(0);
    expect(result.usdEstimate).toBeGreaterThan(0);
  });

  it('should estimate cost correctly based on DEFAULT_PRICE_PER_MILLION', () => {
    const result = tokenizer.count('a'.repeat(1000));

    expect(result.tokens).toBeGreaterThan(0);
    // DEFAULT_PRICE_PER_MILLION = 5 means $5 per 1M tokens
    const expectedCost = (result.tokens / 1_000_000) * 5;
    expect(result.usdEstimate).toBe(expectedCost);
  });

  it('should handle large text', () => {
    const longText = 'test '.repeat(10000);
    const result = tokenizer.count(longText);

    expect(result.tokens).toBeGreaterThan(10000);
    expect(result.usdEstimate).toBeGreaterThan(0);
  });
});
