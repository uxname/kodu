export type CleanOptions = {
  dryRun?: boolean;
  backup?: boolean;
  keepJSDoc?: boolean;
  onProgress?: (current: number, total: number) => void;
};

export type FileCleanReport = {
  file: string;
  removed: number;
  previews: string[];
  bytesBefore: number;
  bytesAfter: number;
};

export type CleanSummary = {
  filesProcessed: number;
  filesChanged: number;
  commentsRemoved: number;
  bytesBefore: number;
  bytesAfter: number;
  reports: FileCleanReport[];
};
