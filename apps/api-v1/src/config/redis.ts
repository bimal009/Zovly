import Redis from "ioredis";
import "dotenv/config";

const url = process.env.REDIS_URL;
if (!url)
  throw new Error("REDIS_URL is not defined in the environment variables.");

export const client = new Redis(url);

export const WORKER_KEY = "worker:";
export const WORKER_LIST_KEY = "worker:list:";
export const CATEGORY_KEY = "category:";

export const TTL = {
  SHORT: 60 * 5, // 5 min  — lists, search results
  MEDIUM: 60 * 30, // 30 min — individual records
  LONG: 60 * 60 * 24, // 24 hr  — categories (rarely change)
} as const;

export const cache = {
  async get<T>(key: string): Promise<T | null> {
    const raw = await client.get(key);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as T;
    } catch {
      return null;
    }
  },

  async set(key: string, value: unknown, ttl: number): Promise<void> {
    await client.set(key, JSON.stringify(value), "EX", ttl);
  },

  async del(...keys: string[]): Promise<void> {
    if (keys.length) await client.del(...keys);
  },

  async delPattern(pattern: string): Promise<void> {
    let cursor = "0";
    do {
      const [next, keys] = await client.scan(
        cursor,
        "MATCH",
        pattern,
        "COUNT",
        100,
      );
      cursor = next;
      if (keys.length) await client.del(...keys);
    } while (cursor !== "0");
  },
};

export const keys = {
  worker: (id: string) => `${WORKER_KEY}${id}`,
  workerList: (suffix: string) => `${WORKER_LIST_KEY}${suffix}`,
  category: (id: string) => `${CATEGORY_KEY}${id}`,
  categoryAll: () => `${CATEGORY_KEY}all`,
  categorySearch: (q: string) => `${CATEGORY_KEY}search:${q}`,
};
