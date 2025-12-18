import { z } from "zod";

export const ogrnSchema = z
  .string()
  .regex(/^\d{13}$|^\d{15}$/, "ОГРН должен содержать 13 или 15 цифр");

export const ogrnipSchema = z
  .string()
  .regex(/^\d{15}$/, "ОГРНИП должен содержать 15 цифр");

export const innSchema = z
  .string()
  .regex(/^\d{10}$|^\d{12}$/, "ИНН должен содержать 10 или 12 цифр");

export const kppSchema = z
  .string()
  .regex(/^\d{9}$/, "КПП должен содержать 9 цифр");

export const searchQuerySchema = z
  .string()
  .min(2, "Минимум 2 символа")
  .max(200, "Максимум 200 символов");

