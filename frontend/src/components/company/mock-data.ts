import type { LegalEntity } from "@/lib/api";

/**
 * Тестовые данные для разработки компонентов компании
 * Используются когда реальные данные из API недоступны
 */
export const mockCompanyData: LegalEntity = {
  id: "1",
  ogrn: "1027700132195",
  inn: "7707083893",
  kpp: "770701001",
  fullName: "ПУБЛИЧНОЕ АКЦИОНЕРНОЕ ОБЩЕСТВО \"СБЕРБАНК РОССИИ\"",
  shortName: "ПАО Сбербанк",
  status: "Действующая",
  registrationDate: "2012-07-17",
  registrationAuthority: "Межрайонная инспекция Федеральной налоговой службы № 46 по г. Москве",
  address: {
    postalCode: "117997",
    regionCode: "77",
    region: "г. Москва",
    city: "Москва",
    street: "ул. Вавилова",
    house: "19",
    fullAddress: "117997, г. Москва, ул. Вавилова, д. 19"
  },
  mainActivity: {
    code: "64.19",
    name: "Денежное посредничество прочее"
  },
  activities: [
    {
      code: "64.19",
      name: "Денежное посредничество прочее"
    },
    {
      code: "64.92",
      name: "Прочие виды кредитования"
    },
    {
      code: "66.19",
      name: "Прочая деятельность по предоставлению услуг в сфере финансов, кроме страхования и пенсионного обеспечения"
    },
    {
      code: "66.12",
      name: "Деятельность по организации торговли ценными бумагами"
    },
    {
      code: "64.91",
      name: "Деятельность по финансовому лизингу"
    }
  ],
  capital: {
    amount: 67760844000,
    currency: "RUB"
  },
  head: {
    lastName: "Греф",
    firstName: "Герман",
    middleName: "Оскарович",
    inn: "770400372208",
    position: "Президент, Председатель Правления"
  },
  founders: [
    {
      type: "RUSSIAN_COMPANY",
      name: "МИНИСТЕРСТВО ФИНАНСОВ РОССИЙСКОЙ ФЕДЕРАЦИИ",
      inn: "7710168360",
      sharePercent: 50.0000000001,
      shareNominalValue: 33880422000.01
    },
    {
      type: "RUSSIAN_COMPANY",
      name: "ЦЕНТРАЛЬНЫЙ БАНК РОССИЙСКОЙ ФЕДЕРАЦИИ",
      inn: "7702235133",
      sharePercent: 49.9999999999,
      shareNominalValue: 33880421999.99
    }
  ],
  history: [
    {
      id: "1",
      grn: "1027700132195",
      date: "2012-07-17",
      reasonCode: "11201",
      reasonDescription: "Государственная регистрация юридического лица",
      authority: {
        code: "7746",
        name: "Межрайонная инспекция Федеральной налоговой службы № 46 по г. Москве"
      },
      certificateSeries: "77",
      certificateNumber: "001234567",
      certificateDate: "2012-07-17"
    },
    {
      id: "2",
      grn: "2207700345678",
      date: "2020-03-25",
      reasonCode: "12101",
      reasonDescription: "Изменение сведений о юридическом лице, содержащихся в ЕГРЮЛ",
      authority: {
        code: "7746",
        name: "Межрайонная инспекция Федеральной налоговой службы № 46 по г. Москве"
      },
      snapshotAddress: "117997, г. Москва, ул. Вавилова, д. 19, стр. 1"
    },
    {
      id: "3",
      grn: "2217700456789",
      date: "2021-06-15",
      reasonCode: "12102",
      reasonDescription: "Изменение размера уставного капитала",
      authority: {
        code: "7746",
        name: "Межрайонная инспекция Федеральной налоговой службы № 46 по г. Москве"
      },
      snapshotFullName: "ПУБЛИЧНОЕ АКЦИОНЕРНОЕ ОБЩЕСТВО \"СБЕРБАНК РОССИИ\""
    }
  ],
  relatedCompanies: [
    {
      id: "1",
      ogrn: "1027739609391",
      name: "АКЦИОНЕРНОЕ ОБЩЕСТВО \"СБЕРБАНК ЛИЗИНГ\"",
      relationshipType: "Дочерняя",
      status: "Действующая"
    },
    {
      id: "2", 
      ogrn: "1027700167110",
      name: "АКЦИОНЕРНОЕ ОБЩЕСТВО \"СБЕРБАНК СТРАХОВАНИЕ\"",
      relationshipType: "Дочерняя",
      status: "Действующая"
    },
    {
      id: "3",
      ogrn: "1027700090932",
      name: "ОБЩЕСТВО С ОГРАНИЧЕННОЙ ОТВЕТСТВЕННОСТЬЮ \"СБЕРБАНК ТЕХНОЛОГИИ\"",
      relationshipType: "Дочерняя", 
      status: "Действующая"
    }
  ]
};

/**
 * Функция для получения тестовых данных по ОГРН
 * Используется в режиме разработки когда API недоступно
 */
export function getMockCompanyByOgrn(ogrn: string): LegalEntity | null {
  if (ogrn === mockCompanyData.ogrn) {
    return mockCompanyData;
  }
  
  // Возвращаем модифицированную версию для других ОГРН
  return {
    ...mockCompanyData,
    ogrn,
    fullName: `ТЕСТОВАЯ КОМПАНИЯ ${ogrn}`,
    shortName: `ТК ${ogrn.slice(-4)}`,
    inn: `77${ogrn.slice(-8)}`,
    kpp: `${ogrn.slice(-9, -6)}01001`
  };
}

/**
 * Проверка, является ли окружение режимом разработки
 */
export const isDevelopment = process.env.NODE_ENV === "development";

/**
 * Флаг для использования mock данных
 * Можно переключать для тестирования компонентов
 */
export const useMockData = isDevelopment && process.env.NEXT_PUBLIC_USE_MOCK_DATA === "true";