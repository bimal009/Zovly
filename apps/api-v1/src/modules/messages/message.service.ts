import { eq, and, desc } from "drizzle-orm"
import { CreateConversationInput, CreateMessageInput } from "@repo/types"
import { db } from "../../config/db/db"
import { messages } from "../../config/db/schema/messages"
import { InternalServerError, NotFoundError } from "../../lib/errors"
import { conversations } from "../../config/db/schema/conversations"

export const create = async (
    message: CreateMessageInput,
    conversation: CreateConversationInput,
    businessId: string
) => {
    const [existing] = await db
        .select()
        .from(conversations)
        .where(and(
            eq(conversations.threadId, conversation.threadId),
            eq(conversations.businessId, businessId)
        ))

    return db.transaction(async (tx) => {
        let convo = existing

        if (!convo) {
            const [inserted] = await tx.insert(conversations).values({
                businessId,
                platform: conversation.platform,
                contactId: conversation.contactId,
                contactName: conversation.contactName,
                contactUsername: conversation.contactUsername,
                contactAvatarUrl: conversation.contactAvatarUrl,
                threadId: conversation.threadId,
            })
            .onConflictDoNothing({ target: [conversations.businessId, conversations.threadId] })
            .returning()

            convo = inserted ?? (await tx
                .select()
                .from(conversations)
                .where(and(
                    eq(conversations.threadId, conversation.threadId),
                    eq(conversations.businessId, businessId)
                ))
            )[0]

            if (!convo) throw new InternalServerError("Failed to find or create conversation")
        }

        const [newMsg] = await tx.insert(messages).values({
            businessId,
            conversationId: convo.id,
            content: message.content,
            direction: message.direction,
            mediaUrl: message.mediaUrl,
            mediaType: message.mediaType,
            platformMessageId: message.platformMessageId,
            replyToMessageId: message.replyToMessageId,
            platformSenderId: message.platformSenderId,
        }).returning()

        if (!newMsg) throw new InternalServerError("Failed to create message")

        return { message: newMsg, conversation: convo }
    })
}

export const getById = async (messageId: string, businessId: string) => {
    const [msg] = await db
        .select()
        .from(messages)
        .where(and(eq(messages.id, messageId), eq(messages.businessId, businessId)))

    if (!msg) throw new NotFoundError("Message not found")
    return msg
}

export const listByConversation = async (
    conversationId: string,
    businessId: string,
    limit = 50,
    offset = 0
) => {
    return db
        .select()
        .from(messages)
        .where(and(eq(messages.conversationId, conversationId), eq(messages.businessId, businessId)))
        .orderBy(desc(messages.sentAt))
        .limit(limit)
        .offset(offset)
}

export const update = async (
    messageId: string,
    businessId: string,
    data: Partial<CreateMessageInput>
) => {
    const [updated] = await db
        .update(messages)
        .set(data)
        .where(and(eq(messages.id, messageId), eq(messages.businessId, businessId)))
        .returning()

    if (!updated) throw new NotFoundError("Message not found")
    return updated
}

export const remove = async (messageId: string, businessId: string) => {
    const [deleted] = await db
        .delete(messages)
        .where(and(eq(messages.id, messageId), eq(messages.businessId, businessId)))
        .returning()

    if (!deleted) throw new NotFoundError("Message not found")
    return deleted
}


