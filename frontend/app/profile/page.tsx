"use client";
import { useTheme } from "next-themes";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button } from "@heroui/button";
import { Card } from "@heroui/card";
import { Divider } from "@heroui/divider";

import { authApi } from "@/api/api";
import { Loading } from "@/components/loading";
import {
  ChatBubbleIcon,
  ChevronRightIcon,
  EditIcon,
  GearIcon,
  ListIcon,
  TrophyIcon,
} from "@/config/icons";
import Avatar from "@/components/avatar";

export default function ProfilePage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);

  const [user, setUser] = useState<any>({});

  const router = useRouter();
  const { theme, setTheme } = useTheme();

  async function fetchData() {
    setIsLoading(true);
    authApi.v1
      .userServiceGetMe()
      .then((response) => {
        console.log(response.data);
        setIsError(false);
        setUser(response.data.user!);
      })
      .catch((error) => {
        console.log(error);
        setIsError(true);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }

  useEffect(() => {
    fetchData();
  }, []);

  if (isLoading) {
    return <Loading />;
  }

  if (isError) {
    return (
      <div className="p-4">
        <h2 className="text-lg text-red-500">Ошибка при загрузке данных</h2>
        <p>Проверьте соединение с сервером или обновите страницу.</p>
      </div>
    );
  }

  function SubPageButton({
    href,
    label,
    icon,
  }: {
    href: string;
    label: string;
    icon?: React.ReactNode;
  }) {
    return (
      <Card isPressable className="p-3" onPress={() => router.push(href)}>
        <div className="flex flex-row justify-between items-center w-full">
          <div className="flex flex-row items-center gap-2">
            {icon}
            <p>{label}</p>
          </div>
          <ChevronRightIcon className="w-4 h-4" fill="currentColor" />
        </div>
      </Card>
    );
  }

  return (
    <div className="p-4 flex-grow gap-4">
      <div className="grid grid-cols-1 gap-4 py-4">
        <div className="flex flex-col gap-4 items-center justify-around">
          <Avatar src={user.profilePictureUrl} />
          <h2 className="text-2xl font-bold">
            {user.firstName} {user.lastName}
          </h2>
        </div>
        <Divider />
        {/* Кнопочки история трениировок и еще какие-то которые я не придумал */}
        <SubPageButton
          href="/profile/edit"
          icon={<EditIcon className="w-4 h-4" fill="currentColor" />}
          label="Редактировать профиль"
        />
        <SubPageButton
          href="/profile/workouts"
          icon={<ListIcon className="w-4 h-4" fill="currentColor" />}
          label="История тренировок"
        />
        <SubPageButton
          href="/profile/analytics"
          icon={<TrophyIcon className="w-4 h-4" fill="currentColor" />}
          label="Аналитика"
        />
        <SubPageButton
          href="/profile/generation-settings"
          icon={<GearIcon className="w-4 h-4" fill="currentColor" />}
          label="Настройки генерации"
        />
        <SubPageButton
          href="/profile/user-facts"
          icon={<ChatBubbleIcon className="w-4 h-4" fill="currentColor" />}
          label="Факты о вас"
        />
        {/*  Light and dark mode switch */}
        <Divider />
        <Button
          color="warning"
          onPress={() => {
            setTheme(theme === "dark" ? "light" : "dark");
          }}
        >
          Сменить тему
        </Button>
      </div>
    </div>
  );
}
