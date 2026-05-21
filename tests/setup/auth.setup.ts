import { expect, test } from '@playwright/test';

import { authFile, e2eAccount } from '../support/env';
import { bootstrapInstance, ensureLocalStorageSettings, login } from '../support/helpers';

test('bootstrap auth state', async ({ page }) => {
  const mode = await bootstrapInstance(page);
  await login(page);
  if (mode === 'setup') {
    await ensureLocalStorageSettings(page);
  }
  await page.goto('/dashboard');
  await expect(page).toHaveURL(/\/dashboard$/);
  await page.context().storageState({ path: authFile });
  test.info().annotations.push({
    type: 'bootstrap-mode',
    description: `${mode} using ${e2eAccount.email}`,
  });
});
