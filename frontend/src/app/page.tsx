import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Search, BarChart3, Bookmark } from "lucide-react";
import { Header } from "@/components/layout/Header";

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col">
      <Header />
      <main className="flex-1 flex items-center justify-center p-4">
      <div className="container mx-auto max-w-4xl space-y-8">
        <div className="text-center space-y-4">
          <h1 className="text-4xl font-bold tracking-tight sm:text-6xl">
            ЕГРЮЛ/ЕГРИП
          </h1>
          <p className="text-xl text-muted-foreground">
            Система поиска и анализа данных ЕГРЮЛ и ЕГРИП
          </p>
        </div>

        <div className="grid gap-6 md:grid-cols-3">
            <Card className="flex flex-col">
            <CardHeader>
              <Search className="h-8 w-8 mb-2 text-primary" />
              <CardTitle>Поиск</CardTitle>
              <CardDescription>
                Быстрый поиск по юридическим лицам и индивидуальным предпринимателям
              </CardDescription>
            </CardHeader>
              <CardContent className="mt-auto">
              <Link href="/search">
                <Button className="w-full">Начать поиск</Button>
              </Link>
            </CardContent>
          </Card>

            <Card className="flex flex-col">
            <CardHeader>
              <BarChart3 className="h-8 w-8 mb-2 text-primary" />
              <CardTitle>Аналитика</CardTitle>
              <CardDescription>
                Статистика и аналитические данные по реестрам
              </CardDescription>
            </CardHeader>
              <CardContent className="mt-auto">
              <Link href="/analytics">
                  <Button className="w-full">Открыть аналитику</Button>
              </Link>
            </CardContent>
          </Card>

            <Card className="flex flex-col">
            <CardHeader>
              <Bookmark className="h-8 w-8 mb-2 text-primary" />
              <CardTitle>Избранное</CardTitle>
              <CardDescription>
                Сохраняйте интересующие компании для быстрого доступа
              </CardDescription>
            </CardHeader>
              <CardContent className="mt-auto">
              <Link href="/watchlist">
                  <Button className="w-full">Перейти к избранному</Button>
              </Link>
            </CardContent>
          </Card>
        </div>
    </div>
      </main>
    </div>
  );
}
