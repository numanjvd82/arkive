import { expect, test } from '@playwright/test';

import { login, logout } from '../support/helpers';

test.describe('guest flows', () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test('guest pages load and root redirects to login', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/login$/);
    await expect(page.getByRole('heading', { name: 'Arkive Core' })).toBeVisible();
  });

  test('login and logout work', async ({ page }) => {
    await login(page);
    await logout(page);
  });
});

test.describe('mobile smoke', () => {
  test.use({
    storageState: { cookies: [], origins: [] },
  });

  test('@mobile login page renders on a mobile viewport', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByRole('heading', { name: 'Arkive Core' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Unlock Vault' })).toBeVisible();
  });
});
