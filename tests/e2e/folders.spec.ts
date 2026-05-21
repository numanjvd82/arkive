import { expect, test } from '@playwright/test';

import {
  createFolder,
  fileEntry,
  folderEntry,
  textUpload,
  unlockTo,
  uploadFiles,
  uniqueName,
  waitForFile,
  waitForFolder,
} from '../support/helpers';

test('folder create and mixed bulk delete work', async ({ page }) => {
  const folderName = uniqueName('folder');
  const file = textUpload(`${uniqueName('bulk-file')}.txt`);

  await unlockTo(page, '/files?view=list');
  await createFolder(page, folderName);

  await page.goto('/dashboard');
  await uploadFiles(page, [file]);

  await page.goto('/files?view=list');
  const fileRow = await waitForFile(page, file.name);
  const folderRow = await waitForFolder(page, folderName);
  await expect(fileRow).toBeVisible();
  await expect(folderRow).toBeVisible();

  await folderEntry(page, folderName).getByLabel('Select folder').check();
  await fileEntry(page, file.name).getByLabel('Select file').check();
  await page.evaluate(() => {
    window.ArkiveEntrySelection?.requestDeleteSelected();
  });

  await expect(page.getByRole('heading', { name: 'Delete items?' })).toBeVisible();
  await page.getByRole('button', { name: 'Delete' }).click();

  await expect(folderEntry(page, folderName)).toHaveCount(0);
  await expect(fileEntry(page, file.name)).toHaveCount(0);
});
