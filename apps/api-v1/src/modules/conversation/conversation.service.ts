import { CreateConversationInput } from '@repo/types';
import { eq, and, desc } from "drizzle-orm"
import { conversations } from '../../config/db/schema/conversations';
import { db } from '../../config/db/db';
import { InternalServerError, NotFoundError } from '../../lib/errors';

export const create = async (convo: CreateConversationInput, businessId: string) => {
    const [newConvo] = await db.insert(conversations).values({
        businessId,
        platform: convo.platform,
        contactId: convo.contactId,
        contactName: convo.contactName,
        contactUsername: convo.contactUsername,
        contactAvatarUrl: convo.contactAvatarUrl,
        threadId: convo.threadId,
    }).returning()

    if (!newConvo) throw new InternalServerError("Failed to create conversation")
    return newConvo
}

export const getById = async (conversationId: string, businessId: string) => {
    const [convo] = await db
        .select()
        .from(conversations)
        .where(and(eq(conversations.id, conversationId), eq(conversations.businessId, businessId)))

    if (!convo) throw new NotFoundError("Conversation not found")
    return convo
}

export const getByThreadId = async (threadId: string, businessId: string) => {
    const [convo] = await db
        .select()
        .from(conversations)
        .where(and(eq(conversations.threadId, threadId), eq(conversations.businessId, businessId)))

    return convo ?? null
}

export const list = async (businessId: string, limit = 50, offset = 0) => {
    return db
        .select()
        .from(conversations)
        .where(eq(conversations.businessId, businessId))
        .orderBy(desc(conversations.createdAt))
        .limit(limit)
        .offset(offset)
}

export const update = async (
    conversationId: string,
    businessId: string,
    data: Partial<CreateConversationInput>
) => {
    const [updated] = await db
        .update(conversations)
        .set(data)
        .where(and(eq(conversations.id, conversationId), eq(conversations.businessId, businessId)))
        .returning()

    if (!updated) throw new NotFoundError("Conversation not found")
    return updated
}

export const remove = async (conversationId: string, businessId: string) => {
    const [deleted] = await db
        .delete(conversations)
        .where(and(eq(conversations.id, conversationId), eq(conversations.businessId, businessId)))
        .returning()

    if (!deleted) throw new NotFoundError("Conversation not found")
    return deleted
}

export const findOrCreate = async (convo: CreateConversationInput, businessId: string) => {
    const existing = await getByThreadId(convo.threadId, businessId)
    if (existing) return existing

    return create(convo, businessId)
}