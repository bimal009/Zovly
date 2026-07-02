import { CreateMessageInput } from "@repo/types";
import { client } from "../../../config/redis";

const DEFAULT_MAX_MESSAGES = 10;
const TTL_SECONDS = 3600;

export const addMessageToRedis = async (
  conversationId: string,
  message: CreateMessageInput,
  n: number = DEFAULT_MAX_MESSAGES
) => {
  const redisKey = `messages:${conversationId}`;

  const existing = await client.get(redisKey);
  let messagesArray: CreateMessageInput[] = existing ? JSON.parse(existing) : [];

  messagesArray.push(message);

  if (messagesArray.length > n) {
    messagesArray = messagesArray.slice(-n);
  }

  await client.set(redisKey, JSON.stringify(messagesArray), "EX", TTL_SECONDS);

  return messagesArray;
};

export const getRecentMessagesFromRedis = async (
  conversationId: string
): Promise<CreateMessageInput[]> => {
  const redisKey = `messages:${conversationId}`;
  const existing = await client.get(redisKey);
  return existing ? JSON.parse(existing) : [];
};