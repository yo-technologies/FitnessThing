"use client";

import { useEffect, useState } from "react";
import { Button } from "@nextui-org/button";
import { Card, CardBody } from "@nextui-org/card";
import { Chip } from "@nextui-org/chip";
import { toast } from "react-toastify";

import { authApi } from "@/api/api";
import { WorkoutUserFact } from "@/api/api.generated";
import { Loading } from "@/components/loading";
import { PageHeader } from "@/components/page-header";
import { ChatBubbleIcon, CircleQuestionIcon } from "@/config/icons";

const FACTS_LIMIT = 32;

export default function UserFactsPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);
  const [userFacts, setUserFacts] = useState<WorkoutUserFact[]>([]);
  const [deletingFactId, setDeletingFactId] = useState<string | null>(null);

  function sortFacts(facts: WorkoutUserFact[]) {
    return [...facts].sort((a, b) => {
      const aTime = a.createdAt ? new Date(a.createdAt).getTime() : 0;
      const bTime = b.createdAt ? new Date(b.createdAt).getTime() : 0;

      return bTime - aTime;
    });
  }

  async function fetchUserFacts() {
    try {
      const response = await authApi.v1.userServiceListUserFacts();

      setUserFacts(sortFacts(response.data.facts || []));
      setIsError(false);
    } catch (error) {
      console.log(error);
      setIsError(true);
      toast.error("Не удалось загрузить факты о пользователе");
    }
  }

  useEffect(() => {
    async function fetchData() {
      setIsLoading(true);
      await fetchUserFacts();
      setIsLoading(false);
    }

    fetchData();
  }, []);

  function formatFactDate(date?: string) {
    if (!date) {
      return "";
    }

    const dateObj = new Date(date);

    return dateObj.toLocaleDateString("ru-RU", {
      day: "numeric",
      month: "long",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  async function handleDeleteFact(factId?: string) {
    if (!factId) {
      return;
    }

    setDeletingFactId(factId);

    try {
      await authApi.v1.userServiceDeleteUserFact(factId);

      setUserFacts((prevFacts) =>
        prevFacts.filter((fact) => fact.id !== factId),
      );
      toast.success("Факт удалён");
    } catch (error) {
      console.log(error);
      toast.error("Не удалось удалить факт");
    } finally {
      setDeletingFactId(null);
    }
  }

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

  return (
    <div className="py-4 flex flex-col h-full gap-4">
      <PageHeader enableBackButton title="Факты о вас" />
      <div className="flex flex-col gap-4 px-4 pb-4">
        <Card>
          <CardBody className="flex flex-col gap-4">
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-center gap-2">
                <CircleQuestionIcon className="w-4 h-4 text-primary" />
                <p className="text-sm font-semibold text-primary">
                  Как используются факты
                </p>
              </div>
              <Chip color="primary" size="sm" variant="flat">
                {userFacts.length}/{FACTS_LIMIT}
              </Chip>
            </div>
            <p className="text-xs text-default-600">
              Ассистент фиксирует важные детали ваших запросов, чтобы помнить
              контекст. Здесь вы можете удалить устаревшие или неверные факты.
            </p>
          </CardBody>
        </Card>

        {userFacts.length === 0 ? (
          <Card>
            <CardBody className="flex flex-col gap-3 items-center text-center">
              <ChatBubbleIcon className="w-6 h-6 text-default-300" />
              <div className="flex flex-col gap-1">
                <p className="text-xs font-semibold text-default-600">
                  Пока что пусто
                </p>
                <p className="text-xs text-default-500">
                  Как только ассистент найдёт важную информацию в диалогах, она
                  появится здесь.
                </p>
              </div>
            </CardBody>
          </Card>
        ) : (
          <div className="flex flex-col gap-3">
            {userFacts.map((fact, index) => (
              <Card key={fact.id} className="shadow-sm">
                <CardBody className="flex flex-col gap-3">
                  <div className="flex items-center justify-between">
                    <Chip color="primary" size="sm" variant="flat">
                      Факт #{index + 1}
                    </Chip>
                    <p className="text-xs text-default-400">
                      {formatFactDate(fact.createdAt)}
                    </p>
                  </div>
                  <p className="text-xs text-default-700 whitespace-pre-wrap">
                    {fact.content}
                  </p>
                  <div className="flex flex-row items-center justify-end">
                    <Button
                      color="danger"
                      isLoading={deletingFactId === fact.id}
                      size="sm"
                      variant="light"
                      onPress={() => handleDeleteFact(fact.id)}
                    >
                      Удалить
                    </Button>
                  </div>
                </CardBody>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
