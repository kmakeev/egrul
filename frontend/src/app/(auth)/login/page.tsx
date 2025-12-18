"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { loginSchema, type LoginInput } from "@/lib/validations/auth";
import { useAuthStore } from "@/store/auth-store";
import { useToast } from "@/components/ui/use-toast";
import { useRouter } from "next/navigation";

export default function LoginPage() {
  const { toast } = useToast();
  const router = useRouter();
  const { setUser } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const onSubmit = async (data: LoginInput) => {
    setIsLoading(true);
    try {
      // TODO: Реализовать реальный API вызов
      // const response = await authApi.login(data);
      
      // Временная заглушка
      setUser({
        id: "1",
        email: data.email,
        firstName: "Пользователь",
        lastName: "Тестовый",
        createdAt: new Date().toISOString(),
      });

      toast({
        title: "Успешный вход",
        description: "Вы успешно вошли в систему",
      });

      router.push("/search");
    } catch (error) {
      toast({
        title: "Ошибка входа",
        description: error instanceof Error ? error.message : "Неизвестная ошибка",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Вход в систему</CardTitle>
          <CardDescription>
            Введите ваши учетные данные для входа
          </CardDescription>
        </CardHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <CardContent className="space-y-4">
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input
                        type="email"
                        placeholder="example@mail.com"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Пароль</FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        placeholder="••••••••"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </CardContent>
            <CardFooter className="flex flex-col space-y-4">
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? "Вход..." : "Войти"}
              </Button>
              <p className="text-center text-sm text-muted-foreground">
                Нет аккаунта?{" "}
              <Link href="/register" className="text-primary hover:underline">
                  Зарегистрироваться
                </Link>
              </p>
            </CardFooter>
          </form>
        </Form>
      </Card>
  );
}

