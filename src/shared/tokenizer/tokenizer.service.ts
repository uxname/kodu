import { Injectable } from '@nestjs/common';
import { encodingForModel, getEncoding, type TiktokenModel } from 'js-tiktoken';
import { ConfigService } from '../../core/config/config.service';
import { DEFAULT_LLM_MODEL } from '../constants';

type TokenEstimate = {
  tokens: number;
  usdEstimate: number;
};

const PRICE_PER_MILLION: Record<string, number> = {
  'gpt-4o': 10,
  'gpt-5-mini': 2.5,
};

@Injectable()
export class TokenizerService {
  private encoder?: ReturnType<typeof encodingForModel>;

  constructor(private readonly configService: ConfigService) {}

  count(text: string): TokenEstimate {
    const tokens = this.getEncoder().encode(text).length;
    const price = this.getPricePerMillion();
    const usdEstimate = price > 0 ? (tokens / 1_000_000) * price : 0;

    return { tokens, usdEstimate };
  }

  private getPricePerMillion(): number {
    const config = this.configService.getConfig();
    const model = config.llm?.model ?? `openai/${DEFAULT_LLM_MODEL}`;
    const key = this.normalizeModelKey(model);
    return PRICE_PER_MILLION[key] ?? 0;
  }

  private createEncoder() {
    try {
      return encodingForModel(DEFAULT_LLM_MODEL as TiktokenModel);
    } catch {
      return getEncoding('o200k_base');
    }
  }

  private getEncoder() {
    if (!this.encoder) {
      this.encoder = this.createEncoder();
    }

    return this.encoder;
  }

  private normalizeModelKey(model: string): string {
    const lower = model.toLowerCase();
    if (lower.includes('/')) {
      return lower.split('/').pop() ?? lower;
    }
    return lower;
  }
}
