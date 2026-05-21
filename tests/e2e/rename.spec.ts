import { expect, test } from '@playwright/test';

import {
  createFolder,
  openFileFromList,
  textUpload,
  unlockTo,
  uploadFiles,
  uniqueName,
  waitForFile,
  waitForFolder,
} from '../support/helpers';

test('folder and file rename survive refresh and file preview', async ({ page }) => {
  const originalFolderName = uniqueName('rename-folder');
  const renamedFolderName = uniqueName('renamed-folder');
  const originalFileName = `${uniqueName('rename-file')}.txt`;
  const renamedFileBase = uniqueName('renamed-file');
  const renamedFileName = `${renamedFileBase}.txt`;
  const file = textUpload(originalFileName);

  await unlockTo(page, '/files?view=list');

  await createFolder(page, originalFolderName);
  const folderRow = await waitForFolder(page, originalFolderName);
  await folderRow.click();
  await page.keyboard.press('F2');
  const folderRenameInput = page.locator('.files-rename-input');
  await expect(folderRenameInput).toBeVisible();
  await folderRenameInput.fill(renamedFolderName);
  await folderRenameInput.press('Enter');
  await expect(await waitForFolder(page, renamedFolderName)).toBeVisible();
  await page.reload();
  await expect(await waitForFolder(page, renamedFolderName)).toBeVisible();

  await page.goto('/dashboard');
  await uploadFiles(page, [file]);

  await page.goto('/files?view=list');
  const fileRow = await waitForFile(page, originalFileName);
  await fileRow.click();
  await page.keyboard.press('F2');
  const fileRenameInput = page.locator('.files-rename-input');
  await expect(fileRenameInput).toBeVisible();
  await expect(fileRenameInput).toHaveValue(originalFileName.replace(/\.txt$/, ''));
  await fileRenameInput.fill(renamedFileBase);
  await fileRenameInput.press('Enter');

  await expect(await waitForFile(page, renamedFileName)).toBeVisible();
  await page.reload();
  await expect(await waitForFile(page, renamedFileName)).toBeVisible();

  await openFileFromList(page, renamedFileName);
  await expect(page.locator('[data-media-title="true"]')).toContainText(renamedFileName);
});
