import { mkdirSync, rmSync } from 'node:fs';

import { authDir, storageDir } from './support/env';

export default async function globalSetup() {
  rmSync(authDir, { force: true, recursive: true });
  rmSync(storageDir, { force: true, recursive: true });
  mkdirSync(authDir, { recursive: true });
  mkdirSync(storageDir, { recursive: true });
}
