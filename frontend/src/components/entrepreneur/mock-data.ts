import type { IndividualEntrepreneur } from "@/lib/api";

/**
 * Тестовые данные для разработки компонентов ИП
 * Используются когда реальные данные из API недоступны
 */
export const mockEntrepreneurData: IndividualEntrepreneur = {
  id: "1",
  ogrnip: "304770000100001",
  ogrnipDate: "2004-07-15",
  inn: "770000000001",
  lastName: "Иванов",
  firstName: "Иван",
  middleName: "Иванович",
  status: "Действующий",
  statusCode: "1",
  registrationDate: "2004-07-15",
  citizenshipType: "Российская Федерация",
  citizenshipCountryCode: "RU",
  citizenshipCountryName: "Российская Федерация",
  address: {
    postalCode: "125009",
    regionCode: "77",
    region: "г. Москва",
    city: "Москва",
    street: "ул. Тверская",
    house: "15",
    building: "1",
    flat: "25",
    fullAddress: "125009, г. Москва, ул. Тверская, д. 15, стр. 1, кв. 25",
    fiasId: "0c5b2444-70a0-4932-980c-b4dc0d3f02b5",
    kladrCode: "7700000000000"
  },
  mainActivity: {
    code: "47.11",
    name: "Торговля розничная преимущественно пищевыми продуктами, включая напитки, и табачными изделиями в неспециализированных магазинах",
    isMain: true
  },
  activities: [
    {
      code: "47.11",
      name: "Торговля розничная преимущественно пищевыми продуктами, включая напитки, и табачными изделиями в неспециализированных магазинах",
      isMain: true
    },
    {
      code: "47.19",
      name: "Торговля розничная прочая в неспециализированных магазинах",
      isMain: false
    },
    {
      code: "56.10",
      name: "Деятельность ресторанов и услуги по доставке продуктов питания",
      isMain: false
    },
    {
      code: "68.20",
      name: "Аренда и управление собственным или арендованным недвижимым имуществом",
      isMain: false
    }
  ],
  regAuthority: {
    code: "7700",
    name: "Межрайонная инспекция Федеральной налоговой службы № 46 по г. Москве"
  },
  taxAuthority: {
    code: "7701",
    name: "Инспекция Федеральной налоговой службы № 1 по г. Москве"
  },
  history: [
    {
      id: "1",
      grn: "2047700000001",
      date: "2004-07-15",
      reasonCode: "1001",
      reasonDescription: "Государственная регистрация физического лица в качестве индивидуального предпринимателя",
      authority: {
        code: "7700",
        name: "Межрайонная инспекция Федеральной налоговой службы № 46 по г. Москве"
      },
      certificateSeries: "77",
      certificateNumber: "001234567",
      certificateDate: "2004-07-15",
      snapshotFullName: "Иванов Иван Иванович",
      snapshotStatus: "Действующий",
      snapshotAddress: "125009, г. Москва, ул. Тверская, д. 15, кв. 25"
    },
    {
      id: "2",
      grn: "2047700000002",
      date: "2010-03-20",
      reasonCode: "2001",
      reasonDescription: "Внесение изменений в сведения об индивидуальном предпринимателе",
      authority: {
        code: "7700",
        name: "Межрайонная инспекция Федеральной налоговой службы № 46 по г. Москве"
      },
      certificateSeries: "77",
      certificateNumber: "001234568",
      certificateDate: "2010-03-20",
      snapshotFullName: "Иванов Иван Иванович",
      snapshotStatus: "Действующий",
      snapshotAddress: "125009, г. Москва, ул. Тверская, д. 15, кв. 25"
    },
    {
      id: "3",
      grn: "2047700000003",
      date: "2015-11-10",
      reasonCode: "2002",
      reasonDescription: "Изменение видов экономической деятельности",
      authority: {
        code: "7700",
        name: "Межрайонная инспекция Федеральной налоговой службы № 46 по г. Москве"
      },
      certificateSeries: "77",
      certificateNumber: "001234569",
      certificateDate: "2015-11-10",
      snapshotFullName: "Иванов Иван Иванович",
      snapshotStatus: "Действующий",
      snapshotAddress: "125009, г. Москва, ул. Тверская, д. 15, кв. 25"
    }
  ],
  historyCount: 3,
  licensesCount: 0,
  extractDate: "2024-01-15",
  lastGrn: "2047700000003",
  lastGrnDate: "2015-11-10",
  sourceFile: "egr_ip_2024_01_15.xml",
  versionDate: "2024-01-15",
  createdAt: "2024-01-15T10:00:00Z",
  updatedAt: "2024-01-15T10:00:00Z"
};

/**
 * Получить mock данные ИП по ОГРНИП
 */
export function getMockEntrepreneurByOgrnip(ogrnip: string): IndividualEntrepreneur | null {
  // В реальном приложении здесь может быть логика для разных ОГРНИП
  if (ogrnip === mockEntrepreneurData.ogrnip) {
    return mockEntrepreneurData;
  }
  
  // Возвращаем базовые mock данные для любого ОГРНИП в режиме разработки
  return {
    ...mockEntrepreneurData,
    ogrnip: ogrnip,
    id: ogrnip
  };
}

/**
 * Флаг для использования mock данных в режиме разработки
 */
export const useMockData = process.env.NODE_ENV === "development" && process.env.NEXT_PUBLIC_USE_MOCK_DATA === "true";