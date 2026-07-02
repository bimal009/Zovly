import { Queue} from "bullmq"
import IORedis from "ioredis"
import "dotenv/config"
export const bullConnection = new IORedis(process.env.REDIS_URL!, {
    maxRetriesPerRequest: null,
})

export const messageQueue = new Queue("messages", {
    connection: bullConnection,
    defaultJobOptions: {
        removeOnComplete: true,
        removeOnFail: 100, 
        attempts: 3,
        backoff: { type: "exponential", delay: 1000 },
    },
})

