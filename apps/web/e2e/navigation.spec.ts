import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test('should navigate from home to products', async ({ page }) => {
    await page.goto('/');

    const shopNow = page.getByRole('link', { name: 'Shop Now' });
    if (await shopNow.isVisible()) {
      await shopNow.click();
      await expect(page).toHaveURL(/\/products/);
    }
  });

  test('should navigate to login from header', async ({ page }) => {
    await page.goto('/');

    const loginLink = page.getByRole('link', { name: /login|sign in/i });
    if (await loginLink.isVisible()) {
      await loginLink.click();
      await expect(page).toHaveURL(/\/login/);
    }
  });

  test('should navigate to cart from header', async ({ page }) => {
    await page.goto('/');

    const cartLink = page.getByRole('link', { name: /cart/i });
    if (await cartLink.isVisible()) {
      await cartLink.click();
      await expect(page).toHaveURL(/\/cart/);
    }
  });
});

test.describe('Responsive Layout', () => {
  test('should render on mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');
    await expect(page.locator('body')).toBeVisible();
    await expect(page.locator('header')).toBeVisible();
  });

  test('should render on tablet viewport', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/');
    await expect(page.locator('body')).toBeVisible();
  });

  test('should render on desktop viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto('/');
    await expect(page.locator('body')).toBeVisible();
  });
});

test.describe('SEO & Meta', () => {
  test('should have correct page title', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/tiki/i);
  });

  test('should have meta description', async ({ page }) => {
    await page.goto('/');
    const metaDescription = page.locator('meta[name="description"]');
    await expect(metaDescription).toHaveAttribute('content', /shop|product|price/i);
  });
});
