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
  await expect(page.getByText('Share Link')).toBeVisible();
  await page.getByRole('button', { name: 'Save Link' }).click();

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

test('one-time public share is burned after first view', async ({ browser, page }) => {
  const imageFile = imageUpload(`${uniqueName('share-once')}.svg`);

  await unlockTo(page, '/dashboard');
  await uploadFiles(page, [imageFile]);

  await page.goto('/files?view=list');
  const row = await waitForFile(page, imageFile.name);
  await row.locator('[data-entry-menu-trigger="true"]').click();
  await page.getByRole('button', { name: 'Share' }).click();
  await expect(page.getByText('Share Link')).toBeVisible();
  await page.getByLabel('One-time link').check();
  await page.getByRole('button', { name: 'Save Link' }).click();

  const shareURL = await page.locator('#share-link-input').inputValue();

  const firstContext = await browser.newContext();
  const firstPage = await firstContext.newPage();
  await firstPage.goto(shareURL);
  await firstPage.getByRole('button', { name: 'View File' }).click();
  await expect(firstPage.locator('.public-share-image')).toBeVisible();
  await firstContext.close();

  const secondContext = await browser.newContext();
  const secondPage = await secondContext.newPage();
  const response = await secondPage.goto(shareURL);
  expect(response?.status()).toBe(404);
  await secondContext.close();
});
