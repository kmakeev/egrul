import { z } from "zod";
import { searchQuerySchema } from "./common";

export const searchFiltersSchema = z.object({
  q: searchQuerySchema.optional().catch(undefined),
  innOgrn: z
    .string()
    .trim()
    .max(20, "Слишком длинное значение")
    .optional()
    .catch(undefined),
  region: z.string().trim().max(100).optional().catch(undefined),
  okved: z.string().trim().max(20).optional().catch(undefined),
  // Статус теперь фильтруется по коду status_code (например, "101", "201").
  // Пустое значение означает "все статусы".
  status: z
    .string()
    .trim()
    .max(10, "Слишком длинный код статуса")
    .optional()
    .catch(undefined),
  // ФИО учредителя для поиска по учредителям ЮЛ
  founderName: z
    .string()
    .trim()
    .max(200, "Слишком длинное значение")
    .optional()
    .catch(undefined),
  dateFrom: z.string().optional().catch(undefined),
  dateTo: z.string().optional().catch(undefined),
  page: z.coerce.number().min(1).default(1),
  pageSize: z.coerce.number().min(1).max(100).default(20),
  sortBy: z
    .enum(["name", "inn", "region", "status", "registrationDate"])
    .optional()
    .catch(undefined),
  sortOrder: z.enum(["asc", "desc"]).default("asc"),
});

export type SearchFiltersInput = z.infer<typeof searchFiltersSchema>;


