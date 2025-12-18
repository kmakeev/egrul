import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().email("Некорректный email"),
  password: z.string().min(6, "Минимум 6 символов"),
});

export const registerSchema = z.object({
  email: z.string().email("Некорректный email"),
  password: z.string().min(8, "Минимум 8 символов"),
  confirmPassword: z.string(),
  firstName: z.string().min(2, "Минимум 2 символа"),
  lastName: z.string().min(2, "Минимум 2 символа"),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Пароли не совпадают",
  path: ["confirmPassword"],
});

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;

