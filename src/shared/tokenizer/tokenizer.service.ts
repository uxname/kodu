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
    const config = this.configService.getConfig();
    const model = config.llm?.model ?? 'gpt-4o';
    const key = model.toLowerCase();
    return PRICE_PER_MILLION[key] ?? 0;
  }

  private createEncoder() {
    const config = this.configService.getConfig();
    const model = config.llm?.model ?? 'gpt-4o';

    try {
      return encodingForModel(model as TiktokenModel);
    } catch {
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
