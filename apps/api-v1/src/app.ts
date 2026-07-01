import express, {
  type Request,
  type Response,
  type NextFunction,
} from "express";
import "dotenv/config";
import helmet from "helmet";
import cors, { type CorsOptions } from "cors";
import rateLimit from "express-rate-limit";
import compression from "compression";
import morgan from "morgan";
import { toNodeHandler } from "better-auth/node";
import { auth } from "./lib/auth";
import imagekitRouter from "./modules/imagekit/imagekit.routes";
import { plansRoutes } from "./modules/plans/plans.routes";
import businessRouter from "./modules/business/business.routes";
import faqRouter from "./modules/faq/faq.route";
import facebookRouter from "./modules/facebook/facebook.routes";
const isProd = process.env.NODE_ENV === "production";

const allowedOrigins =
  process.env.ALLOWED_ORIGINS?.split(",")
    .map((o) => o.trim())
    .filter(Boolean) ?? [];

if (isProd && allowedOrigins.length === 0) {
  throw new Error(
    "ALLOWED_ORIGINS must be set in production (comma-separated list of origins)",
  );
}

const app = express();

const trustProxy = process.env.TRUST_PROXY;
if (trustProxy) {
  app.set(
    "trust proxy",
    Number.isNaN(Number(trustProxy)) ? trustProxy : Number(trustProxy),
  );
}

app.disable("x-powered-by");

app.use(helmet());

const corsOptions: CorsOptions = {
  origin(origin, callback) {
    if (!origin) return callback(null, true);

    if (allowedOrigins.includes(origin)) return callback(null, true);

    return callback(null, false);
  },
  credentials: true,
  methods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
  allowedHeaders: ["Content-Type", "Authorization"],
  maxAge: 600,
};

app.use(cors(corsOptions));

const limiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 100,
  standardHeaders: true,
  legacyHeaders: false,
  message: {
    success: false,
    message: "Too many requests, please try again later.",
  },
});
app.use(limiter);

export const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 20,
  standardHeaders: true,
  legacyHeaders: false,
  skipSuccessfulRequests: true,
  message: {
    success: false,
    message: "Too many attempts, please try again later.",
  },
});
app.all("/api/v1/auth/{*path}", toNodeHandler(auth));
app.use(express.json({ limit: "10kb" }));
app.use(express.urlencoded({ extended: true, limit: "10kb" }));

app.use(compression());
app.use(morgan(isProd ? "combined" : "dev"));

app.get("/api/v1/health", (_req: Request, res: Response) => {
  res.json({
    success: true,
    message: "ok",
    uptime: process.uptime(),
    status: "healthy",
    timestamp: new Date().toISOString(),
  });
});

app.use("/api/v1/images", imagekitRouter);
app.use("/api/v1/plans", plansRoutes);
app.use("/api/v1/business", businessRouter);
app.use("/api/v1/faq", faqRouter);
app.use("/api/v1/apps", facebookRouter);

app.use((_req: Request, res: Response) => {
  res.status(404).json({ success: false, message: "Route not found" });
});

interface HttpError extends Error {
  status?: number;
  statusCode?: number;
}

app.use((err: HttpError, _req: Request, res: Response, _next: NextFunction) => {
  const status = err.statusCode ?? err.status ?? 500;

  if (status >= 500) {
    console.error(err);
  }

  res.status(status).json({
    success: false,
    message: status >= 500 && isProd ? "Internal server error" : err.message,
  });
});

export { app };
