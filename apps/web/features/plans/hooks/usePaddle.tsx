"use client";

import {
  type Environments,
  type Paddle,
  initializePaddle,
} from "@paddle/paddle-js";
import { useEffect, useState } from "react";

export function usePaddle() {
  const [paddle, setPaddle] = useState<Paddle>();

  useEffect(() => {
    initializePaddle({
      environment: process.env.NEXT_PUBLIC_PADDLE_ENV as Environments,
      token: process.env.NEXT_PUBLIC_PADDLE_CLIENT_TOKEN!,
      checkout: {
        settings: {
          variant: "one-page",
        },
      },
    }).then((paddleInstance: Paddle | undefined) => {
      if (paddleInstance) {
        setPaddle(paddleInstance);
      }
    });
  }, []);

  return paddle;
}
