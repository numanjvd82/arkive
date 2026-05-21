import { expect, test } from '@playwright/test';

import {
  expectAppDownload,
  imageUpload,
  openFileFromList,
  switchToGrid,
  switchToList,
  unlockTo,
  uploadFiles,
  uniqueName,
  waitForFile,
} from '../support/helpers';

test('file listing supports list and grid view and image preview/download', async ({ page }) => {
  const imageFile = imageUpload(`${uniqueName('gallery')}.svg`);

  await unlockTo(page, '/dashboard');
  await uploadFiles(page, [imageFile]);

  await page.goto('/files?view=list');
  await expect(await waitForFile(page, imageFile.name)).toBeVisible();

  await switchToGrid(page);
  await expect(page.locator(`[data-file-card][data-file-name="${imageFile.name}"]`).first()).toBeVisible();

  await switchToList(page);
  await openFileFromList(page, imageFile.name);
  await expect(page.locator('.media-image')).toBeVisible();

  await expectAppDownload(page, async () => {
    await page.getByRole('button', { name: 'Download File' }).click();
  });
});
