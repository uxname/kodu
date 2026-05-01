import path from 'node:path';
import { Injectable } from '@nestjs/common';
import { Project } from 'ts-morph';

export type DepsResult = {
  files: string[];
  explain: Map<string, string>;
};

type CollectOptions = {
  maxDepth?: number;
  includeTypes?: boolean;
  includeDynamic?: boolean;
};

@Injectable()
export class DepsService {
  collectDependencies(
    entryFiles: string[],
    projectRoot: string,
    options: CollectOptions = {},
  ): DepsResult {
    const {
      maxDepth = Infinity,
      includeTypes = true,
      includeDynamic = false,
    } = options;

    const tsConfigPath = this.findTsConfig(projectRoot);
    const project = tsConfigPath
      ? new Project({
          tsConfigFilePath: tsConfigPath,
          skipAddingFilesFromTsConfig: true,
        })
      : new Project({
          compilerOptions: {
            allowJs: true,
            resolveJsonModule: true,
            moduleResolution: 2, // NodeJs
          },
        });

    const visited = new Set<string>();
    const explain = new Map<string, string>();

    const absEntries = entryFiles.map((f) =>
      path.isAbsolute(f) ? f : path.resolve(projectRoot, f),
    );

    for (const entry of absEntries) {
      explain.set(entry, 'entry point');
      this.collect(
        project,
        entry,
        projectRoot,
        visited,
        explain,
        0,
        maxDepth,
        includeTypes,
        includeDynamic,
      );
    }

    const files = [...visited].map((abs) =>
      path.relative(projectRoot, abs).split(path.sep).join(path.posix.sep),
    );

    return { files, explain };
  }

  private collect(
    project: Project,
    absFile: string,
    projectRoot: string,
    visited: Set<string>,
    explain: Map<string, string>,
    depth: number,
    maxDepth: number,
    includeTypes: boolean,
    includeDynamic: boolean,
  ): void {
    if (visited.has(absFile)) return;
    visited.add(absFile);

    if (depth >= maxDepth) return;

    let sourceFile = project.getSourceFile(absFile);
    if (!sourceFile) {
      try {
        sourceFile = project.addSourceFileAtPath(absFile);
      } catch {
        return;
      }
    }

    const relFrom = path
      .relative(projectRoot, absFile)
      .split(path.sep)
      .join(path.posix.sep);

    for (const importDecl of sourceFile.getImportDeclarations()) {
      if (!includeTypes && importDecl.isTypeOnly()) continue;

      const resolved = importDecl.getModuleSpecifierSourceFile();
      if (!resolved) continue;

      const absResolved = resolved.getFilePath();
      if (absResolved.includes('node_modules')) continue;

      if (!explain.has(absResolved)) {
        const what = importDecl.isTypeOnly() ? 'type import' : 'import';
        explain.set(absResolved, `${what} from ${relFrom}`);
      }

      this.collect(
        project,
        absResolved,
        projectRoot,
        visited,
        explain,
        depth + 1,
        maxDepth,
        includeTypes,
        includeDynamic,
      );
    }

    void includeDynamic;

    for (const exportDecl of sourceFile.getExportDeclarations()) {
      const resolved = exportDecl.getModuleSpecifierSourceFile();
      if (!resolved) continue;

      const absResolved = resolved.getFilePath();
      if (absResolved.includes('node_modules')) continue;

      if (!explain.has(absResolved)) {
        const relFrom2 = path
          .relative(projectRoot, absFile)
          .split(path.sep)
          .join(path.posix.sep);
        explain.set(absResolved, `re-export from ${relFrom2}`);
      }

      this.collect(
        project,
        absResolved,
        projectRoot,
        visited,
        explain,
        depth + 1,
        maxDepth,
        includeTypes,
        includeDynamic,
      );
    }
  }

  private findTsConfig(projectRoot: string): string | undefined {
    const candidates = [
      path.join(projectRoot, 'tsconfig.json'),
      path.join(projectRoot, 'tsconfig.base.json'),
    ];
    for (const c of candidates) {
      try {
        require('node:fs').accessSync(c);
        return c;
      } catch {
        // not found
      }
    }
    return undefined;
  }
}
