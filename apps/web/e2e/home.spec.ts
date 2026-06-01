import { test, expect } from '@playwright/test';

test.describe('Home Page', () => {
  test('should load home page with key sections', async ({ page }) => {
    await page.goto('/');

    // Hero banner
    await expect(page.getByText('Welcome to Tiki')).toBeVisible();
    await expect(page.getByText('Discover millions of products from trusted sellers')).toBeVisible();

    // Shop Now button
    await expect(page.getByRole('link', { name: 'Shop Now' })).toBeVisible();

    // Categories section
    await expect(page.getByRole('heading', { name: 'Categories' })).toBeVisible();

    // Flash Deals section
    await expect(page.getByRole('heading', { name: /Flash Deals/i })).toBeVisible();

    // Featured Products section
    await expect(page.getByRole('heading', { name: 'Featured Products' })).toBeVisible();
  });

  test('should have working "See All" links', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByRole('link', { name: 'See All →' }).first()).toBeVisible();
    await expect(page.getByRole('link', { name: 'See All →' }).nth(1)).toBeVisible();
  });

  test('should have header and footer', async ({ page }) => {
    await page.goto('/');

    // Header should be present
    await expect(page.locator('header')).toBeVisible();

    // Footer should be present
    await expect(page.locator('footer')).toBeVisible();
  });
});
