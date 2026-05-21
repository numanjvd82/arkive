import { defineConfig, devices } from '@playwright/test';

import { authFile, baseURL } from './tests/support/env';

const isCI = !!process.env.CI;
const headless = process.env.PW_TEST_HEADLESS === '0' ? false : true;

export default defineConfig({
  testDir: './tests',
  timeout: 90_000,
  expect: {
    timeout: 15_000,
  },
  fullyParallel: false,
  forbidOnly: isCI,
  retries: isCI ? 2 : 0,
  workers: 1,
  reporter: isCI
    ? [['github'], ['html', { open: 'never' }]]
    : [['list'], ['html', { open: 'never' }]],
  globalSetup: './tests/global.setup.ts',
  use: {
    baseURL,
    colorScheme: 'dark',
    headless,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'setup',
      testMatch: /tests\/setup\/.*\.setup\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'chromium',
      dependencies: ['setup'],
      grepInvert: /@mobile/,
      testIgnore: /tests\/setup\//,
      use: {
        ...devices['Desktop Chrome'],
        storageState: authFile,
      },
    },
    {
      name: 'firefox',
      dependencies: ['setup'],
      grepInvert: /@mobile/,
      testIgnore: /tests\/setup\//,
      use: {
        ...devices['Desktop Firefox'],
        storageState: authFile,
      },
    },
    {
      name: 'webkit',
      dependencies: ['setup'],
      grepInvert: /@mobile/,
      testIgnore: /tests\/setup\//,
      use: {
        ...devices['Desktop Safari'],
        storageState: authFile,
      },
    },
    {
      name: 'mobile-chromium',
      dependencies: ['setup'],
      grep: /@mobile/,
      testIgnore: /tests\/setup\//,
      use: {
        ...devices['Pixel 7'],
      },
    },
  ],
});
