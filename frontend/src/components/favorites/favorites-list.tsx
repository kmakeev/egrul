"use client";

import { Building2, User, Trash2 } from "lucide-react";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { useToast } from "@/hooks/use-toast";
import {
  type Favorite,
  useDeleteFavoriteMutation,
} from "@/lib/api/favorites-hooks";
import { EntityType } from "@/lib/api/subscription-hooks";

interface FavoritesListProps {
  favorites: Favorite[];
}

export function FavoritesList({ favorites }: FavoritesListProps) {
  const { toast } = useToast();
  const queryClient = useQueryClient();

  const deleteFavorite = useDeleteFavoriteMutation({
    onSuccess: () => {
      // Явно инвалидируем кэш для немедленного обновления UI
      queryClient.invalidateQueries({
        queryKey: ["favorites", "my"],
      });
      queryClient.invalidateQueries({
        queryKey: ["favorite"],
      });
      toast({
        title: "Удалено из избранного",
        description: "Запись успешно удалена",
      });
    },
    onError: (error) => {
      toast({
        title: "Ошибка при удалении",
        description: error.message,
        variant: "destructive",
      });
    },
  });

  const handleDelete = (id: string) => {
    deleteFavorite.mutate(id);
  };

  if (favorites.length === 0) {
    return (
      <div className="text-center py-8">
        <p className="text-muted-foreground">
          Избранных записей не найдено
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {favorites.map((favorite) => {
        const isCompany = favorite.entityType === EntityType.COMPANY;
        const entityUrl = isCompany
          ? `/company/${favorite.entityId}`
          : `/entrepreneur/${favorite.entityId}`;

        return (
          <Card key={favorite.id}>
            <CardHeader className="pb-2">
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-2 flex-1">
                  <div className="p-2 rounded-lg bg-primary/10">
                    {isCompany ? (
                      <Building2 className="h-4 w-4 text-primary" />
                    ) : (
                      <User className="h-4 w-4 text-primary" />
                    )}
                  </div>
                  <div className="flex-1">
                    <Link
                      href={entityUrl}
                      className="hover:underline cursor-pointer"
                    >
                      <CardTitle className="text-base mb-1">
                        {favorite.entityName}
                      </CardTitle>
                    </Link>
                    <p className="text-sm text-muted-foreground">
                      {isCompany ? "ОГРН" : "ОГРНИП"}: {favorite.entityId}
                    </p>
                  </div>
                </div>
              </div>
            </CardHeader>

            <CardContent className="space-y-2">
              {/* Заметки */}
              {favorite.notes && (
                <div>
                  <p className="text-xs font-medium mb-1">Заметки:</p>
                  <p className="text-xs text-muted-foreground">
                    {favorite.notes}
                  </p>
                </div>
              )}

              {/* Метаданные */}
              <div className="flex items-center justify-between text-xs text-muted-foreground pt-1">
                <span>
                  Добавлено: {new Date(favorite.createdAt).toLocaleDateString("ru-RU")}
                </span>
              </div>

              {/* Действия */}
              <div className="flex gap-2 pt-2">
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={deleteFavorite.isPending}
                    >
                      <Trash2 className="h-4 w-4 mr-1" />
                      Удалить
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Удалить из избранного?</AlertDialogTitle>
                      <AlertDialogDescription>
                        Запись {favorite.entityName} будет удалена из избранного.
                        Это действие нельзя отменить.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Отмена</AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() => handleDelete(favorite.id)}
                      >
                        Удалить
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
