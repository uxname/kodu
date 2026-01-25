import { Agent } from '@mastra/core/agent';
import { Injectable } from '@nestjs/common';
import { z } from 'zod';
import { ConfigService } from '../../core/config/config.service';

export type ReviewMode = 'bug' | 'style' | 'security';

type ReviewIssue = {
  severity: 'low' | 'medium' | 'high';
  file?: string;
  line?: number;
  message: string;
};

export type ReviewResult = {
  summary: string;
  issues: ReviewIssue[];
};

@Injectable()
export class AiService {
  constructor(private readonly configService: ConfigService) {}

  async reviewDiff(
    diff: string,
    mode: ReviewMode,
    structured: boolean,
  ): Promise<{ text: string; structured?: ReviewResult }> {
    const agent = this.createAgent(
      'kodu-review-agent',
      'AI Reviewer for staged git diff. Be concise.',
    );

    const userPrompt = this.buildReviewPrompt(diff, mode);

    if (structured) {
      const schema = z.object({
        summary: z.string(),
        issues: z
          .array(
            z.object({
              severity: z.enum(['low', 'medium', 'high']).default('low'),
              file: z.string().optional(),
              line: z.number().int().positive().optional(),
              message: z.string(),
            }),
          )
          .default([]),
      });

      const output = await agent.generate(userPrompt, {
        structuredOutput: { schema },
      });

      return { text: output.text.trim(), structured: output.object };
    }

    const output = await agent.generate(userPrompt);
    return { text: output.text.trim() };
  }

  async generateCommitMessage(diff: string): Promise<string> {
    const agent = this.createAgent(
      'kodu-commit-agent',
      'Generate a concise Conventional Commit message. Only output the message string.',
    );

    const output = await agent.generate(
      `
You generate Conventional Commit messages.
Rules:
- Format: <type>(<optional scope>): <subject>
- Lowercase subject, no trailing period.
- Keep under 70 characters.
- Summarize the diff accurately.

Diff:
${diff}
`.trim(),
    );

    const raw = output.text.trim();
    const cleaned = this.cleanCommitMessage(raw);

    if (!cleaned) {
      throw new Error('AI не вернул валидное сообщение коммита.');
    }

    return cleaned;
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
      model: { id: modelId as `${string}/${string}`, apiKey },
      maxRetries: 1,
    });
  }

  private buildReviewPrompt(diff: string, mode: ReviewMode): string {
    const focusByMode: Record<ReviewMode, string> = {
      bug: 'Найди потенциальные баги, логические ошибки, регрессы.',
      style: 'Проверь читаемость, согласованность, форматирование и нейминг.',
      security:
        'Найди уязвимости, утечки секретов, неправильные проверки прав.',
    };

    return `
Ты — строгий ревьюер кода. Формат ответа: краткий markdown с пунктами.
Режим: ${mode}. ${focusByMode[mode]}
Дай сжатый список проблем и рекомендаций. Если критичных нет — скажи об этом.

Diff:
${diff}
`.trim();
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
    const normalized = model.includes('/') ? model : `openai/${model}`;
    return normalized;
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
