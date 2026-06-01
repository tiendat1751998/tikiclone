import { Redis } from "ioredis";

const redis = new Redis({
  host: process.env.REDIS_HOST || "localhost",
  port: parseInt(process.env.REDIS_PORT || "6379", 10),
  password: process.env.REDIS_PASSWORD,
  keyPrefix: "page-cache:",
});

const CACHE_TTL_SECONDS = 30;

export function getCacheKey(path: string): string {
  return `page:${path}`;
}

export async function getCachedPage(path: string): Promise<string | null> {
  try {
    const cached = await redis.get(path);
    return cached;
  } catch {
    return null;
  }
}

export async function setCachedPage(path: string, html: string): Promise<void> {
  try {
    await redis.setex(path, CACHE_TTL_SECONDS, html);
  } catch {
    // Silently fail - cache is optional
  }
}

export async function clearCachedPage(path: string): Promise<void> {
  try {
    await redis.del(path);
  } catch {
    // Silently fail
  }
}

export default redis;