import { Metadata } from "next";
import type { LegalEntity } from "@/lib/api";

export function generateCompanyMetadata(company: LegalEntity): Metadata {
  const title = `${company.fullName} - ОГРН ${company.ogrn}`;
  const description = `Информация о компании ${company.fullName}. ОГРН: ${company.ogrn}, ИНН: ${company.inn}${company.kpp ? `, КПП: ${company.kpp}` : ""}. Статус: ${company.status}.${company.address?.region ? ` Регион: ${company.address.region}.` : ""}${company.mainActivity ? ` Основная деятельность: ${company.mainActivity.name}.` : ""}`;

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      type: "website",
      siteName: "ЕГРЮЛ Система",
      locale: "ru_RU",
    },
    twitter: {
      card: "summary",
      title,
      description,
    },
    robots: {
      index: true,
      follow: true,
    },
    alternates: {
      canonical: `/company/${company.ogrn}`,
    },
  };
}

export function generateCompanyJsonLd(company: LegalEntity) {
  return {
    "@context": "https://schema.org",
    "@type": "Organization",
    "name": company.fullName,
    "alternateName": company.shortName,
    "identifier": {
      "@type": "PropertyValue",
      "name": "ОГРН",
      "value": company.ogrn
    },
    "taxID": company.inn,
    "address": company.address ? {
      "@type": "PostalAddress",
      "addressCountry": "RU",
      "addressRegion": company.address.region,
      "addressLocality": company.address.city,
      "streetAddress": `${company.address.street || ""} ${company.address.house || ""}`.trim(),
      "postalCode": company.address.postalCode
    } : undefined,
    "foundingDate": company.registrationDate,
    "organizationStatus": company.status,
    "naics": company.mainActivity?.code,
    "description": company.mainActivity?.name,
  };
}