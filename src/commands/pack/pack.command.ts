import { promises as fs } from 'node:fs';
import path from 'node:path';
import clipboard from 'clipboardy';
import { Command, CommandRunner, Option } from 'nest-commander';
import { ConfigService } from '../../core/config/config.service';
import { PromptService } from '../../core/config/prompt.service';
import { FsService } from '../../core/file-system/fs.service';
import { UiService } from '../../core/ui/ui.service';
import { CleanerService } from '../../shared/cleaner/cleaner.service';
import { TokenizerService } from '../../shared/tokenizer/tokenizer.service';

type OutputFormat = 'xml' | 'text';

type PackOptions = {
  copy?: boolean;
  template?: string;
  out?: string;
  path?: string[];
  exclude?: string[];
  list?: boolean;
  format?: OutputFormat;
  clean?: boolean;
};

type TemplateContext = {
  context: string;
  fileList: string;
  tokenCount: number;
  usdEstimate: number;
};

@Command({
  name: 'pack',
  description: 'Collect project context into a single file',
})
export class PackCommand extends CommandRunner {
  constructor(
    private readonly ui: UiService,
    private readonly configService: ConfigService,
    private readonly promptService: PromptService,
    private readonly fsService: FsService,
    private readonly tokenizer: TokenizerService,
    private readonly cleaner: CleanerService,
  ) {
    super();
  }

  @Option({ flags: '-c, --copy', description: 'Copy result to clipboard' })
  parseCopy(): boolean {
    return true;
  }

  @Option({
    flags: '-t, --template <name>',
    description: 'Template name from .kodu/prompts',
  })
  parseTemplate(value: string): string {
    return value;
  }

  @Option({
    flags: '-o, --out <path>',
    description: 'Path to save result',
  })
  parseOut(value: string): string {
    return value;
  }

  @Option({
    flags: '-p, --path <path>',
    description: 'Directory or glob to include (repeatable)',
  })
  parsePath(value: string, previous: string[] = []): string[] {
    return [...previous, value];
  }

  @Option({
    flags: '-e, --exclude <pattern>',
    description: 'Additional exclude pattern (repeatable)',
  })
  parseExclude(value: string, previous: string[] = []): string[] {
    return [...previous, value];
  }

  @Option({
    flags: '-l, --list',
    description: 'Print file list only, without content',
  })
  parseList(): boolean {
    return true;
  }

  @Option({
    flags: '--clean',
    description: 'Strip comments in-memory before packing (files not modified)',
  })
  parseClean(): boolean {
    return true;
  }

  @Option({
    flags: '-f, --format <format>',
    description: 'Output format: xml (default) or text',
  })
  parseFormat(value: string): OutputFormat {
    if (value !== 'xml' && value !== 'text') {
      this.ui.log.warn(`Unknown format "${value}", using "xml"`);
      return 'xml';
    }
    return value;
  }

  async run(_inputs: string[], options: PackOptions): Promise<void> {
    const spinner = this.ui
      .createSpinner({ text: 'Collecting files...' })
      .start();

    try {
      const { packer } = this.configService.getConfig();
      const extraExcludes = options.exclude ?? [];
      const files = await this.fsService.findProjectFiles({
        excludeBinary: true,
        useGitignore: packer.useGitignore,
        ignore: [...packer.ignore, ...extraExcludes],
        contentBasedBinaryDetection: packer.contentBasedBinaryDetection,
        rootPaths: options.path,
      });

      if (files.length === 0) {
        spinner.stop('No files to pack.');
        this.ui.log.warn('No files to pack.');
        return;
      }

      if (options.list) {
        spinner.success(`Found ${files.length} files`);
        for (const file of files) {
          this.ui.log.info(file);
        }
        return;
      }

      const format: OutputFormat = options.format ?? 'xml';
      const context = await this.buildContext(files, format, options.clean);
      const fileList = files.join('\n');
      const { tokens, usdEstimate } = this.tokenizer.count(context);

      const basePrompt = await this.applyConfiguredPrompt({
        context,
        fileList,
        tokenCount: tokens,
        usdEstimate,
      });

      const templateApplied = options.template
        ? await this.applyTemplate(options.template, {
            context,
            fileList,
            tokenCount: tokens,
            usdEstimate,
          })
        : basePrompt;

      const outputPath = await this.writeOutput(templateApplied, options.out);

      if (options.copy) {
        await clipboard.write(templateApplied);
      }

      spinner.success('Collection complete');
      this.ui.log.info(`Files: ${files.length}`);
      this.ui.log.info(`Tokens: ${tokens}`);
      this.ui.log.info(`Cost estimate: ~$${usdEstimate.toFixed(4)}`);
      this.ui.log.info(
        `Format: ${format}${options.clean ? ' (comments stripped)' : ''}`,
      );
      this.ui.log.success(`Saved to ${outputPath}`);

      if (options.copy) {
        this.ui.log.success('Result copied to clipboard');
      }
    } catch (error) {
      spinner.error('Error collecting context');
      const message = error instanceof Error ? error.message : 'Unknown error';
      this.ui.log.error(message);
      process.exitCode = 1;
    }
  }

  private async buildContext(
    files: string[],
    format: OutputFormat,
    clean = false,
  ): Promise<string> {
    const chunks = await Promise.all(
      files.map(async (file) => {
        let content = await this.fsService.readFileRelative(file);
        if (clean) {
          content = this.cleaner.cleanContent(file, content);
        }
        if (format === 'xml') {
          return `<file path="${file}">\n${content}\n</file>`;
        }
        return `// file: ${file}\n${content}`;
      }),
    );

    if (format === 'xml') {
      return `<files>\n${chunks.join('\n\n')}\n</files>`;
    }
    return chunks.join('\n\n');
  }

  private async applyTemplate(
    name: string,
    ctx: TemplateContext,
  ): Promise<string> {
    const template = await this.loadTemplate(name);
    return this.fillTemplate(template, ctx);
  }

  private async loadTemplate(name: string): Promise<string> {
    return this.promptService.loadFromPromptsDir(name);
  }

  private async writeOutput(
    content: string,
    outPath?: string,
  ): Promise<string> {
    const target = outPath ?? path.join(process.cwd(), '.kodu', 'context.txt');
    const dir = path.dirname(target);
    await fs.mkdir(dir, { recursive: true });
    await fs.writeFile(target, `${content}\n`, 'utf8');
    return target;
  }

  private async applyConfiguredPrompt(ctx: TemplateContext): Promise<string> {
    const config = this.configService.getConfig();
    const packPrompt = config.prompts?.pack;

    if (!packPrompt) {
      return ctx.context;
    }

    try {
      const template = await this.promptService.load(packPrompt);
      return this.fillTemplate(template, ctx);
    } catch {
      this.ui.log.warn(
        `Prompt file not found: ${packPrompt}, using raw context`,
      );
      return ctx.context;
    }
  }

  private fillTemplate(template: string, ctx: TemplateContext): string {
    const filled = template
      .replace(/\{\{context\}\}/g, ctx.context)
      .replace(/\{\{fileList\}\}/g, ctx.fileList)
      .replace(/\{\{tokenCount\}\}/g, ctx.tokenCount.toString())
      .replace(/\{\{usdEstimate\}\}/g, ctx.usdEstimate.toFixed(4));

    if (!template.includes('{{context}}')) {
      return `${filled}\n\n${ctx.context}`;
    }

    return filled;
  }
}
