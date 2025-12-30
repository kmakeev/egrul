"use client";

import { useEffect, useRef, useState, useMemo, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Network as NetworkIcon, ZoomIn, ZoomOut, Maximize2, RotateCcw } from "lucide-react";
import { decodeHtmlEntities } from "@/lib/html-utils";
import { GraphSettings } from "./graph-settings";
import type { LegalEntity } from "@/lib/api";
import type { Network } from "vis-network";
import type { DataSet } from "vis-data";

// Импорты vis.js с динамической загрузкой
let VisNetwork: typeof Network | null = null;
let VisDataSet: typeof DataSet | null = null;

// Функция для динамической загрузки vis.js библиотек
const loadVisLibraries = async () => {
  if (typeof window !== 'undefined' && !VisNetwork) {
    try {
      const [visNetworkModule, visDataModule] = await Promise.all([
        import('vis-network'),
        import('vis-data')
      ]);
      VisNetwork = visNetworkModule.Network;
      VisDataSet = visDataModule.DataSet;
      console.log('vis.js libraries loaded successfully');
      return true;
    } catch (error) {
      console.error('Failed to load vis.js libraries:', error);
      return false;
    }
  }
  return VisNetwork !== null && VisDataSet !== null;
};

interface RelatedCompany {
  company: LegalEntity;
  relationshipType: string;
  description?: string;
  commonFounders?: Array<{
    sharePercent?: number;
    name?: string;
    lastName?: string;
    firstName?: string;
    middleName?: string;
  }>;
}

interface CompanyRelationsGraphProps {
  company: LegalEntity;
  relatedCompanies: RelatedCompany[];
  isLoading?: boolean;
}

// Функция для создания информативной подсказки для центральной компании
function createMainCompanyTooltip(company: LegalEntity): string {
  const name = decodeHtmlEntities(company.fullName || company.shortName || "");
  const shortName = name.length > 50 ? name.substring(0, 47) + "..." : name;
  // Более информативный тултип в одну строку
  const tooltip = `🏢 ${shortName} | ОГРН: ${company.ogrn} | ИНН: ${company.inn} | Центральная компания`;
  
  if (process.env.NODE_ENV === 'development') {
    console.log('Main company tooltip created:', tooltip);
  }
  
  return tooltip;
}

// Функция для создания информативной подсказки для связанной компании
function createRelatedCompanyTooltip(relation: RelatedCompany): string {
  const company = relation.company;
  const name = decodeHtmlEntities(company.fullName || company.shortName || "");
  const shortName = name.length > 40 ? name.substring(0, 37) + "..." : name;
  const status = getCompanyStatus(company);
  const relationshipLabel = getRelationshipLabel(relation.relationshipType, relation);
  
  // Более информативный тултип в одну строку с дополнительной информацией
  let tooltip = `🏢 ${shortName} | ОГРН: ${company.ogrn} | ${relationshipLabel} | ${status.text}`;
  
  // Добавляем дополнительную информацию в зависимости от типа связи
  if (relation.relationshipType === 'SUBSIDIARY_COMPANY' && relation.commonFounders?.[0]?.sharePercent) {
    tooltip += ` | Доля: ${relation.commonFounders[0].sharePercent}%`;
  }
  
  if (company.address?.city) {
    tooltip += ` | ${company.address.city}`;
  }
  
  if (process.env.NODE_ENV === 'development') {
    console.log('Related company tooltip created for', company.ogrn, ':', tooltip);
  }
  
  return tooltip;
}

// Функция для создания информативной подсказки для ребра (связи)
function createEdgeTooltip(relation: RelatedCompany, relationshipLabel: string): string {
  const company = relation.company;
  const name = decodeHtmlEntities(company.fullName || company.shortName || "");
  const shortName = name.length > 30 ? name.substring(0, 27) + "..." : name;
  
  let tooltip = `🔗 ${relationshipLabel} | ${shortName}`;
  
  // Добавляем специфичную информацию в зависимости от типа связи
  switch (relation.relationshipType) {
    case 'SUBSIDIARY_COMPANY':
      if (relation.commonFounders?.[0]?.sharePercent) {
        tooltip += ` | Доля владения: ${relation.commonFounders[0].sharePercent}%`;
      }
      break;
    case 'COMMON_FOUNDERS':
      if (relation.commonFounders && relation.commonFounders.length > 0) {
        const founderName = relation.commonFounders[0].lastName || relation.commonFounders[0].name || 'Физлицо';
        tooltip += ` | Общий учредитель: ${founderName}`;
      }
      break;
    case 'COMMON_DIRECTORS':
      tooltip += ` | Общий руководитель`;
      break;
    case 'FOUNDER_COMPANY':
      tooltip += ` | Учредитель основной компании`;
      break;
    case 'FOUNDER_TO_DIRECTOR':
      tooltip += ` | Учредитель → Руководитель`;
      break;
    case 'DIRECTOR_TO_FOUNDER':
      tooltip += ` | Руководитель → Учредитель`;
      break;
  }
  
  if (process.env.NODE_ENV === 'development') {
    console.log('Edge tooltip created for', relation.relationshipType, ':', tooltip);
  }
  
  return tooltip;
}

const RELATIONSHIP_COLORS = {
  FOUNDER_COMPANY: "#3b82f6", // blue
  SUBSIDIARY_COMPANY: "#10b981", // emerald
  COMMON_FOUNDERS: "#f59e0b", // amber
  COMMON_DIRECTORS: "#8b5cf6", // violet
  COMMON_ADDRESS: "#06b6d4", // cyan
  FOUNDER_TO_DIRECTOR: "#ef4444", // red
  DIRECTOR_TO_FOUNDER: "#ef4444", // red
  RELATED_BY_PERSON: "#6b7280", // gray
} as const;

// Функция для получения короткого названия компании
function getShortCompanyName(fullName: string, maxLength: number = 30): string {
  const decoded = decodeHtmlEntities(fullName);
  if (decoded.length <= maxLength) return decoded;
  
  // Пытаемся найти сокращение в кавычках
  const quotedMatch = decoded.match(/"([^"]+)"/);
  if (quotedMatch && quotedMatch[1].length <= maxLength) {
    return quotedMatch[1];
  }
  
  // Обрезаем и добавляем многоточие
  return decoded.substring(0, maxLength - 3) + "...";
}

// Функция для получения статуса компании
function getCompanyStatus(company: LegalEntity): { text: string; color: string } {
  if (company.terminationDate) {
    return { text: "Закрыта", color: "#ef4444" };
  }
  
  if (company.statusCode) {
    switch (company.statusCode) {
      case "101":
        return { text: "Ликвидируется", color: "#f59e0b" };
      case "105":
      case "106":
      case "107":
        return { text: "Исключается", color: "#ef4444" };
      case "113":
      case "114":
      case "115":
      case "116":
      case "117":
        return { text: "Банкротство", color: "#dc2626" };
      default:
        if (company.statusCode.startsWith("12") || company.statusCode.startsWith("13")) {
          return { text: "Реорганизация", color: "#8b5cf6" };
        }
    }
  }
  
  return { text: "Действующая", color: "#10b981" };
}

// Функция для получения описания связи
function getRelationshipLabel(type: string, relation?: RelatedCompany): string {
  switch (type) {
    case "FOUNDER_COMPANY":
      return "Учредитель";
    case "SUBSIDIARY_COMPANY":
      // Показываем долю владения если есть
      if (relation?.commonFounders?.[0]?.sharePercent) {
        return `Дочерняя (${relation.commonFounders[0].sharePercent}%)`;
      }
      return "Дочерняя";
    case "COMMON_FOUNDERS":
      return "Общие учредители";
    case "COMMON_DIRECTORS":
      return "Общие руководители";
    case "COMMON_ADDRESS":
      return "Общий адрес";
    case "FOUNDER_TO_DIRECTOR":
    case "DIRECTOR_TO_FOUNDER":
      return "Перекрестные связи";
    case "RELATED_BY_PERSON":
      return "Связанная";
    default:
      return "Связанная";
  }
}

export function CompanyRelationsGraph({ 
  company, 
  relatedCompanies, 
  isLoading = false 
}: CompanyRelationsGraphProps) {
  const networkRef = useRef<HTMLDivElement>(null);
  const networkInstance = useRef<Network | null>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [visLibsLoaded, setVisLibsLoaded] = useState(false);
  
  // Загружаем vis.js библиотеки при монтировании компонента
  useEffect(() => {
    loadVisLibraries().then(setVisLibsLoaded);
  }, []);
  
  // Мемоизируем доступные типы связей
  const availableTypes = useMemo(() => 
    Array.from(new Set(relatedCompanies.map(r => r.relationshipType))), 
    [relatedCompanies]
  );
  
  const [selectedTypes, setSelectedTypes] = useState<string[]>([]);
  
  // Мемоизируем функции для стабильности
  const getCompanyStatusMemo = useCallback((comp: LegalEntity) => getCompanyStatus(comp), []);
  const getRelationshipLabelMemo = useCallback((type: string, relation?: RelatedCompany) => getRelationshipLabel(type, relation), []);
  const getShortCompanyNameMemo = useCallback((name: string) => getShortCompanyName(name), []);
  const createMainCompanyTooltipMemo = useCallback((comp: LegalEntity) => createMainCompanyTooltip(comp), []);
  const createRelatedCompanyTooltipMemo = useCallback((relation: RelatedCompany) => createRelatedCompanyTooltip(relation), []);
  
  // Инициализируем выбранные типы только один раз
  useEffect(() => {
    if (availableTypes.length > 0 && selectedTypes.length === 0) {
      setSelectedTypes(availableTypes);
    }
  }, [availableTypes, selectedTypes.length]);
  
  // Мемоизируем отфильтрованные компании
  const filteredCompanies = useMemo(() => {
    if (selectedTypes.length === 0) return [];
    
    return relatedCompanies
      .filter(r => selectedTypes.includes(r.relationshipType))
      .filter((relation, index, self) => 
        self.findIndex(r => r.company.ogrn === relation.company.ogrn) === index
      );
  }, [relatedCompanies, selectedTypes]);

  useEffect(() => {
    // Проверяем все необходимые условия
    if (!networkRef.current || 
        isLoading || 
        !visLibsLoaded ||
        !VisNetwork || 
        !VisDataSet ||
        selectedTypes.length === 0) {
      console.log('Graph creation skipped:', {
        hasNetworkRef: !!networkRef.current,
        isLoading,
        visLibsLoaded,
        hasVisNetwork: !!VisNetwork,
        hasVisDataSet: !!VisDataSet,
        selectedTypesLength: selectedTypes.length
      });
      return;
    }

    // Безопасно очищаем предыдущую сеть если она существует
    if (networkInstance.current) {
      try {
        networkInstance.current.destroy();
      } catch (error) {
        // Игнорируем ошибки при уничтожении (например, если DOM уже изменен)
        if (process.env.NODE_ENV === 'development') {
          console.warn('Warning during network cleanup:', error);
        }
      }
      networkInstance.current = null;
    }

    // Если нет отфильтрованных компаний, не создаем граф
    if (filteredCompanies.length === 0) {
      return;
    }

    // Убираем логирование в продакшене для производительности
    if (process.env.NODE_ENV === 'development') {
      console.log('Creating graph with companies:', filteredCompanies.map(r => r.company.ogrn));
    }

    // Создаем узлы
    const nodeData = [
      // Центральная компания
      {
        id: company.ogrn,
        label: getShortCompanyNameMemo(company.fullName || company.shortName || ""),
        title: createMainCompanyTooltipMemo(company),
        color: {
          background: "#1f2937",
          border: "#374151",
          highlight: {
            background: "#374151",
            border: "#4b5563"
          }
        },
        font: {
          color: "#ffffff",
          size: 14,
          face: "Inter, sans-serif"
        },
        size: 30,
        shape: "box",
        margin: 10,
        borderWidth: 2,
        chosen: {
          node: (values: Record<string, unknown>) => {
            values.color = "#60a5fa";
            values.borderColor = "#3b82f6";
          }
        }
      },
      // Связанные компании
      ...filteredCompanies.map((relation) => {
        const status = getCompanyStatusMemo(relation.company);
        const tooltip = createRelatedCompanyTooltipMemo(relation);
        console.log('Creating node for company:', relation.company.ogrn, 'with tooltip:', tooltip);
        return {
          id: relation.company.ogrn,
          label: getShortCompanyNameMemo(relation.company.fullName || relation.company.shortName || ""),
          title: tooltip,
          color: {
            background: "#ffffff",
            border: status.color,
            highlight: {
              background: "#f3f4f6",
              border: status.color
            }
          },
          font: {
            color: "#1f2937",
            size: 12,
            face: "Inter, sans-serif"
          },
          size: 20,
          shape: "box",
          margin: 8,
          borderWidth: 2,
          chosen: {
            node: (values: Record<string, unknown>) => {
              values.color = "#f3f4f6";
              values.borderColor = "#3b82f6";
            }
          }
        };
      })
    ];

    if (process.env.NODE_ENV === 'development') {
      console.log('Node IDs:', nodeData.map(n => n.id));
      console.log('Node tooltips:', nodeData.map(n => ({ id: n.id, title: n.title })));
    }

    try {
      const nodes = new VisDataSet(nodeData);

      // Создаем связи
      const edgeData = filteredCompanies.map((relation) => {
        const relationshipLabel = getRelationshipLabelMemo(relation.relationshipType, relation);
        
        // Создаем детальный тултип для ребра
        const edgeTooltip = createEdgeTooltip(relation, relationshipLabel);
        
        return {
          id: `${company.ogrn}-${relation.company.ogrn}`,
          from: company.ogrn,
          to: relation.company.ogrn,
          label: relationshipLabel,
          title: edgeTooltip, // Добавляем тултип для ребра
          color: {
            color: RELATIONSHIP_COLORS[relation.relationshipType as keyof typeof RELATIONSHIP_COLORS] || "#6b7280",
            highlight: "#3b82f6",
            hover: "#60a5fa"
          },
          font: {
            color: "#4b5563",
            size: 10,
            face: "Inter, sans-serif",
            strokeWidth: 2,
            strokeColor: "#ffffff"
          },
          width: 2,
          arrows: {
            to: {
              enabled: relation.relationshipType === "SUBSIDIARY_COMPANY",
              scaleFactor: 0.8
            },
            from: {
              enabled: relation.relationshipType === "FOUNDER_COMPANY",
              scaleFactor: 0.8
            }
          },
          smooth: {
            enabled: true,
            type: "curvedCW",
            roundness: 0.1
          },
          chosen: {
            edge: (values: Record<string, unknown>) => {
              values.color = "#3b82f6";
              values.width = 3;
            }
          }
        };
      });

      if (process.env.NODE_ENV === 'development') {
        console.log('Edge IDs:', edgeData.map((e: { id: string }) => e.id));
      }

      const edges = new VisDataSet(edgeData);

    // Настройки сети с оптимизацией производительности
    const options = {
      nodes: {
        borderWidth: 2,
        shadow: {
          enabled: false, // Отключаем тени для производительности
          color: "rgba(0,0,0,0.1)",
          size: 5,
          x: 2,
          y: 2
        }
      },
      edges: {
        shadow: {
          enabled: false, // Отключаем тени для производительности
          color: "rgba(0,0,0,0.1)",
          size: 3,
          x: 1,
          y: 1
        },
        smooth: {
          enabled: true,
          type: "curvedCW",
          roundness: 0.1
        }
      },
      physics: {
        enabled: true,
        stabilization: {
          enabled: true,
          iterations: 50, // Уменьшаем количество итераций
          updateInterval: 50 // Увеличиваем интервал обновления
        },
        barnesHut: {
          gravitationalConstant: -8000,
          centralGravity: 0.3,
          springLength: 200,
          springConstant: 0.04,
          damping: 0.09,
          avoidOverlap: 0.1
        }
      },
      interaction: {
        hover: true,
        hoverConnectedEdges: true,
        selectConnectedEdges: false,
        tooltipDelay: 300, // Небольшая задержка перед показом подсказки
        zoomView: true,
        dragView: true
      },
      layout: {
        improvedLayout: true,
        clusterThreshold: 150
      }
    };

      // Создаем сеть
      networkInstance.current = new VisNetwork(networkRef.current, { nodes, edges }, options);

      // Обработчики событий
      networkInstance.current.on("click", (params: Record<string, unknown>) => {
        const nodesList = params.nodes as string[];
        if (nodesList && nodesList.length > 0) {
          const nodeId = nodesList[0];
          if (nodeId !== company.ogrn) {
            // Открываем страницу связанной компании
            window.open(`/company/${nodeId}`, '_blank');
          }
        }
      });

      // Альтернативный способ отображения тултипов через DOM
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      networkInstance.current.on("hoverNode", (params: any) => {
        if (networkRef.current) {
          networkRef.current.style.cursor = "pointer";
          
          // Находим данные узла для отображения тултипа
          const nodeId = params.node;
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const nodeData = nodes.get(nodeId) as any;
          
          if (nodeData && nodeData.title) {
            // Устанавливаем title атрибут для HTML тултипа как fallback
            networkRef.current.title = nodeData.title;
            console.log('Setting HTML title tooltip for node:', nodeId, nodeData.title);
          }
        }
      });

      networkInstance.current.on("blurNode", () => {
        if (networkRef.current) {
          networkRef.current.style.cursor = "default";
          networkRef.current.title = ""; // Очищаем HTML тултип
        }
      });

      // Обработчики событий для ребер (связей)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      networkInstance.current.on("hoverEdge", (params: any) => {
        if (networkRef.current) {
          networkRef.current.style.cursor = "pointer";
          
          // Находим данные ребра для отображения тултипа
          const edgeId = params.edge;
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          const edgeData = edges.get(edgeId) as any;
          
          if (edgeData && edgeData.title) {
            // Устанавливаем title атрибут для HTML тултипа как fallback
            networkRef.current.title = edgeData.title;
            console.log('Setting HTML title tooltip for edge:', edgeId, edgeData.title);
          }
        }
      });

      networkInstance.current.on("blurEdge", () => {
        if (networkRef.current) {
          networkRef.current.style.cursor = "default";
          networkRef.current.title = ""; // Очищаем HTML тултип
        }
      });

      // Центрируем граф после стабилизации
      networkInstance.current.once("stabilizationIterationsDone", () => {
        networkInstance.current?.fit({
          animation: {
            duration: 500, // Уменьшаем время анимации
            easingFunction: "easeInOutQuad"
          }
        });
      });

      if (process.env.NODE_ENV === 'development') {
        console.log('Graph created successfully');
      }
    } catch (error) {
      console.error('Error creating graph:', error);
      // Если произошла ошибка, очищаем ссылку на сеть
      networkInstance.current = null;
    }

    return () => {
      // Безопасная очистка при размонтировании компонента
      if (networkInstance.current) {
        try {
          networkInstance.current.destroy();
        } catch (error) {
          // Игнорируем ошибки при уничтожении во время размонтирования
          if (process.env.NODE_ENV === 'development') {
            console.warn('Warning during component cleanup:', error);
          }
        }
        networkInstance.current = null;
      }
    };
  }, [
    company,
    filteredCompanies, 
    isLoading, 
    selectedTypes.length,
    visLibsLoaded,
    getCompanyStatusMemo,
    getRelationshipLabelMemo,
    getShortCompanyNameMemo,
    createMainCompanyTooltipMemo,
    createRelatedCompanyTooltipMemo
  ]);

  const handleZoomIn = () => {
    if (networkInstance.current) {
      try {
        const scale = networkInstance.current.getScale();
        networkInstance.current.moveTo({
          scale: scale * 1.2,
          animation: { duration: 300, easingFunction: "easeInOutQuad" }
        });
      } catch (error) {
        if (process.env.NODE_ENV === 'development') {
          console.warn('Error during zoom in:', error);
        }
      }
    }
  };

  const handleZoomOut = () => {
    if (networkInstance.current) {
      try {
        const scale = networkInstance.current.getScale();
        networkInstance.current.moveTo({
          scale: scale * 0.8,
          animation: { duration: 300, easingFunction: "easeInOutQuad" }
        });
      } catch (error) {
        if (process.env.NODE_ENV === 'development') {
          console.warn('Error during zoom out:', error);
        }
      }
    }
  };

  const handleFit = () => {
    if (networkInstance.current) {
      try {
        networkInstance.current.fit({
          animation: { duration: 500, easingFunction: "easeInOutQuad" }
        });
      } catch (error) {
        if (process.env.NODE_ENV === 'development') {
          console.warn('Error during fit:', error);
        }
      }
    }
  };

  const handleReset = () => {
    if (networkInstance.current) {
      try {
        // Просто сбрасываем позиции без пересоздания графа
        networkInstance.current.fit({
          animation: { duration: 500, easingFunction: "easeInOutQuad" }
        });
      } catch (error) {
        if (process.env.NODE_ENV === 'development') {
          console.warn('Error during reset:', error);
        }
      }
    }
  };

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <NetworkIcon className="h-5 w-5" />
            Граф связей
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-96 bg-gray-100 rounded-lg flex items-center justify-center">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p className="text-gray-500">Загрузка графа связей...</p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!visLibsLoaded || !VisNetwork || !VisDataSet) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <NetworkIcon className="h-5 w-5" />
            Граф связей
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-96 bg-gray-100 rounded-lg flex items-center justify-center">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p className="text-gray-500 mb-2">Загрузка библиотек визуализации...</p>
              <p className="text-sm text-gray-400">
                Подготовка графа связей
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (relatedCompanies.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <NetworkIcon className="h-5 w-5" />
            Граф связей
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-96 bg-gray-100 rounded-lg flex items-center justify-center">
            <div className="text-center">
              <NetworkIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-500 mb-2">Нет данных для отображения графа</p>
              <p className="text-sm text-gray-400">
                Связанные компании не найдены
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (filteredCompanies.length === 0 && selectedTypes.length > 0) {
    return (
      <div className={isFullscreen ? "fixed inset-0 z-50 bg-white" : ""}>
        <Card className={isFullscreen ? "h-full border-0 rounded-none" : ""}>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <NetworkIcon className="h-5 w-5" />
                Граф связей
                <Badge variant="outline">
                  0 из {relatedCompanies.length}
                </Badge>
              </div>
              <div className="flex items-center gap-2">
                <GraphSettings
                  availableTypes={availableTypes}
                  selectedTypes={selectedTypes}
                  onTypesChange={setSelectedTypes}
                />
                <Button
                  variant="outline"
                  size="sm"
                  onClick={toggleFullscreen}
                  title={isFullscreen ? "Выйти из полноэкранного режима" : "Полноэкранный режим"}
                >
                  {isFullscreen ? "Выйти" : "Развернуть"}
                </Button>
              </div>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-96 bg-gray-50 rounded-lg flex items-center justify-center border-2 border-dashed border-gray-200">
              <div className="text-center max-w-md">
                <NetworkIcon className="h-16 w-16 text-gray-300 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-700 mb-2">
                  Все связи отфильтрованы
                </h3>
                <p className="text-gray-500 mb-4">
                  Выбранные фильтры не содержат связанных компаний. 
                  Попробуйте изменить настройки фильтров или выбрать &ldquo;Все&rdquo;.
                </p>
                <Button
                  variant="outline"
                  onClick={() => setSelectedTypes(availableTypes)}
                  className="text-sm"
                >
                  Показать все типы связей
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={isFullscreen ? "fixed inset-0 z-50 bg-white" : ""}>
      <Card className={isFullscreen ? "h-full border-0 rounded-none" : ""}>
        <CardHeader>
          <CardTitle className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <NetworkIcon className="h-5 w-5" />
              Граф связей
              <Badge variant="secondary">
                {filteredCompanies.length} из {relatedCompanies.length}
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              <GraphSettings
                availableTypes={availableTypes}
                selectedTypes={selectedTypes}
                onTypesChange={setSelectedTypes}
              />
              <Button
                variant="outline"
                size="sm"
                onClick={handleZoomIn}
                title="Увеличить"
              >
                <ZoomIn className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleZoomOut}
                title="Уменьшить"
              >
                <ZoomOut className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleFit}
                title="Показать все"
              >
                <Maximize2 className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleReset}
                title="Сбросить позиции"
              >
                <RotateCcw className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={toggleFullscreen}
                title={isFullscreen ? "Выйти из полноэкранного режима" : "Полноэкранный режим"}
              >
                {isFullscreen ? "Выйти" : "Развернуть"}
              </Button>
            </div>
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <div 
            ref={networkRef} 
            className={`bg-gray-50 ${isFullscreen ? "h-[calc(100vh-80px)]" : "h-96"}`}
            style={{ width: "100%" }}
          />
          
          {/* Легенда */}
          <div className="p-4 border-t bg-gray-50">
            <div className="flex flex-wrap gap-4 text-sm">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded" style={{ backgroundColor: RELATIONSHIP_COLORS.FOUNDER_COMPANY }}></div>
                <span>Учредители</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded" style={{ backgroundColor: RELATIONSHIP_COLORS.SUBSIDIARY_COMPANY }}></div>
                <span>Дочерние</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded" style={{ backgroundColor: RELATIONSHIP_COLORS.COMMON_FOUNDERS }}></div>
                <span>Общие учредители</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded" style={{ backgroundColor: RELATIONSHIP_COLORS.COMMON_DIRECTORS }}></div>
                <span>Общие руководители</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded" style={{ backgroundColor: RELATIONSHIP_COLORS.COMMON_ADDRESS }}></div>
                <span>Общий адрес</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded" style={{ backgroundColor: RELATIONSHIP_COLORS.FOUNDER_TO_DIRECTOR }}></div>
                <span>Перекрестные связи</span>
              </div>
            </div>
            <div className="mt-2 text-xs text-gray-500">
              Кликните на узел для перехода к карточке компании. Используйте мышь для навигации по графу.
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}