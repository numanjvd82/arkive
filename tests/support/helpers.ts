import { readFileSync } from 'node:fs';
import path from 'node:path';
import { expect, type Locator, type Page } from '@playwright/test';

import { e2eAccount, storageDir } from './env';

export const fixture = {
  image: path.join(process.cwd(), 'tests', 'fixtures', 'tiny-image.svg'),
  text: path.join(process.cwd(), 'tests', 'fixtures', 'tiny-note.txt'),
};

export type UploadFile = {
  name: string;
  mimeType: string;
  buffer: Buffer;
};

export function uniqueName(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export async function bootstrapInstance(page: Page) {
  await page.goto('/setup');
  await page.waitForLoadState('domcontentloaded');

  if (page.url().includes('/login')) {
    return 'login';
  }

  await expect(page.getByRole('heading', { name: 'Arkive Core Initialization' })).toBeVisible();

  await page.getByLabel('Instance name').fill(e2eAccount.brandName);
  await page.getByLabel('Admin email').fill(e2eAccount.email);
  await page.getByLabel('Master password').fill(e2eAccount.password);
  await page.getByLabel('Confirm password').fill(e2eAccount.password);
  await page.getByLabel('Storage limit in GB').fill('0');
  await page.getByLabel('Storage path').fill(storageDir);

  await page.getByRole('button', { name: 'Initialize Arkive Core' }).click();
  await expect(page).toHaveURL(/\/setup\/recovery-key$/);

  const recoveryValue = page.locator('[data-recovery-value="true"]');
  await expect.poll(async () => (await recoveryValue.inputValue()).trim().length).toBeGreaterThan(0);

  await page.getByLabel(/I have written down/i).check();
  await page.getByRole('button', { name: 'Confirm & Secure Vault' }).click();
  await expect(page).toHaveURL(/\/login\?msg=account-created$/);
  return 'setup';
}

export async function login(page: Page) {
  await page.goto('/login');
  await expect(page.getByRole('heading', { name: 'Arkive Core' })).toBeVisible();
  await page.getByLabel('Email').fill(e2eAccount.email);
  await page.locator('input[name="password"]').fill(e2eAccount.password);
  await page.getByRole('button', { name: 'Unlock Vault' }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
}

export async function unlockTo(page: Page, nextPath: string) {
  await page.goto(`/lock?next=${encodeURIComponent(nextPath)}`);
  await expect(page.getByRole('heading', { name: 'Arkive Vault Locked' })).toBeVisible();
  await page.locator('input[name="password"]').fill(e2eAccount.password);
  await page.getByRole('button', { name: 'Unlock Vault' }).click();
  await expect(page).toHaveURL(new RegExp(escapeForRegExp(nextPath) + '$'));
}

export async function logout(page: Page) {
  await page.getByRole('button', { name: 'Open account menu' }).click();
  await page.getByRole('menuitem', { name: 'Logout' }).click();
  await expect(page).toHaveURL(/\/login$/);
}

export async function ensureLocalStorageSettings(page: Page) {
  await page.goto('/settings#settings-provider');
  await expect(page.getByRole('heading', { name: 'Storage Provider' })).toBeVisible();
  await page.getByRole('radio', { name: 'Local' }).check();
  await page.getByLabel('Storage limit in GB').fill('0');
  await page.getByLabel('Local path').fill(storageDir);
  await page.getByRole('button', { name: 'Save storage settings' }).click();
  await expect(page).toHaveURL(/\/settings\?msg=storage-updated$/);
}

export async function createFolder(page: Page, name: string) {
  await page.getByRole('button', { name: 'New Folder' }).click();
  await page.getByLabel('Folder name').fill(name);
  await page.getByRole('button', { name: 'Create folder' }).click();
  await waitForFolder(page, name);
}

export async function uploadFiles(page: Page, files: UploadFile[]) {
  await page.setInputFiles('#upload-file', files);
  await page.getByRole('button', { name: 'Start upload' }).click();

  for (const file of files) {
    const fileName = file.name;
    const queueItem = page.locator('#upload-queue-list li', { hasText: fileName });
    await expect(queueItem).toBeVisible();
    await expect(queueItem.locator('.queue-item-badge')).toHaveText('completed', { timeout: 60_000 });
  }
}

export function fileEntry(page: Page, name: string): Locator {
  return page.locator(`[data-file-item][data-file-name="${escapeAttribute(name)}"]`).first();
}

export function folderEntry(page: Page, name: string): Locator {
  return page.locator(`[data-folder-item][data-folder-name="${escapeAttribute(name)}"]`).first();
}

export async function openFileFromList(page: Page, name: string) {
  const entry = await waitForFile(page, name);
  await entry.dblclick();
  await expect(page.locator('[data-media-title="true"]')).toContainText(name);
}

export async function openFolderFromList(page: Page, name: string) {
  const entry = await waitForFolder(page, name);
  await entry.dblclick();
}

export async function expectAppDownload(page: Page, trigger: () => Promise<void>) {
  const statusBadge = page.locator('#download-status-badge');
  await Promise.all([
    expect(statusBadge).toHaveText('complete', { timeout: 60_000 }),
    trigger(),
  ]);
}

export function imageUpload(name: string): UploadFile {
  return {
    name,
    mimeType: 'image/svg+xml',
    buffer: readFileSync(fixture.image),
  };
}

export function textUpload(name: string): UploadFile {
  return {
    name,
    mimeType: 'text/plain',
    buffer: readFileSync(fixture.text),
  };
}

export async function switchToGrid(page: Page) {
  await page.getByLabel('Grid view').click();
  await expect(page).toHaveURL(/view=grid/);
}

export async function switchToList(page: Page) {
  await page.getByLabel('List view').click();
  await expect(page).toHaveURL(/view=list/);
}

function escapeForRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function escapeAttribute(value: string) {
  return value.replace(/\\/g, '\\\\').replace(/"/g, '\\"');
}

export async function waitForFile(page: Page, name: string) {
  await expect
    .poll(async () => page.locator(`[data-file-item][data-file-name="${escapeAttribute(name)}"]`).count(), {
      timeout: 60_000,
    })
    .toBeGreaterThan(0);
  return fileEntry(page, name);
}

export async function waitForFolder(page: Page, name: string) {
  await expect
    .poll(async () => page.locator(`[data-folder-item][data-folder-name="${escapeAttribute(name)}"]`).count(), {
      timeout: 60_000,
    })
    .toBeGreaterThan(0);
  return folderEntry(page, name);
}
