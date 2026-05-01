import { Injectable } from '@nestjs/common';
import { encodingForModel, getEncoding } from 'js-tiktoken';
import { DEFAULT_PRICE_PER_MILLION } from '../constants';

type TokenEstimate = {
  tokens: number;
  usdEstimate: number;
};

@Injectable()
export class TokenizerService {
  private encoder?: ReturnType<typeof getEncoding>;

  count(text: string): TokenEstimate {
    const tokens = this.getEncoder().encode(text).length;
    const usdEstimate = (tokens / 1_000_000) * DEFAULT_PRICE_PER_MILLION;
    return { tokens, usdEstimate };
  }

  private getEncoder() {
    if (!this.encoder) {
      try {
        this.encoder = encodingForModel('gpt-4o');
      } catch {
        this.encoder = getEncoding('o200k_base');
      }
    }
    return this.encoder;
  }
}
