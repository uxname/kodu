import { Agent } from '@mastra/core/agent';
import { Injectable } from '@nestjs/common';
import { ConfigService } from '../../core/config/config.service';
import {
  DEFAULT_COMMIT_PROMPT,
  DEFAULT_REVIEW_PROMPTS,
  replacePromptVariables,
  STANDARD_REVIEW_MODES,
} from '../../core/config/default-prompts';

export type ReviewMode = string;

@Injectable()
export class AiService {
  constructor(private readonly configService: ConfigService) {}

  async reviewDiff(diff: string, mode: ReviewMode): Promise<{ text: string }> {
    const agent = this.createAgent(
      'kodu-review-agent',
      'AI Reviewer for staged git diff. Be concise.',
    );

    const userPrompt = this.buildReviewPrompt(diff, mode);

    const output = await agent.generate(userPrompt);
    return { text: output.text.trim() };
  }

  async generateCommitMessage(diff: string): Promise<string> {
    const agent = this.createAgent(
      'kodu-commit-agent',
      'Generate a concise Conventional Commit message. Only output the message string.',
    );

    const prompt = this.buildCommitPrompt(diff);
    const output = await agent.generate(prompt);

    const raw = output.text.trim();
    const cleaned = this.cleanCommitMessage(raw);

    if (!cleaned) {
      throw new Error('AI не вернул валидное сообщение коммита.');
    }

    return cleaned;
  }

  getAvailableReviewModes(): string[] {
    const config = this.configService.getConfig();
    const standardModes = [...STANDARD_REVIEW_MODES];
    const customModes = config.prompts?.review
      ? Object.keys(config.prompts.review)
      : [];

    return Array.from(new Set([...standardModes, ...customModes]));
  }

  hasApiKey(): boolean {
    const config = this.configService.getConfig();
    if (!config.llm) {
      return false;
    }
    const envName = config.llm.apiKeyEnv ?? 'OPENAI_API_KEY';
    const value = process.env[envName];
    return Boolean(value);
  }

  getApiKeyEnvName(): string {
    const config = this.configService.getConfig();
    return config.llm?.apiKeyEnv ?? 'OPENAI_API_KEY';
  }

  private createAgent(id: string, instructions: string): Agent {
    const apiKey = this.getApiKey();
    const modelId = this.getModelId();

    return new Agent({
      id,
      name: id,
      instructions,
      model: {
        id: modelId as `${string}/${string}`,
        apiKey,
      },
      maxRetries: 1,
    });
  }

  private buildReviewPrompt(diff: string, mode: ReviewMode): string {
    const config = this.configService.getConfig();

    const customPrompt = config.prompts?.review?.[mode];
    if (customPrompt) {
      return replacePromptVariables(customPrompt, { diff, mode });
    }

    if (
      STANDARD_REVIEW_MODES.includes(
        mode as (typeof STANDARD_REVIEW_MODES)[number],
      )
    ) {
      const defaultPrompt =
        DEFAULT_REVIEW_PROMPTS[mode as keyof typeof DEFAULT_REVIEW_PROMPTS];
      return replacePromptVariables(defaultPrompt, { diff, mode });
    }

    return replacePromptVariables(DEFAULT_REVIEW_PROMPTS.bug, { diff, mode });
  }

  private buildCommitPrompt(diff: string): string {
    const config = this.configService.getConfig();
    const customPrompt = config.prompts?.commit;

    if (customPrompt) {
      return replacePromptVariables(customPrompt, { diff });
    }

    return replacePromptVariables(DEFAULT_COMMIT_PROMPT, { diff });
  }

  private getApiKey(): string {
    const config = this.configService.getConfig();
    if (!config.llm) {
      throw new Error(
        'LLM конфигурация не найдена. Добавьте секцию llm в kodu.json',
      );
    }
    const envName = config.llm.apiKeyEnv ?? 'OPENAI_API_KEY';
    const value = process.env[envName];

    if (!value) {
      throw new Error(`Не найден API ключ: установите ${envName} в окружении.`);
    }

    return value;
  }

  private getModelId(): string {
    const config = this.configService.getConfig();
    if (!config.llm) {
      throw new Error(
        'LLM конфигурация не найдена. Добавьте секцию llm в kodu.json',
      );
    }
    const model = config.llm.model;

    // Model should already be in provider/model format, but validate
    if (!model.includes('/')) {
      throw new Error(
        `Неверный формат модели: "${model}". Ожидается формат "provider/model-name" (например, "openai/gpt-4o")`,
      );
    }

    return model;
  }

  private cleanCommitMessage(text: string): string {
    const unfenced = text.replace(/^```[a-zA-Z]*\s*/g, '').replace(/```$/g, '');
    const lines = unfenced
      .split('\n')
      .map((line) => line.trim())
      .filter((line) => line.length > 0);

    if (lines.length === 0) {
      return '';
    }

    const first = lines[0]
      .replace(/^"|"$/g, '')
      .replace(/^'|'$/g, '')
      .replace(/^Commit message:?\s*/i, '')
      .trim();

    return first;
  }
}
