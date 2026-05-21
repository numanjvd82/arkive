import { expect, test } from '@playwright/test';

import { imageUpload, openFileFromList, unlockTo, uploadFiles, uniqueName } from '../support/helpers';

test('image preview opens after encrypted file metadata decrypts', async ({ page }) => {
  const imageFile = imageUpload(`${uniqueName('preview')}.svg`);

  await unlockTo(page, '/dashboard');
  await uploadFiles(page, [imageFile]);

  await page.goto('/files?view=list');
  await openFileFromList(page, imageFile.name);
  await expect(page.locator('.media-image')).toBeVisible({ timeout: 60_000 });
  await expect(page.locator('[data-media-title="true"]')).toContainText(imageFile.name);
});
