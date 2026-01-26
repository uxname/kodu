export const DEFAULT_REVIEW_PROMPTS = {
  bug: `You are a strict code reviewer. Response format: concise markdown with bullet points.
Mode: {mode}. Find potential bugs, logical errors, and regressions.
Provide a concise list of issues and recommendations. If no critical issues found, state that.

Diff:
{diff}`,

  style: `You are a strict code reviewer. Response format: concise markdown with bullet points.
Mode: {mode}. Check readability, consistency, formatting, and naming.
Provide a concise list of issues and recommendations. If no critical issues found, state that.

Diff:
{diff}`,

  security: `You are a strict code reviewer. Response format: concise markdown with bullet points.
Mode: {mode}. Find vulnerabilities, secret leaks, improper permission checks.
Provide a concise list of issues and recommendations. If no critical issues found, state that.

Diff:
{diff}`,
} as const;

export const DEFAULT_COMMIT_PROMPT = `You generate Conventional Commit messages.
Rules:
- Format: <type>(<optional scope>): <subject>
- Lowercase subject, no trailing period.
- Keep under 70 characters.
- Summarize the diff accurately.

Diff:
{diff}`;

export const DEFAULT_PACK_PROMPT = `Project context below. Keep instructions concise.

Files packed ({{tokenCount}} tokens, ~{{usdEstimate}}):
{{fileList}}

Context:
{{context}}`;

export const STANDARD_REVIEW_MODES = ['bug', 'style', 'security'] as const;

export function replacePromptVariables(
  prompt: string,
  variables: Record<string, string>,
): string {
  let result = prompt;
  for (const [key, value] of Object.entries(variables)) {
    result = result.replace(new RegExp(`\\{${key}\\}`, 'g'), value);
  }
  return result;
}
