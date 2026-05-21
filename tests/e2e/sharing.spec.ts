import { expect, test } from '@playwright/test';

import {
  expectAppDownload,
  imageUpload,
  unlockTo,
  uploadFiles,
  uniqueName,
  waitForFile,
} from '../support/helpers';

test('public share opens in a fresh browser context and downloads', async ({ browser, page }) => {
  const imageFile = imageUpload(`${uniqueName('share')}.svg`);

  await unlockTo(page, '/dashboard');
  await uploadFiles(page, [imageFile]);

  await page.goto('/files?view=list');
  const row = await waitForFile(page, imageFile.name);
  await row.locator('[data-entry-menu-trigger="true"]').click();
  await page.getByRole('button', { name: 'Share' }).click();
  await expect(page.getByText('Share Page Settings')).toBeVisible();
  await page.getByRole('button', { name: 'Save Share Settings' }).click();

  const shareLink = page.locator('#share-link-input');
  await expect.poll(async () => (await shareLink.inputValue()).trim()).not.toBe('');
  const shareURL = await shareLink.inputValue();

  const publicContext = await browser.newContext();
  const publicPage = await publicContext.newPage();
  await publicPage.goto(shareURL);

  await expect(publicPage.locator('[data-public-share-name="true"]')).toContainText(imageFile.name);
  await expect(publicPage.locator('.public-share-image')).toBeVisible();

  await expectAppDownload(publicPage, async () => {
    await publicPage.getByRole('link', { name: 'Download File' }).click();
  });

  await publicContext.close();
});
