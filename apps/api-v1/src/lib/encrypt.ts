
import crypto from "crypto";


const ALGORITHM = "aes-256-gcm";
const IV_LENGTH = 12; 
const SALT = "better-auth-token-encryption";

function deriveKey(secret: string): Buffer {
  if (!secret) {
    throw new Error("Missing secret for token encryption. Set BETTER_AUTH_SECRET.");
  }
  return crypto.scryptSync(secret, SALT, 32);
}

export function encryptToken(
  plainText: string,
  secret: string = process.env.BETTER_AUTH_SECRET!
): string {
  const key = deriveKey(secret);
  const iv = crypto.randomBytes(IV_LENGTH);
  const cipher = crypto.createCipheriv(ALGORITHM, key, iv);

  const encrypted = Buffer.concat([
    cipher.update(plainText, "utf8"),
    cipher.final(),
  ]);
  const authTag = cipher.getAuthTag();

  return [iv.toString("hex"), authTag.toString("hex"), encrypted.toString("hex")].join(":");
}

export function decryptToken(
  encryptedPayload: string,
  secret: string = process.env.BETTER_AUTH_SECRET!
): string {
  const [ivHex, authTagHex, dataHex] = encryptedPayload.split(":");
  if (!ivHex || !authTagHex || !dataHex) {
    throw new Error("Invalid encrypted payload format. Expected 'iv:authTag:data'.");
  }

  const key = deriveKey(secret);
  const iv = Buffer.from(ivHex, "hex");
  const authTag = Buffer.from(authTagHex, "hex");
  const encrypted = Buffer.from(dataHex, "hex");

  const decipher = crypto.createDecipheriv(ALGORITHM, key, iv);
  decipher.setAuthTag(authTag);

  const decrypted = Buffer.concat([decipher.update(encrypted), decipher.final()]);
  return decrypted.toString("utf8");
}

export function isEncryptedToken(value: string | null | undefined): boolean {
  if (!value) return false;
  const parts = value.split(":");
  return parts.length === 3 && parts.every((p) => /^[0-9a-f]+$/i.test(p));
}