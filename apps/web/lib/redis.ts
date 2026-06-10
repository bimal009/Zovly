import { Redis } from "ioredis";

const addr = process.env.REDIS_ADDR ?? "localhost:6379";
const [host, port] = addr.split(":");

export const redis = new Redis({
  host: host || "localhost",
  port: Number(port || 6379),
  password: process.env.REDIS_PASSWORD || undefined,
  db: Number(process.env.REDIS_DB ?? 0),
});
