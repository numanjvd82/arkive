import { expect, test } from '@playwright/test';

import { imageUpload, textUpload, unlockTo, uploadFiles, uniqueName, waitForFile } from '../support/helpers';

test('small files upload successfully from the dashboard', async ({ page }) => {
  const textFile = textUpload(`${uniqueName('note')}.txt`);
  const imageFile = imageUpload(`${uniqueName('image')}.svg`);

  await unlockTo(page, '/dashboard');
  await uploadFiles(page, [textFile, imageFile]);

  await page.goto('/files?view=list');
  await expect(await waitForFile(page, textFile.name)).toBeVisible();
  await expect(await waitForFile(page, imageFile.name)).toBeVisible();
});
