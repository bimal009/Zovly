import { faqs } from './../../config/db/schema/faqs';
import { CreateFaqInput, PaginationQuery, UpdateFaqInput } from "@repo/types";
import { db } from "../../config/db/db";
import { and, count, eq, ilike, or, sql } from "drizzle-orm";
import { embedFaqs, EmbedResponse } from "../embeddings/embedding.service";
import { knowledgeChunks } from "../../config/db/schema/businessKnowledge";
import { InternalServerError, NotFoundError } from '../../lib/errors';

export const create = async (input: CreateFaqInput, businessId: string) => {
  let embedResults;
  try {
    embedResults = await embedFaqs(input.question, input.answer);
  } catch (err) {
    throw new InternalServerError(
      "Failed to generate embeddings for FAQ. Error: " + err,
    );
  }

  try {
    const faq = await db.transaction(async (tx) => {
      const [newFaq] = await tx
        .insert(faqs)
        .values({
          businessId,
          question: input.question,
          answer: input.answer,
          isActive: input.isActive,
          sortOrder: sql`(SELECT COALESCE(MAX(sort_order), 0) + 1 FROM faqs WHERE business_id = ${businessId})`,
        })
        .returning();

      await tx.insert(knowledgeChunks).values(
        embedResults.map((chunk) => ({
          businessId,
          sourceType: "faq" as const,
          sourceId: newFaq.id,
          chunkIndex: chunk.chunk_index,
          content: chunk.content,
          embedding: chunk.embedding,
          metadata: {
            question: newFaq.question,
          },
        })),
      );

      return newFaq;
    });

    return faq;
  } catch (err) {
    throw new InternalServerError(
      "Failed to create FAQ. Error: " + err,
    );
  }
};


export const get = async (query: PaginationQuery, businessId: string) => {
  const search = query.search ? `%${query.search.trim()}%` : undefined;
  const offset = (query.page - 1) * query.limit;

  const whereClause = and(
    eq(faqs.businessId, businessId),
    search
      ? or(ilike(faqs.question, search), ilike(faqs.answer, search))
      : undefined,
  );

  try {
    const [data, [{ total }]] = await Promise.all([
      db
        .select()
        .from(faqs)
        .where(whereClause)
        .limit(query.limit)
        .offset(offset),
      db
        .select({ total: count(faqs.id) })
        .from(faqs)
        .where(whereClause),
    ]);

    return {
      data,
      total: Number(total),
      page: query.page,
      limit: query.limit,
      totalPages: Math.ceil(Number(total) / query.limit),
    };
  } catch (err) {
    throw new InternalServerError(
      "Failed to fetch FAQs. Error: " + err,
    );
  }
};


export const update = async (
  id: string,
  input: UpdateFaqInput,
  businessId: string,
) => {
  const [existing] = await db
    .select()
    .from(faqs)
    .where(and(eq(faqs.id, id), eq(faqs.businessId, businessId)));

  if (!existing) {
    throw new NotFoundError("FAQ not found");
  }

  const questionChanged =
    input.question !== undefined && input.question !== existing.question;
  const answerChanged =
    input.answer !== undefined && input.answer !== existing.answer;
  const needsReembedding = questionChanged || answerChanged;

  let embedResults:EmbedResponse[] | undefined;
  if (needsReembedding) {
    try {
      embedResults = await embedFaqs(
        input.question ?? existing.question,
        input.answer ?? existing.answer,
      );
    } catch (err) {
      throw new InternalServerError(
        "Failed to generate embeddings for FAQ. Error: " + err,
      );
    }
  }

  try {
    const faq = await db.transaction(async (tx) => {
      const [updatedFaq] = await tx
        .update(faqs)
        .set({
          ...(input.question !== undefined && { question: input.question }),
          ...(input.answer !== undefined && { answer: input.answer }),
          ...(input.isActive !== undefined && { isActive: input.isActive }),
          updatedAt: new Date(),
        })
        .where(and(eq(faqs.id, id), eq(faqs.businessId, businessId)))
        .returning();

      if (needsReembedding && embedResults) {
        await tx
          .delete(knowledgeChunks)
          .where(
            and(
              eq(knowledgeChunks.sourceType, "faq"),
              eq(knowledgeChunks.sourceId, id),
            ),
          );

        await tx.insert(knowledgeChunks).values(
          embedResults.map((chunk) => ({
            businessId,
            sourceType: "faq" as const,
            sourceId: updatedFaq.id,
            chunkIndex: chunk.chunk_index,
            content: chunk.content,
            embedding: chunk.embedding,
            metadata: {
              question: updatedFaq.question,
            },
            isActive: updatedFaq.isActive,
          })),
        );
      } else if (input.isActive !== undefined) {
        await tx
          .update(knowledgeChunks)
          .set({ isActive: updatedFaq.isActive })
          .where(
            and(
              eq(knowledgeChunks.sourceType, "faq"),
              eq(knowledgeChunks.sourceId, id),
            ),
          );
      }

      return updatedFaq;
    });

    return faq;
  } catch (err) {
    throw new InternalServerError("Failed to update FAQ. Error: " + err);
  }
};
export const remove = async (id: string, businessId: string) => {
  try {
    const faq = await db.transaction(async (tx) => {
      const [deletedFaq] = await tx
        .delete(faqs)
        .where(and(eq(faqs.id, id), eq(faqs.businessId, businessId)))
        .returning();

      if (!deletedFaq) {
        throw new NotFoundError("FAQ not found");
      }

      await tx
        .delete(knowledgeChunks)
        .where(
          and(
            eq(knowledgeChunks.sourceType, "faq"),
            eq(knowledgeChunks.sourceId, id),
          ),
        );

      return deletedFaq;
    });

    return faq;
  } catch (err) {
    if (err instanceof NotFoundError) throw err;
    throw new InternalServerError("Failed to delete FAQ. Error: " + err);
  }
};