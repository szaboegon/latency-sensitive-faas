import type { AppResult } from "../models/models";

function randomBase64(length: number) {
  const chars =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

export const resultsMock: AppResult[] = [
  {
    timestamp: new Date().toISOString(),
    event: JSON.stringify({
      status: "ok",
      value: Math.random(),
      info: "Single value result",
    }),
  },
  {
    timestamp: new Date().toISOString(),
    event: JSON.stringify({
      status: "error",
      message: "Something went wrong",
      code: 500,
      details: {
        trace: randomBase64(200),
      },
    }),
  },
  {
    timestamp: new Date().toISOString(),
    event: JSON.stringify({
      status: "ok",
      data: {
        id: Math.floor(Math.random() * 10000),
        payload: randomBase64(500),
        meta: {
          info: "This is a very long payload",
          extra: randomBase64(1000),
        },
      },
    }),
  },
  {
    timestamp: new Date().toISOString(),
    event: JSON.stringify({
      status: "ok",
      values: Array.from({ length: 10 }, (_, i) => ({
        idx: i,
        value: Math.random(),
        blob: randomBase64(300),
      })),
    }),
  },
  {
    timestamp: new Date().toISOString(),
    event: JSON.stringify({
      status: "ok",
      summary: "Short result",
      data: "Hello world!",
    }),
  },
];
