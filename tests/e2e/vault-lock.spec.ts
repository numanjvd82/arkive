import { expect, test } from '@playwright/test';

import { e2eAccount } from '../support/env';
import { unlockTo } from '../support/helpers';

test('manual lock redirects to /lock and can re-unlock', async ({ page }) => {
  await unlockTo(page, '/files?view=list');
  await page.getByRole('button', { name: 'Lock vault' }).click();
  await expect(page).toHaveURL(/\/lock\?next=%2Ffiles%3Fview%3Dlist$/);

  await page.waitForFunction(() => {
    return !!window.ArkiveVault && typeof window.ArkiveVault.unlockVault === 'function';
  });
  await page.locator('input[name="password"]').fill(e2eAccount.password);
  await page.getByRole('button', { name: 'Unlock Vault' }).click();
  await expect(page).toHaveURL(/\/files\?view=list$/);
});
