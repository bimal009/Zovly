import { InternalServerError, ServiceUnavailableError } from "../../lib/errors";

export type EmbedResponse = {
  chunk_index: number;
  content: string;
  embedding: number[];
};

export const embedFaqs = async (
  question: string,
  answer: string,
): Promise<EmbedResponse[]> => {
  const res = await fetch(`${process.env.AI_SERVICE_URL}/api/v1/ml/embed/faq`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ question, answer }),
  });

  if (res.status === 502) {
    throw new ServiceUnavailableError("AI service is not available");
  }

  const data = await res.json();

  if (!res.ok) {
    throw new InternalServerError(data.message ?? "Embedding failed");
  }

  return data;
};
