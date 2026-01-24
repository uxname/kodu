import confirm from '@inquirer/confirm';
import input from '@inquirer/input';
import select from '@inquirer/select';
import { Injectable } from '@nestjs/common';
import pc from 'picocolors';
import yoctoSpinner, {
  type Spinner,
  type Options as SpinnerOptions,
} from 'yocto-spinner';

type InputOptions = Parameters<typeof input>[0];
type ConfirmOptions = Parameters<typeof confirm>[0];
type SelectOptions<TValue> = Parameters<typeof select<TValue>>[0];

@Injectable()
export class UiService {
  readonly log = {
    success: (message: string) => console.log(pc.green(`✔ ${message}`)),
    warn: (message: string) => console.log(pc.yellow(`⚠ ${message}`)),
    error: (message: string) => console.log(pc.red(`✖ ${message}`)),
    info: (message: string) => console.log(pc.cyan(`ℹ ${message}`)),
  };

  createSpinner(options?: SpinnerOptions & { text?: string }): Spinner {
    return yoctoSpinner({ text: options?.text ?? '', ...options });
  }

  promptInput(options: InputOptions): Promise<string> {
    return input(options);
  }

  promptConfirm(options: ConfirmOptions): Promise<boolean> {
    return confirm(options);
  }

  promptSelect<TValue>(options: SelectOptions<TValue>): Promise<TValue> {
    return select<TValue>(options);
  }
}
