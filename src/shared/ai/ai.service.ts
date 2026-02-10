import { Agent } from '@mastra/core/agent';
import { Injectable } from '@nestjs/common';
import { ConfigService } from '../../core/config/config.service';
import {
  DEFAULT_COMMIT_PROMPT,
  DEFAULT_REVIEW_PROMPTS,
  replacePromptVariables,
  STANDARD_REVIEW_MODES,
} from '../../core/config/default-prompts';
import { PromptService } from '../../core/config/prompt.service';
import { DEFAULT_COMMIT_TOKENS, DEFAULT_REVIEW_TOKENS } from '../constants';

export type ReviewMode = string;

type ModelSettings = Record<string, unknown> & {
  maxOutputTokens?: number;
};

@Injectable()
export class AiService {
  constructor(
    private readonly configService: ConfigService,
    private readonly promptService: PromptService,
  ) {}

  async reviewDiff(diff: string, mode: ReviewMode): Promise<{ text: string }> {
    const agent = this.createAgent(
      'kodu-review-agent',
      'AI Reviewer for staged git diff. Be concise.',
    );

    const userPrompt = await this.buildReviewPrompt(diff, mode);

    const modelSettings = this.getModelSettingsForCommand('review');
    const output = await agent.generate(userPrompt, { modelSettings });
    return { text: output.text.trim() };
  }

  async generateCommitMessage(diff: string): Promise<string> {
    const agent = this.createAgent(
      'kodu-commit-agent',
      'Generate a concise Conventional Commit message. Only output the message string.',
    );

    const prompt = await this.buildCommitPrompt(diff);
    const modelSettings = this.getModelSettingsForCommand('commit');
    const output = await agent.generate(prompt, { modelSettings });

    const raw = output.text.trim();
    const cleaned = this.cleanCommitMessage(raw);

    if (!cleaned) {
      throw new Error('AI did not return a valid commit message.');
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

  private async buildReviewPrompt(
    diff: string,
    mode: ReviewMode,
  ): Promise<string> {
    const config = this.configService.getConfig();

    const customPrompt = config.prompts?.review?.[mode];
    if (customPrompt) {
      return this.promptService.load(customPrompt, { diff, mode });
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

  private async buildCommitPrompt(diff: string): Promise<string> {
    const config = this.configService.getConfig();
    const customPrompt = config.prompts?.commit;

    if (customPrompt) {
      return this.promptService.load(customPrompt, { diff });
    }

    return replacePromptVariables(DEFAULT_COMMIT_PROMPT, { diff });
  }

  private getApiKey(): string {
    const config = this.configService.getConfig();
    if (!config.llm) {
      throw new Error(
        'LLM configuration not found. Add llm section to kodu.json',
      );
    }
    const envName = config.llm.apiKeyEnv ?? 'OPENAI_API_KEY';
    const value = process.env[envName];

    if (!value) {
      throw new Error(`API key not found: set ${envName} in environment.`);
    }

    return value;
  }

  private getModelId(): string {
    const config = this.configService.getConfig();
    if (!config.llm) {
      throw new Error(
        'LLM configuration not found. Add llm section to kodu.json',
      );
    }
    const model = config.llm.model;

    // Model should already be in provider/model format, but validate
    if (!model.includes('/')) {
      throw new Error(
        `Invalid model format: "${model}". Expected format "provider/model-name" (e.g., "openai/gpt-4o")`,
      );
    }

    return model;
  }

  private getModelSettingsForCommand(
    command: 'commit' | 'review',
  ): ModelSettings {
    const config = this.configService.getConfig();
    const commands = config.llm?.commands;
    const defaultMax =
      command === 'commit' ? DEFAULT_COMMIT_TOKENS : DEFAULT_REVIEW_TOKENS;
    const base: ModelSettings = { maxOutputTokens: defaultMax };

    if (!commands) {
      return base;
    }

    const override = commands[command];
    const settings = override?.modelSettings as ModelSettings | undefined;

    if (!settings) {
      return base;
    }

    return {
      ...settings,
      maxOutputTokens: settings.maxOutputTokens ?? defaultMax,
    };
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
