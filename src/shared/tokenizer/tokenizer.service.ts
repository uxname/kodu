import { Injectable } from '@nestjs/common';
import { encodingForModel, type TiktokenModel } from 'js-tiktoken';
import { ConfigService } from '../../core/config/config.service';

type TokenEstimate = {
  tokens: number;
  usdEstimate: number;
};

const PRICE_PER_MILLION: Record<string, number> = {
  'gpt-4o': 2.5,
  'gpt-4o-mini': 0.15,
  o1: 15,
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
    const model = this.configService.getConfig().llm.model;
    const key = model.toLowerCase();
    return PRICE_PER_MILLION[key] ?? 0;
  }

  private createEncoder() {
    const model = this.configService.getConfig().llm.model;

    try {
      return encodingForModel(model as TiktokenModel);
    } catch {
      // fallback to o200k_base which covers 4o/4o-mini/o1 families
      return encodingForModel('gpt-4o');
    }
  }

  private getEncoder() {
    if (!this.encoder) {
      this.encoder = this.createEncoder();
    }

    return this.encoder;
  }
}
