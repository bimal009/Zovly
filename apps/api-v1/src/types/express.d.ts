import { Business } from "../config/db/schema/business";
import { BusinessMember } from "../config/db/schema/members";
import { Session, User } from "../lib/auth";

declare global {
  namespace Express {
    interface Request {
      user?: User;
      session?: Session;
      business?: Business;
      member?: BusinessMember;
    }
  }
}
